package muon

import (
	"encoding/binary"
	"io"
	"math"

	"ekyu.moe/leb128"
)

type Reader struct {
	in    []byte
	scanp int
}

type Token struct {
	A    TokenEnum
	Data interface{}
}

func NewByteReader(in []byte) Reader {
	return Reader{in: in}
}

func (r *Reader) Next() (Token, error) {
	if r.scanp >= len(r.in) {
		return Token{}, io.EOF
	}

	first := r.in[r.scanp]
	r.scanp++

	if token, ok := tokenMapping[first]; ok {
		return Token{A: token}, nil
	}

	// float special values
	if first == nanValue {
		return Token{A: TokenFloat, Data: math.NaN()}, nil
	}
	if first == negativeInfValue {
		return Token{A: TokenFloat, Data: math.Inf(-1)}, nil
	}
	if first == positiveInfValue {
		return Token{A: TokenFloat, Data: math.Inf(1)}, nil
	}

	// inline integers 0–9
	if first >= 0xA0 && first <= 0xA0+9 {
		return Token{A: tokenInt, Data: int(first - 0xA0)}, nil
	}

	// signed LEB128 integer
	if first == 0xBB {
		v, n := leb128.DecodeSleb128(r.in[r.scanp:])
		r.scanp += int(n)
		return Token{A: tokenInt, Data: int(v)}, nil
	}

	// float64
	if first == floatF64 {
		if r.scanp+8 > len(r.in) {
			return Token{}, io.EOF
		}
		bits := binary.LittleEndian.Uint64(r.in[r.scanp:])
		r.scanp += 8
		return Token{A: TokenFloat, Data: math.Float64frombits(bits)}, nil
	}

	// float32
	if first == 0xB9 {
		if r.scanp+4 > len(r.in) {
			return Token{}, io.EOF
		}
		bits := binary.LittleEndian.Uint32(r.in[r.scanp:])
		r.scanp += 4
		return Token{A: TokenFloat, Data: float64(math.Float32frombits(bits))}, nil
	}

	// size-tagged (fixed-length) string: 0x8B + ULEB128(len) + bytes
	if first == tagSize {
		length, n := leb128.DecodeUleb128(r.in[r.scanp:])
		r.scanp += int(n)
		end := r.scanp + int(length)
		if end > len(r.in) {
			return Token{}, io.EOF
		}
		s := string(r.in[r.scanp:end])
		r.scanp = end
		return Token{A: TokenString, Data: s}, nil
	}

	// empty null-terminated string
	if first == stringEnd {
		return Token{A: TokenString, Data: ""}, nil
	}

	// null-terminated string
	from := r.scanp - 1
	for to := r.scanp; to < len(r.in); to++ {
		if r.in[to] == stringEnd {
			r.scanp = to + 1
			return Token{A: TokenString, Data: string(r.in[from:to])}, nil
		}
	}

	return Token{}, io.EOF
}
