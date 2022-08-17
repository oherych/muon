package muon

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/go-interpreter/wagon/wasm/leb128"
	"io"
	"reflect"
)

type Decoder struct {
	b *bufio.Reader
}

func NewDecoder(r io.Reader) Decoder {
	return Decoder{
		b: bufio.NewReader(r),
	}
}

func (r *Decoder) Unmarshal(target interface{}) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		if e, ok := r.(error); ok {
			err = e
			return
		}

		err = fmt.Errorf("%v", r)
	}()

	rv := reflect.ValueOf(target)

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("parameter should be pointer")
	}

	rv = rv.Elem()
	if err := r.set(rv); err != nil {
		return err
	}

	if _, err := r.Next(); err != io.EOF {
		panic("expected EOF")
	}

	return nil
}

func (r *Decoder) set(rv reflect.Value) error {
	token, err := r.Next()
	if err != nil {
		return err
	}

	switch token.A {
	case TokenSignature:
		// skip
		return r.set(rv)
	case tokenNil:
		r.setNil(rv)
	case TokenString:
		r.setString(rv, token.D.(string))
	case tokenTrue:
		r.setBool(rv, true)
	case tokenFalse:
		r.setBool(rv, false)
	case TokenNumber:
		r.setNumber(rv, token.D)
	}

	return nil
}

func (r *Decoder) setNil(rv reflect.Value) {
	if !isType(rv, reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice) {
		panic("wrong type")
	}

	rv.Set(reflect.Zero(rv.Type()))
}

func (r *Decoder) setNumber(rv reflect.Value, v interface{}) {
	if !isRange(rv, reflect.Int, reflect.Uint64) {
		panic("wrong type")
	}

	r.setValue(rv, v)
}

func (r *Decoder) setString(rv reflect.Value, v string) {
	if !isType(rv, reflect.String) {
		panic("wrong type")
	}

	r.setValue(rv, v)
}

func (r *Decoder) setBool(rv reflect.Value, v bool) {
	if !isType(rv, reflect.Bool) {
		panic("wrong type")
	}

	r.setValue(rv, v)
}

func (r *Decoder) setValue(target reflect.Value, v interface{}) {
	var n = reflect.ValueOf(v)
	if target.Type() != reflect.TypeOf(v) {
		n = n.Convert(target.Type())
	}

	target.Set(n)
}

func (r *Decoder) Next() (Token, error) {
	first, err := r.b.ReadByte()
	if err != nil {
		return Token{}, err
	}

	if token, ok := tokenMapping[first]; ok {
		return token, nil
	}

	if first == signatureStart {
		return r.readSignature()
	}

	if r.inRange(first, zeroNumber, zeroNumber+9) {
		return Token{A: TokenNumber, D: int(first - zeroNumber)}, nil
	}

	if r.inRange(first, typeInt8, typeFloat64) {
		t, ok := muonTypeToType[first]
		if !ok {
			panic("sd")
		}

		rv := reflect.New(t)
		target := rv.Interface()
		if err := binary.Read(r.b, binary.LittleEndian, target); err != nil {
			return Token{}, err
		}

		return Token{A: TokenNumber, D: rv.Elem().Interface()}, nil
	}

	if first == typedArray {
		tb, err := r.b.ReadByte()
		if err != nil {
			return Token{}, err
		}

		t, ok := muonTypeToType[tb]
		if !ok {
			panic("sd")
		}

		size, err := r.readCount()
		if err != nil {
			return Token{}, err
		}

		rv := reflect.MakeSlice(reflect.SliceOf(t), size, size)

		target := rv.Interface()
		if err := binary.Read(r.b, binary.LittleEndian, target); err != nil {
			return Token{}, err
		}

		return Token{A: TokenTypedArray, D: rv.Interface()}, nil
	}

	if first == stringStart {
		size, err := r.readCount()
		if err != nil {
			return Token{}, err
		}

		buf := make([]byte, size)
		if _, err := r.b.Read(buf); err != nil {
			return Token{}, err
		}

		return Token{A: TokenString, D: string(buf)}, nil
	}

	if first == stringEnd {
		return Token{A: TokenString, D: ""}, nil
	}

	if err := r.b.UnreadByte(); err != nil {
		return Token{}, err
	}

	str, err := r.b.ReadString(stringEnd)
	if err != nil {
		return Token{}, err
	}

	return Token{A: TokenString, D: str[:len(str)-1]}, nil
}

func (Decoder) inRange(v, a, b byte) bool {
	return v >= a && v <= b
}

func (r Decoder) readCount() (int, error) {
	v, err := leb128.ReadVarint64(r.b)

	return int(v), err
}

func (r *Decoder) readSignature() (Token, error) {
	target := make([]byte, len(signature)-1)
	if _, err := r.b.Read(target); err != nil {
		return Token{}, err
	}

	return Token{A: TokenSignature}, nil
}

func isType(rv reflect.Value, exp ...reflect.Kind) bool {
	kind := rv.Kind()

	if kind == reflect.Interface && rv.NumMethod() == 0 {
		return true
	}

	for _, e := range exp {
		if kind == e {
			return true
		}
	}

	return false
}

func isRange(rv reflect.Value, from, to reflect.Kind) bool {
	kind := rv.Kind()

	if kind == reflect.Interface && rv.NumMethod() == 0 {
		return true
	}

	return kind >= from && kind <= to
}
