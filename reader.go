package muon

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"ekyu.moe/leb128"
)

type Reader struct {
	in             []byte
	scanp          int
	lru            []string
	lastIntKeyType byte // type byte of the most recently decoded typed int key (0xB0..0xB7 or 0xBB)
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
		r.lastIntKeyType = first
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
		r.lastIntKeyType = 0xBB
		v, n := leb128.DecodeSleb128(r.in[r.scanp:])
		r.scanp += int(n)
		return Token{A: tokenInt, Data: int(v)}, nil
	}

	// float16
	if first == 0xB8 {
		if r.scanp+2 > len(r.in) {
			return Token{}, io.EOF
		}
		bits := binary.LittleEndian.Uint16(r.in[r.scanp:])
		r.scanp += 2
		return Token{A: TokenFloat, Data: float16ToFloat64(bits)}, nil
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

	// chunked TypedArray: 0x85 + type_byte + (ULEB128(n) + n×bytes)* + ULEB128(0)
	if first == typedArrayChunk {
		if r.scanp >= len(r.in) {
			return Token{}, io.EOF
		}
		typeByte := r.in[r.scanp]
		r.scanp++
		data, err := r.readChunkedTypedElems(typeByte)
		if err != nil {
			return Token{}, err
		}
		return Token{A: TokenTypedArray, Data: data}, nil
	}

	// string reference: 0x81 + ULEB128(index) → LRU lookup
	if first == stringRef {
		idx, n := leb128.DecodeUleb128(r.in[r.scanp:])
		r.scanp += int(n)
		if int(idx) >= len(r.lru) {
			return Token{}, fmt.Errorf("string ref index %d out of range (lru size %d)", idx, len(r.lru))
		}
		return Token{A: TokenString, Data: r.lru[idx]}, nil
	}

	// referenced string tag: 0x8C — read next string and add to LRU
	if first == tagRefString {
		tok, err := r.Next()
		if err != nil {
			return Token{}, err
		}
		if tok.A != TokenString {
			return Token{}, fmt.Errorf("0x8C tag must be followed by a string, got %s", tok.A)
		}
		s := tok.Data.(string)
		r.lruPrepend(s)
		return tok, nil
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

func (r *Reader) readChunkedTypedElems(typeByte byte) (interface{}, error) {
	// read chunks until zero-length terminator, aggregate into one slice
	var allElems []interface{}
	for {
		count, n := leb128.DecodeUleb128(r.in[r.scanp:])
		r.scanp += int(n)
		if count == 0 {
			break
		}
		chunk, err := r.readTypedElems(typeByte, int(count))
		if err != nil {
			return nil, err
		}
		allElems = append(allElems, chunk)
	}
	if len(allElems) == 0 {
		return r.readTypedElems(typeByte, 0)
	}
	return mergeTypedSlices(allElems), nil
}

func mergeTypedSlices(chunks []interface{}) interface{} {
	switch chunks[0].(type) {
	case []int8:
		var out []int8
		for _, c := range chunks {
			out = append(out, c.([]int8)...)
		}
		return out
	case []int16:
		var out []int16
		for _, c := range chunks {
			out = append(out, c.([]int16)...)
		}
		return out
	case []int32:
		var out []int32
		for _, c := range chunks {
			out = append(out, c.([]int32)...)
		}
		return out
	case []int64:
		var out []int64
		for _, c := range chunks {
			out = append(out, c.([]int64)...)
		}
		return out
	case []uint8:
		var out []uint8
		for _, c := range chunks {
			out = append(out, c.([]uint8)...)
		}
		return out
	case []uint16:
		var out []uint16
		for _, c := range chunks {
			out = append(out, c.([]uint16)...)
		}
		return out
	case []uint32:
		var out []uint32
		for _, c := range chunks {
			out = append(out, c.([]uint32)...)
		}
		return out
	case []uint64:
		var out []uint64
		for _, c := range chunks {
			out = append(out, c.([]uint64)...)
		}
		return out
	case []float32:
		var out []float32
		for _, c := range chunks {
			out = append(out, c.([]float32)...)
		}
		return out
	case []float64:
		var out []float64
		for _, c := range chunks {
			out = append(out, c.([]float64)...)
		}
		return out
	}
	return chunks[0]
}

// NextIntKey reads the next dict integer key given the type byte of the first key.
// Dict keys after the first have no type prefix — the caller must supply the type.
func (r *Reader) NextIntKey(typeByte byte) (Token, error) {
	// skip padding
	for r.scanp < len(r.in) && r.in[r.scanp] == tagPadding {
		r.scanp++
	}
	if r.scanp >= len(r.in) {
		return Token{}, io.EOF
	}
	// check for dictEnd
	if r.in[r.scanp] == dictEnd {
		r.scanp++
		return Token{A: tokenDictEnd}, nil
	}
	// check for SLEB128 (0xBB) int key
	if typeByte == 0xBB {
		v, n := leb128.DecodeSleb128(r.in[r.scanp:])
		r.scanp += int(n)
		return Token{A: tokenInt, Data: int(v)}, nil
	}
	// typed LE integer
	sizes := [8]int{1, 2, 4, 8, 1, 2, 4, 8} // B0..B7
	if typeByte < typeInt8 || typeByte > typeUint64 {
		return Token{}, fmt.Errorf("unexpected dict int key type: 0x%02X", typeByte)
	}
	size := sizes[typeByte-typeInt8]
	if r.scanp+size > len(r.in) {
		return Token{}, io.EOF
	}
	b := r.in[r.scanp : r.scanp+size]
	r.scanp += size
	signed := typeByte <= typeInt64
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
	return Token{}, io.EOF
}

func (r *Reader) lruPrepend(s string) {
	if len(r.lru) >= lruMaxSize {
		r.lru = r.lru[:lruMaxSize-1]
	}
	r.lru = append([]string{s}, r.lru...)
}

// float16ToFloat64 converts an IEEE 754 half-precision (binary16) value to float64.
func float16ToFloat64(bits uint16) float64 {
	sign := uint64(bits>>15) << 63
	exp := (bits >> 10) & 0x1F
	mant := uint64(bits & 0x3FF)

	var f64bits uint64
	switch exp {
	case 0: // subnormal
		if mant == 0 {
			f64bits = sign
		} else {
			// normalize
			exp64 := uint64(1023 - 14)
			for mant&0x400 == 0 {
				mant <<= 1
				exp64--
			}
			mant &^= 0x400
			f64bits = sign | (exp64 << 52) | (mant << 42)
		}
	case 0x1F: // inf or nan
		f64bits = sign | (0x7FF << 52) | (mant << 42)
	default:
		f64bits = sign | (uint64(exp+1023-15) << 52) | (mant << 42)
	}
	return math.Float64frombits(f64bits)
}
