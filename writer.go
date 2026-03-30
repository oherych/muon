package muon

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"

	"ekyu.moe/leb128"
	"oherych/muon/internal"
)

const (
	longStringFactor = 512
)

var (
	kindToType = map[reflect.Kind]byte{
		reflect.Int:     0,
		reflect.Int8:    0,
		reflect.Int16:   0,
		reflect.Int32:   0,
		reflect.Int64:   0,
		reflect.Uint:    0,
		reflect.Uint8:   0,
		reflect.Uint16:  0,
		reflect.Uint32:  0,
		reflect.Uint64:  0,
		reflect.Uintptr: 0,
		reflect.Float32: 0,
		reflect.Float64: 0,
	}
)

type Encoder struct{}

func (e Encoder) Write(w io.Writer, in interface{}) error {
	return e.write(w, in)
}

func (e Encoder) write(w io.Writer, in interface{}) error {
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
		return e.write(w, rv.Elem().Interface())
	}

	return fmt.Errorf("type %s not supportable", rv.Type())
}

func (e Encoder) writeBool(w io.Writer, v bool) error {
	if v {
		return e.writeByte(w, boolTrue)
	}

	return e.writeByte(w, boolFalse)
}

func (e Encoder) writeInteger(w io.Writer, rv reflect.Value) error {
	v := rv.Int()
	if v >= 0 && v <= 9 {
		return e.writeBytes(w, []byte{0xA0 + byte(v)})
	}

	return e.writeBytes(w, []byte{0xBB}, leb128.AppendSleb128(nil, v))
}

func (e Encoder) writeUint(w io.Writer, rv reflect.Value) error {
	v := rv.Uint()
	if v >= 0 && v <= 9 {
		return e.writeBytes(w, []byte{0xA0 + byte(v)})
	}

	return e.writeBytes(w, []byte{0xBB}, leb128.AppendSleb128(nil, int64(v)))
}

func (e Encoder) writeFloat(w io.Writer, rv reflect.Value) error {
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

func (e Encoder) writeString(w io.Writer, v string) error {
	// must be encoded as fixed-length if:
	// longer than `longStringFactor` bytes, or contains any 0x00 bytes
	if len(v) >= longStringFactor || strings.ContainsRune(v, stringEnd) {
		return e.writeBytes(w, []byte{tagSize}, leb128.AppendUleb128(nil, uint64(len(v))), []byte(v))
	}

	// TODO: with ref ID

	return e.writeBytes(w, []byte(v), []byte{stringEnd})
}

// TODO

func (e Encoder) writeList(w io.Writer, rv reflect.Value) error {
	kind := rv.Kind()

	if tb, ok := kindToType[kind]; ok {
		if err := e.writeBytes(w, []byte{typedArray, tb}); err != nil {
			return err
		}

		if err := e.write(w, rv.Len()); err != nil {
			return err
		}

		// TODO: implement me

		return nil
	}

	if err := e.writeByte(w, listStart); err != nil {
		return err
	}

	for i := 0; i < rv.Len(); i++ {
		if err := e.write(w, rv.Index(i).Interface()); err != nil {
			return err
		}
	}

	if err := e.writeByte(w, listEnd); err != nil {
		return err
	}

	return nil
}

func (e Encoder) writeMap(w io.Writer, rv reflect.Value) error {
	if err := e.writeBytes(w, []byte{dictStart}); err != nil {
		return err
	}

	for _, k := range rv.MapKeys() {
		iv := rv.MapIndex(k)
		// TODO: type validation

		if err := e.write(w, iv.Interface()); err != nil {
			return err
		}

		if err := e.write(w, k.Interface()); err != nil {
			return err
		}
	}

	if err := e.writeBytes(w, []byte{dictEnd}); err != nil {
		return err
	}

	return nil
}

func (e Encoder) writeStruct(w io.Writer, rv reflect.Value) error {
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

	if err := e.writeBytes(w, []byte{dictEnd}); err != nil {
		return err
	}

	return nil
}

func (Encoder) writeBytes(w io.Writer, val ...[]byte) error {
	for _, v := range val {
		if _, err := w.Write(v); err != nil {
			return err
		}
	}

	return nil
}

func (Encoder) writeByte(w io.Writer, val byte) error {
	_, err := w.Write([]byte{val})
	return err
}
