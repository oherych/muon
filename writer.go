package muon

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/oherych/muon/internal"

	"github.com/go-interpreter/wagon/wasm/leb128"
)

const (
	longStringFactor = 512
)

const (
	defaultTag = "muon"
)

type Encoder struct {
	WithSignature bool
	TagName       string

	b *bufio.Writer
}

func NewEncoder(w io.Writer) Encoder {
	return Encoder{
		TagName: defaultTag,

		b: bufio.NewWriter(w),
	}
}

func (e Encoder) Marshal(in interface{}) error {
	if e.WithSignature {
		if err := e.writeBytes(Signature); err != nil {
			return err
		}
	}

	if err := e.add(in); err != nil {
		return err
	}

	return e.b.Flush()
}

func (e Encoder) add(in interface{}) error {
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
		return e.b.WriteByte(NullValue)
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
		return e.writeInteger(rv, in)
	}

	if kind >= reflect.Uint && kind <= reflect.Uint64 {
		return e.writeUint(rv, in)
	}

	if kind == reflect.Float32 || kind == reflect.Float64 {
		return e.writeFloat(rv, in)
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

	if kind == reflect.Ptr {
		return e.add(rv.Elem().Interface())
	}

	return fmt.Errorf("type %s not supportable", rv.Type())
}

func (e Encoder) writeBool(v bool) error {
	if v {
		return e.b.WriteByte(BoolTrue)
	}

	return e.b.WriteByte(BoolFalse)
}

func (e Encoder) writeInteger(rv reflect.Value, raw interface{}) error {
	v := rv.Int()
	if v >= 0 && v <= 9 {
		return e.b.WriteByte(ZeroNumber + byte(v))
	}

	if err := e.writeTyped(rv.Kind()); err != nil {
		return err
	}

	if rv.Kind() == reflect.Int {
		raw = v
	}

	return e.writeLittleEndian(raw)

	//_, err := leb128.WriteVarint64(e.b, v)
	//
	//return err
}

func (e Encoder) writeUint(rv reflect.Value, raw interface{}) error {
	v := rv.Uint()
	if v <= 9 {
		return e.b.WriteByte(ZeroNumber + byte(v))
	}

	if err := e.writeTyped(rv.Kind()); err != nil {
		return err
	}

	if rv.Kind() == reflect.Uint {
		raw = v
	}

	return e.writeLittleEndian(raw)

	//return e.writeBytes(leb128.AppendUleb128(nil, v))
}

func (e Encoder) writeFloat(rv reflect.Value, raw interface{}) error {
	v := rv.Float()

	if math.IsNaN(v) {
		return e.b.WriteByte(NanValue)
	}
	if math.IsInf(v, -1) {
		return e.b.WriteByte(NegativeInfValue)
	}
	if math.IsInf(v, 1) {
		return e.b.WriteByte(PositiveInfValue)
	}

	if err := e.writeTyped(rv.Kind()); err != nil {
		return err
	}

	return e.writeLittleEndian(raw)
}

func (e Encoder) writeString(v string) error {
	// TODO: with ref ID

	// must be encoded as fixed-length if:
	// longer than `longStringFactor` bytes, or contains any 0x00 bytes
	if len(v) > longStringFactor || !utf8.ValidString(v) || strings.ContainsRune(v, StringEnd) {
		if err := e.b.WriteByte(CountTag); err != nil {
			return err
		}

		if err := e.writeCount(len(v)); err != nil {
			return err
		}

		return e.writeBytes([]byte(v))
	}

	return e.writeBytes([]byte(v), []byte{StringEnd})
}

func (e Encoder) writeCount(v int) error {
	return e.writeBytes(leb128.AppendUleb128(nil, uint64(v)))
}

func (e Encoder) writeTyped(k reflect.Kind) error {
	tb, ok := KindToMuonType[k]
	if !ok {
		return errors.New("unexpected error: cannot find type in KindToMuonType map")
	}

	return e.b.WriteByte(tb)
}

func (e Encoder) writeLittleEndian(in interface{}) error {
	return binary.Write(e.b, binary.LittleEndian, in)
}

func (e Encoder) writeList(rv reflect.Value) error {
	kind := rv.Type().Elem().Kind()

	if tb, ok := KindToMuonType[kind]; ok {
		if err := e.b.WriteByte(TypedArray); err != nil {
			return err
		}

		if err := e.b.WriteByte(tb); err != nil {
			return err
		}

		if err := e.writeCount(rv.Len()); err != nil {
			return err
		}

		for i := 0; i < rv.Len(); i++ {
			var v interface{}
			switch kind {
			case reflect.Int:
				v = rv.Index(i).Int()
			case reflect.Uint:
				v = rv.Index(i).Uint()
			default:
				v = rv.Index(i).Interface()
			}

			if err := e.writeLittleEndian(v); err != nil {
				return err
			}
		}

		return nil
	}

	if err := e.b.WriteByte(ListStart); err != nil {
		return err
	}

	for i := 0; i < rv.Len(); i++ {
		if err := e.add(rv.Index(i).Interface()); err != nil {
			return err
		}
	}

	if err := e.b.WriteByte(ListEnd); err != nil {
		return err
	}

	return nil
}

func (e Encoder) writeMap(rv reflect.Value) error {
	if err := e.b.WriteByte(DictStart); err != nil {
		return err
	}

	for _, k := range rv.MapKeys() {
		if k.Kind() != reflect.String && KindToMuonType[k.Kind()] == 0 {
			return errors.New("wrong type of dict key")
		}

		iv := rv.MapIndex(k)

		if err := e.add(k.Interface()); err != nil {
			return err
		}

		if err := e.add(iv.Interface()); err != nil {
			return err
		}
	}

	if err := e.b.WriteByte(DictEnd); err != nil {
		return err
	}

	return nil
}

func (e Encoder) writeStruct(rv reflect.Value) error {
	if err := e.b.WriteByte(DictStart); err != nil {
		return err
	}

	tt := rv.Type()

	for i := 0; i < tt.NumField(); i++ {
		tf := tt.Field(i)

		params := internal.ParseTags(e.TagName, tf)
		if params.Skip {
			continue
		}

		if err := e.writeString(params.Name); err != nil {
			return err
		}

		vf := rv.Field(i)
		if err := e.add(vf.Interface()); err != nil {
			return err
		}
	}

	if err := e.b.WriteByte(DictEnd); err != nil {
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
