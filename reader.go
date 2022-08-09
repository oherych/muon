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
		return Token{A: token}, nil
	}

	if r.inRange(first, zeroNumber, zeroNumber+9) {
		return Token{A: TokenNumber, Data: int(first - zeroNumber)}, nil
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

		return Token{A: TokenNumber, Data: rv.Elem().Interface()}, nil
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

		return Token{A: TokenString, Data: string(buf)}, nil
	}

	if first == stringEnd {
		return Token{A: TokenString, Data: ""}, nil
	}

	if err := r.b.UnreadByte(); err != nil {
		return Token{}, err
	}

	str, err := r.b.ReadString(stringEnd)
	if err != nil {
		return Token{}, err
	}

	return Token{A: TokenString, Data: str[:len(str)-1]}, nil
}

func (Decoder) inRange(v, a, b byte) bool {
	return v >= a && v <= b
}

func (r Decoder) readCount() (int, error) {
	v, err := leb128.ReadVarint64(r.b)

	return int(v), err
}
