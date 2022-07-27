package muon

import (
	"ekyu.moe/leb128"
	"fmt"
	"io"
	"reflect"
)

func Write(w io.Writer, in interface{}) error {
	if in == nil {
		return writeNil(w)
	}

	switch v := in.(type) {
	case string:
		return writeString(w, v)

	case bool:
		return writeBool(w, v)

	case int:
		return writeInt64(w, int64(v))
	case int8:
		return writeInt64(w, int64(v))
	case int16:
		return writeInt64(w, int64(v))
	case int32:
		return writeInt64(w, int64(v))
	case int64:
		return writeInt64(w, v)

	case uint:
		return writeUintInt64(w, uint64(v))
	case uint8:
		return writeUintInt64(w, uint64(v))
	case uint16:
		return writeUintInt64(w, uint64(v))
	case uint32:
		return writeUintInt64(w, uint64(v))
	case uint64:
		return writeUintInt64(w, v)
	}

	rv := reflect.ValueOf(in)

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		return writeSlice(w, rv)
	case reflect.Map:
		return writeMap(w, rv)
	case reflect.Struct:
		return writeStruct(w, rv)
	}

	return fmt.Errorf("type %s not supportable", rv.Type())
}

func writeNil(w io.Writer) error {
	return write(w, []byte{nilValue})
}

func writeString(w io.Writer, v string) error {
	return write(w, []byte(v), []byte{stringEnd})
}

func writeBool(w io.Writer, v bool) error {
	if v {
		return write(w, []byte{boolTrue})
	}

	return write(w, []byte{boolFalse})
}

func writeInt64(w io.Writer, v int64) error {
	if v >= 0 && v <= 9 {
		return write(w, []byte{0xA0 + byte(v)})
	}

	return write(w, leb128.AppendSleb128(nil, v))
}

func writeUintInt64(w io.Writer, v uint64) error {
	if v >= 0 && v <= 9 {
		return write(w, []byte{0xA0 + byte(v)})
	}

	return write(w, leb128.AppendUleb128(nil, v))
}

func writeSlice(w io.Writer, rv reflect.Value) error {
	if err := write(w, []byte{listStart}); err != nil {
		return err
	}

	for i := 0; i < rv.Len(); i++ {
		if err := Write(w, rv.Index(i).Interface()); err != nil {
			return err
		}
	}

	if err := write(w, []byte{listEnd}); err != nil {
		return err
	}

	return nil
}

func writeMap(w io.Writer, rv reflect.Value) error {
	if err := write(w, []byte{dictStart}); err != nil {
		return err
	}

	for _, k := range rv.MapKeys() {
		if err := Write(w, rv.MapIndex(k).Interface()); err != nil {
			return err
		}

		if err := Write(w, k.Interface()); err != nil {
			return err
		}
	}

	if err := write(w, []byte{dictEnd}); err != nil {
		return err
	}

	return nil
}

func writeStruct(w io.Writer, rv reflect.Value) error {
	if err := write(w, []byte{dictStart}); err != nil {
		return err
	}

	tt := rv.Type()

	for i := 0; i < tt.NumField(); i++ {
		tf := tt.Field(i)
		vf := rv.Field(i)

		if err := writeString(w, tf.Name); err != nil {
			return err
		}

		if err := Write(w, vf.Interface()); err != nil {
			return err
		}
	}

	if err := write(w, []byte{dictEnd}); err != nil {
		return err
	}

	return nil
}

func write(w io.Writer, val ...[]byte) error {
	for _, v := range val {
		if _, err := w.Write(v); err != nil {
			return err
		}
	}

	return nil
}
