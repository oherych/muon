package muon

import (
	"bufio"
	"fmt"
	"github.com/go-interpreter/wagon/wasm/leb128"
	"io"
	"math"
	"reflect"
	"strings"
)

const (
	longStringFactor = 512
)

var (
	kindToType = map[reflect.Kind]byte{
		reflect.Int:     0, // TODO:
		reflect.Int8:    typeInt8,
		reflect.Int16:   typeInt16,
		reflect.Int32:   typeInt32,
		reflect.Int64:   typeInt64,
		reflect.Uint:    0, // TODO:
		reflect.Uint8:   typeUint8,
		reflect.Uint16:  typeUint16,
		reflect.Uint32:  typeUint32,
		reflect.Uint64:  typeUint64,
		reflect.Float32: typeFloat32,
		reflect.Float64: typeFloat64,
	}
)

type Encoder struct {
	b *bufio.Writer
}

func NewEncoder(w io.Writer) Encoder {
	return Encoder{
		b: bufio.NewWriter(w),
	}
}

func (e Encoder) Write(in interface{}) error {
	if err := e.write(in); err != nil {
		return err
	}

	return e.b.Flush()
}

func (e Encoder) write(in interface{}) error {
	if m, ok := in.(Marshaler); ok {
		data, err := m.MarshalMuon()
		if err != nil {
			return err
		}

		return e.writeBytes(data)
	}

	if m, ok := in.(MarshalerStream); ok {
		return m.MarshalMuon(e.b)
	}

	if in == nil {
		return e.writeByte(nilValue)
	}

	rv := reflect.ValueOf(in)
	kind := rv.Kind()

	if kind == reflect.Bool {
		return e.writeBool(rv.Bool())
	}

	if kind == reflect.String {
		return e.writeString(rv.String())
	}

	if kind >= reflect.Int && kind <= reflect.Int64 {
		return e.writeInteger(rv)
	}

	if kind >= reflect.Uint && kind <= reflect.Uint64 {
		return e.writeUint(rv)
	}

	if kind == reflect.Float32 || kind == reflect.Float64 {
		return e.writeFloat(rv)
	}

	if kind == reflect.Slice || kind == reflect.Array {
		return e.writeList(rv)
	}

	if kind == reflect.Map {
		return e.writeMap(rv)
	}

	if kind == reflect.Struct {
		return e.writeStruct(rv)
	}

	if kind == reflect.Pointer {
		return e.write(rv.Elem().Interface())
	}

	return fmt.Errorf("type %s not supportable", rv.Type())
}

func (e Encoder) writeBool(v bool) error {
	if v {
		return e.writeByte(boolTrue)
	}

	return e.writeByte(boolFalse)
}

func (e Encoder) writeInteger(rv reflect.Value) error {
	v := rv.Int()
	if v >= 0 && v <= 9 {
		return e.writeByte(0xA0 + byte(v))
	}

	tb, ok := kindToType[rv.Kind()]
	if !ok {
		panic("TODO")
	}

	if err := e.writeByte(tb); err != nil {
		return err
	}

	_, err := leb128.WriteVarint64(e.b, v)

	return err
}

func (e Encoder) writeUint(rv reflect.Value) error {
	v := rv.Uint()
	if v >= 0 && v <= 9 {
		return e.writeByte(0xA0 + byte(v))
	}

	tb, ok := kindToType[rv.Kind()]
	if !ok {
		panic("TODO")
	}

	if err := e.writeByte(tb); err != nil {
		return err
	}

	return e.writeBytes(leb128.AppendUleb128(nil, v))
}

func (e Encoder) writeFloat(rv reflect.Value) error {
	v := rv.Float()

	if math.IsNaN(v) {
		return e.writeByte(nanValue)
	}
	if math.IsInf(v, -1) {
		return e.writeByte(negativeInfValue)
	}
	if math.IsInf(v, 1) {
		return e.writeByte(positiveInfValue)
	}

	panic("implement me")
}

func (e Encoder) writeString(v string) error {
	// must be encoded as fixed-length if:
	// longer than `longStringFactor` bytes, or contains any 0x00 bytes
	if len(v) > longStringFactor || strings.ContainsRune(v, stringEnd) {
		if err := e.writeByte(stringStart); err != nil {
			return err
		}

		if err := e.writeCount(len(v)); err != nil {
			return err
		}

		return e.writeBytes([]byte(v))
	}

	// TODO: with ref ID

	return e.writeBytes([]byte(v), []byte{stringEnd})
}

func (e Encoder) writeCount(v int) error {
	_, err := leb128.WriteVarint64(e.b, int64(v))

	return err
}

// TODO

func (e Encoder) writeList(rv reflect.Value) error {
	kind := rv.Type().Elem().Kind()

	if tb, ok := kindToType[kind]; ok {
		if err := e.writeByte(typedArray); err != nil {
			return err
		}

		if err := e.writeByte(tb); err != nil {
			return err
		}

		if err := e.writeCount(rv.Len()); err != nil {
			return err
		}

		// TODO: implement me

		return nil
	}

	if err := e.writeByte(listStart); err != nil {
		return err
	}

	for i := 0; i < rv.Len(); i++ {
		if err := e.write(rv.Index(i).Interface()); err != nil {
			return err
		}
	}

	if err := e.writeByte(listEnd); err != nil {
		return err
	}

	return nil
}

func (e Encoder) writeMap(rv reflect.Value) error {
	if err := e.writeByte(dictStart); err != nil {
		return err
	}

	for _, k := range rv.MapKeys() {
		iv := rv.MapIndex(k)
		// TODO: type validation

		if err := e.write(k.Interface()); err != nil {
			return err
		}

		if err := e.write(iv.Interface()); err != nil {
			return err
		}
	}

	if err := e.writeByte(dictEnd); err != nil {
		return err
	}

	return nil
}

func (e Encoder) writeStruct(rv reflect.Value) error {
	if err := e.writeByte(dictStart); err != nil {
		return err
	}

	tt := rv.Type()

	for i := 0; i < tt.NumField(); i++ {
		tf := tt.Field(i)
		vf := rv.Field(i)

		if err := e.writeString(tf.Name); err != nil {
			return err
		}

		if err := e.write(vf.Interface()); err != nil {
			return err
		}
	}

	if err := e.writeByte(dictEnd); err != nil {
		return err
	}

	return nil
}

func (e Encoder) writeBytes(val ...[]byte) error {
	for _, v := range val {
		if _, err := e.b.Write(v); err != nil {
			return err
		}
	}

	return nil
}

func (e Encoder) writeByte(val byte) error {
	return e.b.WriteByte(val)
}
