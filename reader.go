package muon

import (
	"bufio"
	"encoding/binary"
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

	return Token{A: TokenString}, nil
}
