package muon

import (
	"bufio"
	"encoding/binary"
	"github.com/go-interpreter/wagon/wasm/leb128"
	"math"
	"reflect"
)

var (
	TokenMapping = map[byte]Token{
		ListStart:        {A: TokenListStart},
		ListEnd:          {A: TokenListEnd},
		DictStart:        {A: TokenDictStart},
		DictEnd:          {A: TokenDictEnd},
		BoolFalse:        {A: TokenLiteral, D: false},
		BoolTrue:         {A: TokenLiteral, D: true},
		NullValue:        {A: TokenLiteral, D: nil},
		NanValue:         {A: TokenLiteral, D: math.NaN()},
		NegativeInfValue: {A: TokenLiteral, D: math.Inf(-1)},
		PositiveInfValue: {A: TokenLiteral, D: math.Inf(1)},
		StringEnd:        {A: TokenLiteral, D: ""},
	}
)

type Reader struct {
	B  *bufio.Reader
	fn readerFn
}

type readerFn func() (Token, readerFn, error)

func (r *Reader) Reset() {
	r.fn = r.next
}

func (r Reader) Next() (Token, error) {
	token, nf, err := r.fn()

	r.fn = nf

	return token, err
}

func (r Reader) next() (Token, readerFn, error) {
	var (
		sizeTagValue int
	)

readByte:
	first, err := r.B.ReadByte()
	if err != nil {
		return Token{}, nil, err
	}

	if first == CountTag {
		sizeTagValue, err = r.readCount()
		if err != nil {
			return Token{}, nil, err
		}

		goto readByte
	}

	if first == SizeTag {
		// ignore
		_, err = r.readCount()
		if err != nil {
			return Token{}, nil, err
		}

		goto readByte
	}

	if first == SignatureStart {
		token, err := r.readSignature()

		return token, r.next, err
	}

	if token, ok := TokenMapping[first]; ok {
		return token, r.next, nil
	}

	if r.inRange(first, ZeroNumber, ZeroNumber+9) {
		return Token{A: TokenLiteral, D: int(first - ZeroNumber)}, r.next, nil
	}

	if r.inRange(first, TypeInt8, TypeFloat64) {
		token, err := r.readTypedNumber(first)

		return token, r.next, err
	}

	if first == TypedArray {
		token, err := r.readTypedArray()

		return token, r.next, err
	}

	token, err := r.readString(sizeTagValue)

	return token, r.next, err
}

func (r *Reader) readString(sizeTagValue int) (Token, error) {
	if err := r.B.UnreadByte(); err != nil {
		return Token{}, err
	}

	if sizeTagValue > 0 {
		buf := make([]byte, sizeTagValue)
		if _, err := r.B.Read(buf); err != nil {
			return Token{}, err
		}

		return Token{A: TokenLiteral, D: string(buf)}, nil
	}

	str, err := r.B.ReadString(StringEnd)
	if err != nil {
		return Token{}, err
	}

	return Token{A: TokenLiteral, D: str[:len(str)-1]}, nil
}

func (r *Reader) readTypedArray() (Token, error) {
	tb, err := r.B.ReadByte()
	if err != nil {
		return Token{}, err
	}

	t, ok := MuonTypeToType[tb]
	if !ok {
		return Token{}, newError(ErrCodeUnexpectedSystem, "can't find element '%v' in MuonTypeToType", tb)
	}

	size, err := r.readCount()
	if err != nil {
		return Token{}, err
	}

	rv := reflect.MakeSlice(reflect.SliceOf(t), size, size)

	target := rv.Interface()
	if err := binary.Read(r.B, binary.LittleEndian, target); err != nil {
		return Token{}, err
	}

	return Token{A: TokenTypedArray, D: rv.Interface()}, nil
}

func (r *Reader) readTypedNumber(first byte) (Token, error) {
	t, ok := MuonTypeToType[first]
	if !ok {
		return Token{}, newError(ErrCodeUnexpectedSystem, "can't find element '%v' in MuonTypeToType", first)
	}

	rv := reflect.New(t)
	target := rv.Interface()
	if err := binary.Read(r.B, binary.LittleEndian, target); err != nil {
		return Token{}, err
	}

	return Token{A: TokenLiteral, D: rv.Elem().Interface()}, nil
}

func (r *Reader) readSignature() (Token, error) {
	target := make([]byte, len(Signature)-1)
	if _, err := r.B.Read(target); err != nil {
		return Token{}, err
	}

	return Token{A: TokenSignature}, nil
}

func (Reader) inRange(v, a, b byte) bool {
	return v >= a && v <= b
}

func (r Reader) readCount() (int, error) {
	v, err := leb128.ReadVarint64(r.B)

	return int(v), err
}
