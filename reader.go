package muon

import (
	"encoding/binary"
	"fmt"
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
	// skip padding bytes
	for r.scanp < len(r.in) && r.in[r.scanp] == tagPadding {
		r.scanp++
	}

	if r.scanp >= len(r.in) {
		return Token{}, io.EOF
	}

	first := r.in[r.scanp]
	r.scanp++

	// magic signature: 0x8F 0xB5 0x30 0x31
	if first == tagMagicByte {
		if r.scanp+3 > len(r.in) {
			return Token{}, io.EOF
		}
		r.scanp += 3 // skip 0xB5 0x30 0x31
		return Token{A: TokenMagic}, nil
	}

	// count tag: 0x8A + ULEB128
	if first == tagCount {
		n, size := leb128.DecodeUleb128(r.in[r.scanp:])
		r.scanp += int(size)
		return Token{A: TokenCount, Data: n}, nil
	}

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

	// typed LE integers: 0xB0..0xB7
	if first >= typeInt8 && first <= typeUint64 {
		sizes := [8]int{1, 2, 4, 8, 1, 2, 4, 8} // B0..B7
		size := sizes[first-typeInt8]
		if r.scanp+size > len(r.in) {
			return Token{}, io.EOF
		}
		b := r.in[r.scanp : r.scanp+size]
		r.scanp += size
		signed := first <= typeInt64
		switch size {
		case 1:
			if signed {
				return Token{A: tokenInt, Data: int(int8(b[0]))}, nil
			}
			return Token{A: tokenInt, Data: int(b[0])}, nil
		case 2:
			v := binary.LittleEndian.Uint16(b)
			if signed {
				return Token{A: tokenInt, Data: int(int16(v))}, nil
			}
			return Token{A: tokenInt, Data: int(v)}, nil
		case 4:
			v := binary.LittleEndian.Uint32(b)
			if signed {
				return Token{A: tokenInt, Data: int(int32(v))}, nil
			}
			return Token{A: tokenInt, Data: int(v)}, nil
		case 8:
			v := binary.LittleEndian.Uint64(b)
			if signed {
				return Token{A: tokenInt, Data: int64(v)}, nil
			}
			return Token{A: tokenInt, Data: uint64(v)}, nil
		}
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

	// TypedArray: 0x84 + type_byte + ULEB128(count) + packed LE bytes
	if first == typedArray {
		if r.scanp >= len(r.in) {
			return Token{}, io.EOF
		}
		typeByte := r.in[r.scanp]
		r.scanp++
		count, n := leb128.DecodeUleb128(r.in[r.scanp:])
		r.scanp += int(n)
		data, err := r.readTypedElems(typeByte, int(count))
		if err != nil {
			return Token{}, err
		}
		return Token{A: TokenTypedArray, Data: data}, nil
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

func (r *Reader) readTypedElems(typeByte byte, count int) (interface{}, error) {
	read := func(n int) ([]byte, error) {
		end := r.scanp + n*count
		if end > len(r.in) {
			return nil, io.EOF
		}
		b := r.in[r.scanp:end]
		r.scanp = end
		return b, nil
	}

	switch typeByte {
	case typeInt8:
		b, err := read(1)
		if err != nil {
			return nil, err
		}
		out := make([]int8, count)
		for i := range out {
			out[i] = int8(b[i])
		}
		return out, nil
	case typeInt16:
		b, err := read(2)
		if err != nil {
			return nil, err
		}
		out := make([]int16, count)
		for i := range out {
			out[i] = int16(binary.LittleEndian.Uint16(b[i*2:]))
		}
		return out, nil
	case typeInt32:
		b, err := read(4)
		if err != nil {
			return nil, err
		}
		out := make([]int32, count)
		for i := range out {
			out[i] = int32(binary.LittleEndian.Uint32(b[i*4:]))
		}
		return out, nil
	case typeInt64:
		b, err := read(8)
		if err != nil {
			return nil, err
		}
		out := make([]int64, count)
		for i := range out {
			out[i] = int64(binary.LittleEndian.Uint64(b[i*8:]))
		}
		return out, nil
	case typeUint8:
		b, err := read(1)
		if err != nil {
			return nil, err
		}
		out := make([]uint8, count)
		copy(out, b)
		return out, nil
	case typeUint16:
		b, err := read(2)
		if err != nil {
			return nil, err
		}
		out := make([]uint16, count)
		for i := range out {
			out[i] = binary.LittleEndian.Uint16(b[i*2:])
		}
		return out, nil
	case typeUint32:
		b, err := read(4)
		if err != nil {
			return nil, err
		}
		out := make([]uint32, count)
		for i := range out {
			out[i] = binary.LittleEndian.Uint32(b[i*4:])
		}
		return out, nil
	case typeUint64:
		b, err := read(8)
		if err != nil {
			return nil, err
		}
		out := make([]uint64, count)
		for i := range out {
			out[i] = binary.LittleEndian.Uint64(b[i*8:])
		}
		return out, nil
	case typeFloat32:
		b, err := read(4)
		if err != nil {
			return nil, err
		}
		out := make([]float32, count)
		for i := range out {
			out[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
		}
		return out, nil
	case typeFloat64:
		b, err := read(8)
		if err != nil {
			return nil, err
		}
		out := make([]float64, count)
		for i := range out {
			out[i] = math.Float64frombits(binary.LittleEndian.Uint64(b[i*8:]))
		}
		return out, nil
	}
	return nil, fmt.Errorf("unknown typed array type byte: 0x%02X", typeByte)
}
