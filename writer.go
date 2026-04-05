package muon

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
	"strings"

	"ekyu.moe/leb128"
	"oherych/muon/internal"
)

const (
	longStringFactor = 512
	lruMaxSize       = 512
)

var (
	// maps element kind → TypedArray type byte; int/uint omitted (platform-dependent)
	elemKindToTypeByte = map[reflect.Kind]byte{
		reflect.Int8:    typeInt8,
		reflect.Int16:   typeInt16,
		reflect.Int32:   typeInt32,
		reflect.Int64:   typeInt64,
		reflect.Uint8:   typeUint8,
		reflect.Uint16:  typeUint16,
		reflect.Uint32:  typeUint32,
		reflect.Uint64:  typeUint64,
		reflect.Float32: typeFloat32,
		reflect.Float64: typeFloat64,
	}
)

type Encoder struct {
	// LRU enables string reference deduplication. When true, repeated strings
	// are written as back-references (0x81 + index) instead of full strings.
	LRU bool
	// Deterministic enforces canonical encoding: sorted dict keys, LRU disabled.
	Deterministic bool
	lru           []string
}

func (e *Encoder) Write(w io.Writer, in interface{}) error {
	return e.write(w, in)
}

var magic = []byte{tagMagicByte, 0xB5, 0x30, 0x31}

func (e *Encoder) WriteWithMagic(w io.Writer, in interface{}) error {
	if err := e.writeBytes(w, magic); err != nil {
		return err
	}
	return e.write(w, in)
}

func (e *Encoder) WritePadding(w io.Writer, n int) error {
	pad := make([]byte, n)
	for i := range pad {
		pad[i] = tagPadding
	}
	_, err := w.Write(pad)
	return err
}

// WriteChunkedTypedArray writes a chunked TypedArray (0x85).
// Each argument must be a slice of the type corresponding to typeByte.
func (e *Encoder) WriteChunkedTypedArray(w io.Writer, typeByte byte, chunks ...interface{}) error {
	if err := e.writeBytes(w, []byte{typedArrayChunk, typeByte}); err != nil {
		return err
	}
	for _, chunk := range chunks {
		rv := reflect.ValueOf(chunk)
		if rv.Kind() != reflect.Slice {
			return fmt.Errorf("WriteChunkedTypedArray: chunk must be a slice, got %T", chunk)
		}
		n := rv.Len()
		if err := e.writeBytes(w, leb128.AppendUleb128(nil, uint64(n))); err != nil {
			return err
		}
		for i := 0; i < n; i++ {
			if err := e.writeTypedElem(w, rv.Index(i), typeByte); err != nil {
				return err
			}
		}
	}
	// terminating zero-length chunk
	return e.writeBytes(w, []byte{0x00})
}

func (e *Encoder) write(w io.Writer, in interface{}) error {
	if m, ok := in.(Marshaler); ok {
		data, err := m.MarshalMuon()
		if err != nil {
			return err
		}
		return e.writeBytes(w, data)
	}

	if m, ok := in.(MarshalerStream); ok {
		return m.MarshalMuon(w)
	}

	if in == nil {
		return e.writeByte(w, nilValue)
	}

	rv := reflect.ValueOf(in)
	kind := rv.Kind()

	if kind == reflect.Bool {
		return e.writeBool(w, rv.Bool())
	}
	if kind == reflect.String {
		return e.writeString(w, rv.String())
	}
	if kind >= reflect.Int && kind <= reflect.Int64 {
		return e.writeInteger(w, rv)
	}
	if kind >= reflect.Uint && kind <= reflect.Uint64 {
		return e.writeUint(w, rv)
	}
	if kind == reflect.Float32 || kind == reflect.Float64 {
		return e.writeFloat(w, rv)
	}
	if kind == reflect.Slice || kind == reflect.Array {
		return e.writeList(w, rv)
	}
	if kind == reflect.Map {
		return e.writeMap(w, rv)
	}
	if kind == reflect.Struct {
		return e.writeStruct(w, rv)
	}
	if kind == reflect.Pointer {
		if rv.IsNil() {
			return e.writeByte(w, nilValue)
		}
		return e.write(w, rv.Elem().Interface())
	}

	return fmt.Errorf("type %s not supportable", rv.Type())
}

func (e *Encoder) writeBool(w io.Writer, v bool) error {
	if v {
		return e.writeByte(w, boolTrue)
	}
	return e.writeByte(w, boolFalse)
}

func (e *Encoder) writeInteger(w io.Writer, rv reflect.Value) error {
	v := rv.Int()
	if v >= 0 && v <= 9 {
		return e.writeBytes(w, []byte{0xA0 + byte(v)})
	}
	return e.writeBytes(w, []byte{0xBB}, leb128.AppendSleb128(nil, v))
}

func (e *Encoder) writeUint(w io.Writer, rv reflect.Value) error {
	v := rv.Uint()
	if v <= 9 {
		return e.writeBytes(w, []byte{0xA0 + byte(v)})
	}
	return e.writeBytes(w, []byte{0xBB}, leb128.AppendSleb128(nil, int64(v)))
}

func (e *Encoder) writeFloat(w io.Writer, rv reflect.Value) error {
	v := rv.Float()
	if math.IsNaN(v) {
		return e.writeByte(w, nanValue)
	}
	if math.IsInf(v, -1) {
		return e.writeByte(w, negativeInfValue)
	}
	if math.IsInf(v, 1) {
		return e.writeByte(w, positiveInfValue)
	}
	var buf [9]byte
	buf[0] = floatF64
	binary.LittleEndian.PutUint64(buf[1:], math.Float64bits(v))
	return e.writeBytes(w, buf[:])
}

func (e *Encoder) writeString(w io.Writer, v string) error {
	if e.LRU && !e.Deterministic {
		for i, s := range e.lru {
			if s == v {
				return e.writeBytes(w, []byte{stringRef}, leb128.AppendUleb128(nil, uint64(i)))
			}
		}
		// not in LRU — write with 0x8C tag and remember
		e.lruPrepend(v)
		if err := e.writeByte(w, tagRefString); err != nil {
			return err
		}
	}

	// fixed-length string: length >= 512 bytes or contains 0x00
	if len(v) >= longStringFactor || strings.ContainsRune(v, stringEnd) {
		return e.writeBytes(w, []byte{tagSize}, leb128.AppendUleb128(nil, uint64(len(v))), []byte(v))
	}
	return e.writeBytes(w, []byte(v), []byte{stringEnd})
}

func (e *Encoder) lruPrepend(s string) {
	if len(e.lru) >= lruMaxSize {
		e.lru = e.lru[:lruMaxSize-1]
	}
	e.lru = append([]string{s}, e.lru...)
}

func (e *Encoder) writeList(w io.Writer, rv reflect.Value) error {
	elemKind := rv.Type().Elem().Kind()
	if tb, ok := elemKindToTypeByte[elemKind]; ok {
		return e.writeTypedArray(w, rv, tb)
	}
	if err := e.writeByte(w, listStart); err != nil {
		return err
	}
	for i := 0; i < rv.Len(); i++ {
		if err := e.write(w, rv.Index(i).Interface()); err != nil {
			return err
		}
	}
	return e.writeByte(w, listEnd)
}

func (e *Encoder) writeTypedArray(w io.Writer, rv reflect.Value, typeByte byte) error {
	n := rv.Len()
	if err := e.writeBytes(w, []byte{typedArray, typeByte}, leb128.AppendUleb128(nil, uint64(n))); err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		if err := e.writeTypedElem(w, rv.Index(i), typeByte); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) writeTypedElem(w io.Writer, rv reflect.Value, typeByte byte) error {
	var buf [8]byte
	switch typeByte {
	case typeInt8:
		buf[0] = byte(rv.Int())
		return e.writeBytes(w, buf[:1])
	case typeInt16:
		binary.LittleEndian.PutUint16(buf[:], uint16(rv.Int()))
		return e.writeBytes(w, buf[:2])
	case typeInt32:
		binary.LittleEndian.PutUint32(buf[:], uint32(rv.Int()))
		return e.writeBytes(w, buf[:4])
	case typeInt64:
		binary.LittleEndian.PutUint64(buf[:], uint64(rv.Int()))
		return e.writeBytes(w, buf[:8])
	case typeUint8:
		buf[0] = byte(rv.Uint())
		return e.writeBytes(w, buf[:1])
	case typeUint16:
		binary.LittleEndian.PutUint16(buf[:], uint16(rv.Uint()))
		return e.writeBytes(w, buf[:2])
	case typeUint32:
		binary.LittleEndian.PutUint32(buf[:], uint32(rv.Uint()))
		return e.writeBytes(w, buf[:4])
	case typeUint64:
		binary.LittleEndian.PutUint64(buf[:], rv.Uint())
		return e.writeBytes(w, buf[:8])
	case typeFloat32:
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(float32(rv.Float())))
		return e.writeBytes(w, buf[:4])
	case typeFloat64:
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(rv.Float()))
		return e.writeBytes(w, buf[:8])
	}
	return fmt.Errorf("unsupported typed array element type byte: 0x%02X", typeByte)
}

func (e *Encoder) writeMap(w io.Writer, rv reflect.Value) error {
	keys := rv.MapKeys()
	if len(keys) == 0 {
		return e.writeBytes(w, []byte{dictStart, dictEnd})
	}

	firstKind := keys[0].Kind()
	isString := firstKind == reflect.String
	isInt := firstKind >= reflect.Int && firstKind <= reflect.Int64 ||
		firstKind >= reflect.Uint && firstKind <= reflect.Uint64
	if !isString && !isInt {
		return fmt.Errorf("dict keys must be string or integer, got %s", firstKind)
	}
	for _, k := range keys[1:] {
		kk := k.Kind()
		if isString && kk != reflect.String {
			return fmt.Errorf("mixed dict key types: expected string, got %s", kk)
		}
		if isInt {
			isKInt := kk >= reflect.Int && kk <= reflect.Int64 ||
				kk >= reflect.Uint && kk <= reflect.Uint64
			if !isKInt {
				return fmt.Errorf("mixed dict key types: expected integer, got %s", kk)
			}
		}
	}

	if e.Deterministic {
		if isString {
			sort.Slice(keys, func(i, j int) bool {
				return keys[i].String() < keys[j].String()
			})
		} else {
			sort.Slice(keys, func(i, j int) bool {
				ki, kj := keys[i], keys[j]
				if ki.Kind() >= reflect.Uint && ki.Kind() <= reflect.Uint64 {
					return ki.Uint() < kj.Uint()
				}
				return ki.Int() < kj.Int()
			})
		}
	}

	if err := e.writeByte(w, dictStart); err != nil {
		return err
	}
	for i, k := range keys {
		if isInt {
			if err := e.writeDictIntKey(w, k, i == 0); err != nil {
				return err
			}
		} else {
			if err := e.writeString(w, k.String()); err != nil {
				return err
			}
		}
		if err := e.write(w, rv.MapIndex(k).Interface()); err != nil {
			return err
		}
	}
	return e.writeByte(w, dictEnd)
}

func (e *Encoder) writeDictIntKey(w io.Writer, rv reflect.Value, first bool) error {
	kind := rv.Kind()
	isUint := kind >= reflect.Uint && kind <= reflect.Uint64

	var typeByte byte
	var leSize int
	switch kind {
	case reflect.Int8:
		typeByte, leSize = typeInt8, 1
	case reflect.Uint8:
		typeByte, leSize = typeUint8, 1
	case reflect.Int16:
		typeByte, leSize = typeInt16, 2
	case reflect.Uint16:
		typeByte, leSize = typeUint16, 2
	case reflect.Int32:
		typeByte, leSize = typeInt32, 4
	case reflect.Uint32:
		typeByte, leSize = typeUint32, 4
	case reflect.Int64:
		typeByte, leSize = typeInt64, 8
	case reflect.Uint64:
		typeByte, leSize = typeUint64, 8
	}

	if leSize > 0 {
		if first {
			if err := e.writeByte(w, typeByte); err != nil {
				return err
			}
		}
		var buf [8]byte
		if isUint {
			binary.LittleEndian.PutUint64(buf[:], rv.Uint())
		} else {
			binary.LittleEndian.PutUint64(buf[:], uint64(rv.Int()))
		}
		return e.writeBytes(w, buf[:leSize])
	}

	// int/uint (platform-dependent): SLEB128, omit 0xBB prefix after first key
	if first {
		if err := e.writeByte(w, 0xBB); err != nil {
			return err
		}
	}
	if isUint {
		return e.writeBytes(w, leb128.AppendSleb128(nil, int64(rv.Uint())))
	}
	return e.writeBytes(w, leb128.AppendSleb128(nil, rv.Int()))
}

func (e *Encoder) writeStruct(w io.Writer, rv reflect.Value) error {
	if err := e.writeBytes(w, []byte{dictStart}); err != nil {
		return err
	}
	tt := rv.Type()
	for i := 0; i < tt.NumField(); i++ {
		tf := tt.Field(i)
		vf := rv.Field(i)
		if !tf.IsExported() {
			continue
		}
		info := internal.ParseTags(tf)
		if info.Skip {
			continue
		}
		if err := e.writeString(w, info.Name); err != nil {
			return err
		}
		if err := e.write(w, vf.Interface()); err != nil {
			return err
		}
	}
	return e.writeBytes(w, []byte{dictEnd})
}

func (e *Encoder) writeBytes(w io.Writer, val ...[]byte) error {
	for _, v := range val {
		if _, err := w.Write(v); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) writeByte(w io.Writer, val byte) error {
	_, err := w.Write([]byte{val})
	return err
}
