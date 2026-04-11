package muon

// Spec compliance tests based on https://github.com/vshymanskyy/muon/blob/main/docs/README.md
// These tests cover all token types, edge cases, and encoding rules from the official spec.
// DO NOT fix failures here — review them first.

import (
	"bytes"
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func encode(t *testing.T, v any) []byte {
	t.Helper()
	var buf bytes.Buffer
	var enc Encoder
	require.NoError(t, enc.Write(&buf, v))
	return buf.Bytes()
}

func encodeWith(t *testing.T, enc *Encoder, v any) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, enc.Write(&buf, v))
	return buf.Bytes()
}

func tokens(t *testing.T, data []byte) []Token {
	t.Helper()
	r := NewByteReader(data)
	var out []Token
	for {
		tok, err := r.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		out = append(out, tok)
	}
	return out
}

// ---------------------------------------------------------------------------
// 1. Strings
// ---------------------------------------------------------------------------

func TestSpec_String_NullTerminated(t *testing.T) {
	// Regular UTF-8 string: bytes + 0x00
	data := encode(t, "hello")
	assert.Equal(t, append([]byte("hello"), 0x00), data)
}

func TestSpec_String_Empty(t *testing.T) {
	// Empty string: just 0x00
	data := encode(t, "")
	assert.Equal(t, []byte{0x00}, data)
}

func TestSpec_String_ContainsNull(t *testing.T) {
	// String with embedded null → fixed-length: 0x8B + ULEB128(len) + bytes
	s := "te\x00st"
	data := encode(t, s)
	assert.Equal(t, byte(tagSize), data[0], "must use size tag for string with null")
	// round-trip
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.Equal(t, TokenString, toks[0].A)
	assert.Equal(t, s, toks[0].Data)
}

func TestSpec_String_Long(t *testing.T) {
	// String >= 512 bytes → fixed-length
	s := string(make([]byte, 512))
	data := encode(t, s)
	assert.Equal(t, byte(tagSize), data[0], "must use size tag for string >= 512 bytes")
}

func TestSpec_String_LRU_FirstOccurrence(t *testing.T) {
	// 0x8C tag must precede the string on first LRU write
	enc := &Encoder{LRU: true}
	data := encodeWith(t, enc, "foo")
	assert.Equal(t, byte(tagRefString), data[0], "first LRU string must be preceded by 0x8C")
	// the string itself follows
	assert.Equal(t, []byte("foo"), data[1:4])
	assert.Equal(t, byte(0x00), data[4])
}

func TestSpec_String_LRU_Reference(t *testing.T) {
	// Second occurrence of same string → 0x81 + ULEB128(index)
	enc := &Encoder{LRU: true}
	var buf bytes.Buffer
	require.NoError(t, enc.Write(&buf, "foo"))
	require.NoError(t, enc.Write(&buf, "foo"))
	data := buf.Bytes()

	r := NewByteReader(data)
	tok1, err := r.Next()
	require.NoError(t, err)
	assert.Equal(t, TokenString, tok1.A)
	assert.Equal(t, "foo", tok1.Data)

	tok2, err := r.Next()
	require.NoError(t, err)
	assert.Equal(t, TokenString, tok2.A)
	assert.Equal(t, "foo", tok2.Data.(string))
}

func TestSpec_String_LRU_MustNotDuplicateInList(t *testing.T) {
	// Spec: encoder should avoid adding the same string to LRU twice
	enc := &Encoder{LRU: true}
	var buf bytes.Buffer
	require.NoError(t, enc.Write(&buf, "x"))
	require.NoError(t, enc.Write(&buf, "x"))
	require.NoError(t, enc.Write(&buf, "x"))

	// Count how many 0x8C tags appear — must be exactly 1
	count := 0
	for _, b := range buf.Bytes() {
		if b == tagRefString {
			count++
		}
	}
	assert.Equal(t, 1, count, "0x8C must appear only once for the same string")
}

// ---------------------------------------------------------------------------
// 2. Integers
// ---------------------------------------------------------------------------

func TestSpec_Int_Inline_0_to_9(t *testing.T) {
	// 0xA0..0xA9 — single byte inline encoding
	for i := 0; i <= 9; i++ {
		data := encode(t, i)
		assert.Equal(t, []byte{0xA0 + byte(i)}, data, "inline int %d", i)
	}
}

func TestSpec_Int_SLEB128_Large(t *testing.T) {
	// Values outside 0-9 → 0xBB + SLEB128
	data := encode(t, 100)
	assert.Equal(t, byte(0xBB), data[0])

	data = encode(t, -1)
	assert.Equal(t, byte(0xBB), data[0])
}

func TestSpec_Int_SLEB128_RoundTrip(t *testing.T) {
	for _, v := range []int{-1, 10, 127, -128, 1000, -1000, math.MaxInt32, math.MinInt32} {
		data := encode(t, v)
		toks := tokens(t, data)
		require.Len(t, toks, 1)
		assert.Equal(t, TokenInt, toks[0].A)
		assert.Equal(t, v, toks[0].Data, "value %d", v)
	}
}

func TestSpec_Uint_SLEB128_RoundTrip(t *testing.T) {
	// uint > 9 also uses 0xBB + SLEB128 per spec
	for _, v := range []uint{10, 255, 1000, 65535} {
		data := encode(t, v)
		toks := tokens(t, data)
		require.Len(t, toks, 1)
		assert.Equal(t, TokenInt, toks[0].A)
	}
}

func TestSpec_TypedInt_Int8(t *testing.T) {
	// Typed integers are only used inside TypedArrays (not standalone)
	// 0xB0 = int8, 1 byte LE
	data := encode(t, []int8{-4, -3, -2, -1, 0, 1, 2, 3, 4})
	assert.Equal(t, byte(typedArray), data[0])
	assert.Equal(t, byte(typeInt8), data[1])
	// count = 9, then 9 raw bytes
	assert.Equal(t, byte(9), data[2])
	assert.Equal(t, byte(0xFC), data[3]) // -4 as uint8
}

func TestSpec_TypedInt_Uint8(t *testing.T) {
	data := encode(t, []uint8{0, 1, 2, 3, 4})
	assert.Equal(t, byte(typedArray), data[0])
	assert.Equal(t, byte(typeUint8), data[1])
	assert.Equal(t, byte(5), data[2])
}

func TestSpec_TypedInt_Int32(t *testing.T) {
	data := encode(t, []int32{-4, -3, -2, -1, 0, 1, 2, 3, 4})
	assert.Equal(t, byte(typedArray), data[0])
	assert.Equal(t, byte(typeInt32), data[1])
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.Equal(t, TokenTypedArray, toks[0].A)
	assert.Equal(t, []int32{-4, -3, -2, -1, 0, 1, 2, 3, 4}, toks[0].Data)
}

func TestSpec_TypedInt_Int64(t *testing.T) {
	data := encode(t, []int64{-4, -3, -2, -1, 0, 1, 2, 3, 4})
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.Equal(t, []int64{-4, -3, -2, -1, 0, 1, 2, 3, 4}, toks[0].Data)
}

func TestSpec_TypedFloat_Float32(t *testing.T) {
	data := encode(t, []float32{1.2, 3.4, 5.6})
	assert.Equal(t, byte(typedArray), data[0])
	assert.Equal(t, byte(typeFloat32), data[1])
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.Equal(t, TokenTypedArray, toks[0].A)
	result := toks[0].Data.([]float32)
	assert.InDelta(t, 1.2, float64(result[0]), 1e-5)
}

func TestSpec_TypedFloat_Float64(t *testing.T) {
	data := encode(t, []float64{1.2, 3.4, 5.6})
	assert.Equal(t, byte(typedArray), data[0])
	assert.Equal(t, byte(typeFloat64), data[1])
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	result := toks[0].Data.([]float64)
	assert.InDelta(t, 1.2, result[0], 1e-10)
}

// ---------------------------------------------------------------------------
// 3. Floats
// ---------------------------------------------------------------------------

func TestSpec_Float_F64_Pi(t *testing.T) {
	data := encode(t, math.Pi)
	assert.Equal(t, byte(floatF64), data[0])
	toks := tokens(t, data)
	assert.InDelta(t, math.Pi, toks[0].Data.(float64), 1e-12)
}

func TestSpec_Float_NaN(t *testing.T) {
	data := encode(t, math.NaN())
	assert.Equal(t, []byte{nanValue}, data)
	toks := tokens(t, data)
	assert.True(t, math.IsNaN(toks[0].Data.(float64)))
}

func TestSpec_Float_NegInf(t *testing.T) {
	data := encode(t, math.Inf(-1))
	assert.Equal(t, []byte{negativeInfValue}, data)
	toks := tokens(t, data)
	assert.True(t, math.IsInf(toks[0].Data.(float64), -1))
}

func TestSpec_Float_PosInf(t *testing.T) {
	data := encode(t, math.Inf(1))
	assert.Equal(t, []byte{positiveInfValue}, data)
}

func TestSpec_Float_F32_StandaloneWrite(t *testing.T) {
	// Spec: float32 is 0xB9 + 4 bytes. Writer uses float64 (0xBA) for all standalone floats.
	// This test checks that a float32 value written standalone uses f64 encoding.
	data := encode(t, float32(1.5))
	assert.Equal(t, byte(floatF64), data[0], "standalone float32 must be encoded as f64")
}

func TestSpec_Float_F16_Read(t *testing.T) {
	// 0xB8 = float16 — reader must handle it (writer doesn't emit it)
	// float16 of 1.0 = 0x3C00 (little-endian: 0x00, 0x3C)
	f16data := []byte{0xB8, 0x00, 0x3C}
	toks := tokens(t, f16data)
	require.Len(t, toks, 1)
	assert.Equal(t, TokenFloat, toks[0].A, "0xB8 must produce TokenFloat")
	if toks[0].A == TokenFloat {
		assert.InDelta(t, 1.0, toks[0].Data.(float64), 0.01)
	}
}

// ---------------------------------------------------------------------------
// 4. Special values
// ---------------------------------------------------------------------------

func TestSpec_Special_True(t *testing.T) {
	assert.Equal(t, []byte{boolTrue}, encode(t, true))
}

func TestSpec_Special_False(t *testing.T) {
	assert.Equal(t, []byte{boolFalse}, encode(t, false))
}

func TestSpec_Special_Null(t *testing.T) {
	assert.Equal(t, []byte{nilValue}, encode(t, nil))
}

// ---------------------------------------------------------------------------
// 5. TypedArray
// ---------------------------------------------------------------------------

func TestSpec_TypedArray_AllIntTypes(t *testing.T) {
	cases := []struct {
		name string
		v    any
		tb   byte
	}{
		{"int8", []int8{1, 2, 3}, typeInt8},
		{"int16", []int16{1, 2, 3}, typeInt16},
		{"int32", []int32{1, 2, 3}, typeInt32},
		{"int64", []int64{1, 2, 3}, typeInt64},
		{"uint8", []uint8{1, 2, 3}, typeUint8},
		{"uint16", []uint16{1, 2, 3}, typeUint16},
		{"uint32", []uint32{1, 2, 3}, typeUint32},
		{"uint64", []uint64{1, 2, 3}, typeUint64},
		{"float32", []float32{1, 2, 3}, typeFloat32},
		{"float64", []float64{1, 2, 3}, typeFloat64},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := encode(t, tc.v)
			assert.Equal(t, byte(typedArray), data[0])
			assert.Equal(t, tc.tb, data[1])
		})
	}
}

func TestSpec_TypedArray_Chunked_Write(t *testing.T) {
	var buf bytes.Buffer
	var enc Encoder
	require.NoError(t, enc.WriteChunkedTypedArray(&buf, typeUint8,
		[]uint8{0, 1, 2, 3, 4},
		[]uint8{5, 6, 7, 8, 9},
		[]uint8{10, 11, 12, 13, 14},
	))
	data := buf.Bytes()
	assert.Equal(t, byte(typedArrayChunk), data[0])
	assert.Equal(t, byte(typeUint8), data[1])

	// round-trip through reader
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.Equal(t, TokenTypedArray, toks[0].A)
	assert.Equal(t, []uint8{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}, toks[0].Data)
}

func TestSpec_TypedArray_Empty(t *testing.T) {
	data := encode(t, []int32{})
	assert.Equal(t, byte(typedArray), data[0])
	assert.Equal(t, byte(typeInt32), data[1])
	assert.Equal(t, byte(0x00), data[2]) // count = 0
}

// ---------------------------------------------------------------------------
// 6. List
// ---------------------------------------------------------------------------

func TestSpec_List_Empty(t *testing.T) {
	data := encode(t, []any{})
	assert.Equal(t, []byte{listStart, listEnd}, data)
}

func TestSpec_List_Mixed(t *testing.T) {
	data := encode(t, []any{"a", 1, true, nil})
	toks := tokens(t, data)
	assert.Equal(t, TokenListStart, toks[0].A)
	assert.Equal(t, TokenString, toks[1].A)
	assert.Equal(t, TokenInt, toks[2].A)
	assert.Equal(t, TokenTrue, toks[3].A)
	assert.Equal(t, TokenNil, toks[4].A)
	assert.Equal(t, TokenListEnd, toks[5].A)
}

func TestSpec_List_Nested(t *testing.T) {
	data := encode(t, []any{[]any{}, []any{}})
	toks := tokens(t, data)
	// outer [ inner[] inner[] ]
	assert.Equal(t, TokenListStart, toks[0].A)
	assert.Equal(t, TokenListStart, toks[1].A)
	assert.Equal(t, TokenListEnd, toks[2].A)
	assert.Equal(t, TokenListStart, toks[3].A)
	assert.Equal(t, TokenListEnd, toks[4].A)
	assert.Equal(t, TokenListEnd, toks[5].A)
}

// ---------------------------------------------------------------------------
// 7. Dict
// ---------------------------------------------------------------------------

func TestSpec_Dict_Empty(t *testing.T) {
	data := encode(t, map[string]int{})
	assert.Equal(t, []byte{dictStart, dictEnd}, data)
}

func TestSpec_Dict_StringKeys(t *testing.T) {
	data := encode(t, map[string]any{"key": "val"})
	toks := tokens(t, data)
	assert.Equal(t, TokenDictStart, toks[0].A)
	assert.Equal(t, TokenString, toks[1].A)
	assert.Equal(t, "key", toks[1].Data)
	assert.Equal(t, TokenString, toks[2].A)
	assert.Equal(t, "val", toks[2].Data)
	assert.Equal(t, TokenDictEnd, toks[3].A)
}

func TestSpec_Dict_IntKeys_Int8(t *testing.T) {
	// First key: 0xB0 (typeInt8) + value; subsequent keys: value only (no type prefix)
	m := map[int8]string{1: "a", 2: "b"}
	data := encode(t, m)
	assert.Equal(t, byte(dictStart), data[0])
	assert.Equal(t, byte(typeInt8), data[1], "first int key must have type prefix")
}

func TestSpec_Dict_IntKeys_Int64(t *testing.T) {
	m := map[int64]string{10: "a"}
	data := encode(t, m)
	assert.Equal(t, byte(dictStart), data[0])
	assert.Equal(t, byte(typeInt64), data[1])
}

func TestSpec_Dict_IntKeys_SLEB128(t *testing.T) {
	// int/uint (platform-dependent) uses 0xBB prefix for first key
	m := map[int]string{42: "x"}
	data := encode(t, m)
	assert.Equal(t, byte(dictStart), data[0])
	assert.Equal(t, byte(0xBB), data[1])
}

func TestSpec_Dict_MixedKeyTypes_Error(t *testing.T) {
	// map[any]any allows mixing key types at runtime.
	// Spec: all keys must be the same type — encoder must return an error.
	m := map[any]any{
		"string_key": "v1",
		42:           "v2",
	}
	var buf bytes.Buffer
	var enc Encoder
	err := enc.Write(&buf, m)
	assert.Error(t, err, "mixed string+int keys must be rejected")
}

func TestSpec_Dict_Struct(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	data := encode(t, Person{Name: "Alice", Age: 30})
	toks := tokens(t, data)
	assert.Equal(t, TokenDictStart, toks[0].A)
	assert.Equal(t, TokenString, toks[1].A)
	assert.Equal(t, "name", toks[1].Data, "field name must be lowercased per muon tag default")
}

func TestSpec_Dict_StructTag_Custom(t *testing.T) {
	type S struct {
		Foo string `muon:"bar"`
	}
	toks := tokens(t, encode(t, S{Foo: "v"}))
	assert.Equal(t, "bar", toks[1].Data)
}

func TestSpec_Dict_StructTag_Skip(t *testing.T) {
	type S struct {
		Visible string
		Hidden  string `muon:"-"`
	}
	toks := tokens(t, encode(t, S{Visible: "v", Hidden: "h"}))
	// dictStart + "visible" + "v" + dictEnd — Hidden must be absent
	assert.Len(t, toks, 4)
}

func TestSpec_Dict_DuplicateKeys_NotAllowed(t *testing.T) {
	// Go maps enforce unique keys natively so encoding always produces valid output.
	// This test documents the expectation rather than triggering an error.
	m := map[string]int{"a": 1}
	data := encode(t, m)
	r := NewByteReader(data)
	keyCount := 0
	r.Next() // dictStart
	for {
		tok, _ := r.Next()
		if tok.A == TokenDictEnd {
			break
		}
		if tok.A == TokenString {
			keyCount++
			r.Next() // value
		}
	}
	assert.Equal(t, 1, keyCount)
}

// ---------------------------------------------------------------------------
// 8. Tags
// ---------------------------------------------------------------------------

func TestSpec_Tag_Magic_Write(t *testing.T) {
	var buf bytes.Buffer
	var enc Encoder
	require.NoError(t, enc.WriteWithMagic(&buf, "hi"))
	data := buf.Bytes()
	assert.Equal(t, []byte{tagMagicByte, 0xB5, 0x30, 0x31}, data[:4])
}

func TestSpec_Tag_Magic_Read_Skipped(t *testing.T) {
	// TokenMagic is returned but the next Next() call returns the actual value
	var buf bytes.Buffer
	var enc Encoder
	enc.WriteWithMagic(&buf, true)
	r := NewByteReader(buf.Bytes())
	tok, err := r.Next()
	require.NoError(t, err)
	assert.Equal(t, TokenMagic, tok.A)
	tok, err = r.Next()
	require.NoError(t, err)
	assert.Equal(t, TokenTrue, tok.A)
}

func TestSpec_Tag_Magic_Decoder_Transparent(t *testing.T) {
	// Decoder must skip magic and return the actual value
	var buf bytes.Buffer
	var enc Encoder
	enc.WriteWithMagic(&buf, "hello")
	d := NewDecoder(buf.Bytes())
	v, err := d.Decode()
	require.NoError(t, err)
	assert.Equal(t, "hello", v)
}

func TestSpec_Tag_Padding_Write(t *testing.T) {
	var buf bytes.Buffer
	var enc Encoder
	require.NoError(t, enc.WritePadding(&buf, 4))
	assert.Equal(t, []byte{0xFF, 0xFF, 0xFF, 0xFF}, buf.Bytes())
}

func TestSpec_Tag_Padding_Read_Skipped(t *testing.T) {
	// Reader must skip 0xFF bytes before each token
	data := []byte{0xFF, 0xFF, 0xFF, boolTrue}
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.Equal(t, TokenTrue, toks[0].A)
}

func TestSpec_Tag_Count_Read(t *testing.T) {
	// 0x8A + ULEB128(n) → TokenCount{Data: n}
	data := []byte{tagCount, 0x05, listStart, listEnd}
	r := NewByteReader(data)
	tok, err := r.Next()
	require.NoError(t, err)
	assert.Equal(t, TokenCount, tok.A)
	assert.Equal(t, uint64(5), tok.Data)
}

func TestSpec_Tag_Count_Decoder_Transparent(t *testing.T) {
	// Count tag must be transparent in the high-level Decoder
	data := []byte{tagCount, 0x03, listStart,
		0xA1, 0xA2, 0xA3, // 1, 2, 3 inline
		listEnd}
	d := NewDecoder(data)
	v, err := d.Decode()
	require.NoError(t, err)
	assert.Equal(t, []any{1, 2, 3}, v)
}

func TestSpec_Tag_Size_String(t *testing.T) {
	// 0x8B = size tag for strings (already used for long/null strings)
	s := "abc"
	// manually construct: 0x8B + ULEB128(3) + "abc" — reader must handle it
	data := []byte{tagSize, 0x03, 'a', 'b', 'c'}
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.Equal(t, TokenString, toks[0].A)
	assert.Equal(t, s, toks[0].Data)
}

// ---------------------------------------------------------------------------
// 9. Deterministic encoding
// ---------------------------------------------------------------------------

func TestSpec_Deterministic_SameInputSameBytes(t *testing.T) {
	enc := &Encoder{Deterministic: true}
	m := map[string]int{"c": 3, "a": 1, "b": 2}
	b1 := encodeWith(t, enc, m)
	enc2 := &Encoder{Deterministic: true}
	b2 := encodeWith(t, enc2, m)
	assert.Equal(t, b1, b2, "deterministic encoder must produce identical bytes for same input")
}

func TestSpec_Deterministic_StringKeys_Sorted(t *testing.T) {
	enc := &Encoder{Deterministic: true}
	m := map[string]int{"z": 26, "a": 1, "m": 13}
	data := encodeWith(t, enc, m)

	r := NewByteReader(data)
	r.Next() // dictStart
	var keys []string
	for {
		tok, _ := r.Next()
		if tok.A == TokenDictEnd {
			break
		}
		if tok.A == TokenString {
			keys = append(keys, tok.Data.(string))
			r.Next() // value
		}
	}
	assert.Equal(t, []string{"a", "m", "z"}, keys)
}

func TestSpec_Deterministic_IntKeys_Sorted(t *testing.T) {
	enc := &Encoder{Deterministic: true}
	m := map[int32]string{30: "c", 10: "a", 20: "b"}
	data := encodeWith(t, enc, m)

	// Decode and verify keys are in sorted order
	d := NewDecoder(data)
	v, err := d.Decode()
	require.NoError(t, err)

	result := v.(map[any]any)
	assert.Equal(t, "a", result[10])
	assert.Equal(t, "b", result[20])
	assert.Equal(t, "c", result[30])

	// also verify key ordering via raw Reader
	r := NewByteReader(data)
	tok, _ := r.Next() // dictStart
	require.Equal(t, TokenDictStart, tok.A)
	// first key has type prefix
	tok, _ = r.Next()
	require.Equal(t, TokenInt, tok.A)
	var keys []int32
	keys = append(keys, int32(tok.Data.(int)))
	r.Next() // value
	for {
		tok, err = r.NextIntKey(r.lastIntKeyType)
		require.NoError(t, err)
		if tok.A == TokenDictEnd {
			break
		}
		keys = append(keys, int32(tok.Data.(int)))
		r.Next() // value
	}
	assert.Equal(t, []int32{10, 20, 30}, keys)
}

func TestSpec_Deterministic_NoLRU(t *testing.T) {
	enc := &Encoder{LRU: true, Deterministic: true}
	var buf bytes.Buffer
	require.NoError(t, enc.Write(&buf, "repeat"))
	require.NoError(t, enc.Write(&buf, "repeat"))
	for _, b := range buf.Bytes() {
		assert.NotEqual(t, byte(tagRefString), b, "0x8C must not appear in deterministic mode")
		assert.NotEqual(t, byte(stringRef), b, "0x81 must not appear in deterministic mode")
	}
}

func TestSpec_Deterministic_InlineIntegers(t *testing.T) {
	// 0-9 must use 0xA0..0xA9 even in deterministic mode
	enc := &Encoder{Deterministic: true}
	for i := 0; i <= 9; i++ {
		data := encodeWith(t, enc, i)
		assert.Equal(t, []byte{0xA0 + byte(i)}, data)
	}
}

func TestSpec_Deterministic_FloatSpecialValues(t *testing.T) {
	enc := &Encoder{Deterministic: true}
	assert.Equal(t, []byte{nanValue}, encodeWith(t, enc, math.NaN()))
	assert.Equal(t, []byte{negativeInfValue}, encodeWith(t, enc, math.Inf(-1)))
	assert.Equal(t, []byte{positiveInfValue}, encodeWith(t, enc, math.Inf(1)))
}

func TestSpec_Deterministic_FloatF64(t *testing.T) {
	enc := &Encoder{Deterministic: true}
	data := encodeWith(t, enc, 1.5)
	assert.Equal(t, byte(floatF64), data[0])
}

// ---------------------------------------------------------------------------
// 10. Chaining / Streaming Decoder
// ---------------------------------------------------------------------------

func TestSpec_Chaining_MultipleRoots(t *testing.T) {
	var buf bytes.Buffer
	var enc Encoder
	enc.Write(&buf, "first")
	enc.Write(&buf, 42)
	enc.Write(&buf, true)

	d := NewDecoder(buf.Bytes())
	v1, err := d.Decode()
	require.NoError(t, err)
	assert.Equal(t, "first", v1)

	v2, err := d.Decode()
	require.NoError(t, err)
	assert.Equal(t, 42, v2)

	v3, err := d.Decode()
	require.NoError(t, err)
	assert.Equal(t, true, v3)

	_, err = d.Decode()
	assert.Equal(t, io.EOF, err)
}

func TestSpec_Chaining_PaddingBetweenObjects(t *testing.T) {
	// 0xFF padding between objects must be transparent
	data := []byte{
		boolTrue,
		tagPadding, tagPadding, tagPadding,
		boolFalse,
	}
	d := NewDecoder(data)
	v1, err := d.Decode()
	require.NoError(t, err)
	assert.Equal(t, true, v1)
	v2, err := d.Decode()
	require.NoError(t, err)
	assert.Equal(t, false, v2)
}

func TestSpec_Chaining_MagicMidStream(t *testing.T) {
	// Magic can appear multiple times in a chained stream; Decoder must skip it
	var buf bytes.Buffer
	var enc Encoder
	enc.WriteWithMagic(&buf, "a")
	enc.WriteWithMagic(&buf, "b")

	d := NewDecoder(buf.Bytes())
	v1, err := d.Decode()
	require.NoError(t, err)
	assert.Equal(t, "a", v1)

	v2, err := d.Decode()
	require.NoError(t, err)
	assert.Equal(t, "b", v2)
}

func TestSpec_Decoder_NestedList(t *testing.T) {
	data := encode(t, []any{[]any{"x"}, []any{"y"}})
	d := NewDecoder(data)
	v, err := d.Decode()
	require.NoError(t, err)
	outer := v.([]any)
	require.Len(t, outer, 2)
	assert.Equal(t, "x", outer[0].([]any)[0])
}

func TestSpec_Decoder_NestedDict(t *testing.T) {
	enc := &Encoder{Deterministic: true}
	data := encodeWith(t, enc, map[string]any{
		"inner": map[string]any{"k": "v"},
	})
	d := NewDecoder(data)
	v, err := d.Decode()
	require.NoError(t, err)
	outer := v.(map[string]any)
	inner := outer["inner"].(map[string]any)
	assert.Equal(t, "v", inner["k"])
}

func TestSpec_Decoder_TypedArray(t *testing.T) {
	data := encode(t, []float64{1.1, 2.2, 3.3})
	d := NewDecoder(data)
	v, err := d.Decode()
	require.NoError(t, err)
	arr := v.([]float64)
	assert.InDelta(t, 1.1, arr[0], 1e-10)
}

// ---------------------------------------------------------------------------
// 11. Pointer / interface passthrough
// ---------------------------------------------------------------------------

func TestSpec_Pointer_Dereference(t *testing.T) {
	s := "pointed"
	data := encode(t, &s)
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.Equal(t, "pointed", toks[0].Data)
}

func TestSpec_Pointer_Nil(t *testing.T) {
	var p *string
	// nil pointer — writer panics or errors: reflect.Value.Interface on zero Value
	// Expected: encode as nilValue (0xAC) or return a graceful error
	var buf bytes.Buffer
	var enc Encoder
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("nil pointer caused panic: %v — should encode as nil or return error", r)
		}
	}()
	err := enc.Write(&buf, p)
	if err == nil {
		toks := tokens(t, buf.Bytes())
		require.Len(t, toks, 1)
		assert.Equal(t, TokenNil, toks[0].A)
	}
}

// ---------------------------------------------------------------------------
// 12. Custom Marshaler interfaces
// ---------------------------------------------------------------------------

type customMarshal struct{ val string }

func (c customMarshal) MarshalMuon() ([]byte, error) {
	return append([]byte(c.val), 0x00), nil
}

func TestSpec_Marshaler_Interface(t *testing.T) {
	data := encode(t, customMarshal{"custom"})
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.Equal(t, "custom", toks[0].Data)
}

type streamMarshal struct{ val string }

func (s streamMarshal) MarshalMuon(w io.Writer) error {
	w.Write(append([]byte(s.val), 0x00))
	return nil
}

func TestSpec_MarshalerStream_Interface(t *testing.T) {
	data := encode(t, streamMarshal{"stream"})
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.Equal(t, "stream", toks[0].Data)
}

// ---------------------------------------------------------------------------
// 13. readTypedElems — all types via round-trip
// ---------------------------------------------------------------------------

func TestSpec_ReadTypedElems_AllTypes(t *testing.T) {
	cases := []struct {
		name string
		v    any
	}{
		{"int8", []int8{-4, -3, -2, -1, 0, 1, 2, 3, 4}},
		{"int16", []int16{-4, -3, -2, -1, 0, 1, 2, 3, 4}},
		{"int32", []int32{-4, -3, -2, -1, 0, 1, 2, 3, 4}},
		{"int64", []int64{-4, -3, -2, -1, 0, 1, 2, 3, 4}},
		{"uint8", []uint8{0, 1, 2, 3, 4}},
		{"uint16", []uint16{0, 1, 2, 3, 4}},
		{"uint32", []uint32{0, 1, 2, 3, 4}},
		{"uint64", []uint64{0, 1, 2, 3, 4}},
		{"float32", []float32{1.5, 2.5, 3.5}},
		{"float64", []float64{1.5, 2.5, 3.5}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := encode(t, tc.v)
			toks := tokens(t, data)
			require.Len(t, toks, 1)
			assert.Equal(t, TokenTypedArray, toks[0].A)
			assert.Equal(t, tc.v, toks[0].Data)
		})
	}
}

// ---------------------------------------------------------------------------
// 14. mergeTypedSlices — all types via chunked TypedArray
// ---------------------------------------------------------------------------

func TestSpec_MergeTypedSlices_AllTypes(t *testing.T) {
	cases := []struct {
		name     string
		typeByte byte
		chunks   []any
		want     any
	}{
		{"int8", typeInt8, []any{[]int8{1, 2}, []int8{3, 4}}, []int8{1, 2, 3, 4}},
		{"int16", typeInt16, []any{[]int16{1, 2}, []int16{3, 4}}, []int16{1, 2, 3, 4}},
		{"int32", typeInt32, []any{[]int32{1, 2}, []int32{3, 4}}, []int32{1, 2, 3, 4}},
		{"int64", typeInt64, []any{[]int64{1, 2}, []int64{3, 4}}, []int64{1, 2, 3, 4}},
		{"uint8", typeUint8, []any{[]uint8{1, 2}, []uint8{3, 4}}, []uint8{1, 2, 3, 4}},
		{"uint16", typeUint16, []any{[]uint16{1, 2}, []uint16{3, 4}}, []uint16{1, 2, 3, 4}},
		{"uint32", typeUint32, []any{[]uint32{1, 2}, []uint32{3, 4}}, []uint32{1, 2, 3, 4}},
		{"uint64", typeUint64, []any{[]uint64{1, 2}, []uint64{3, 4}}, []uint64{1, 2, 3, 4}},
		{"float32", typeFloat32, []any{[]float32{1.5, 2.5}, []float32{3.5}}, []float32{1.5, 2.5, 3.5}},
		{"float64", typeFloat64, []any{[]float64{1.5, 2.5}, []float64{3.5}}, []float64{1.5, 2.5, 3.5}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			var enc Encoder
			require.NoError(t, enc.WriteChunkedTypedArray(&buf, tc.typeByte, tc.chunks...))
			toks := tokens(t, buf.Bytes())
			require.Len(t, toks, 1)
			assert.Equal(t, TokenTypedArray, toks[0].A)
			assert.Equal(t, tc.want, toks[0].Data)
		})
	}
}

// ---------------------------------------------------------------------------
// 15. float16ToFloat64 — subnormal, ±inf, NaN, negative
// ---------------------------------------------------------------------------

func TestSpec_Float16_Zero(t *testing.T) {
	// +0.0: bits = 0x0000
	data := []byte{0xB8, 0x00, 0x00}
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.Equal(t, float64(0), toks[0].Data.(float64))
}

func TestSpec_Float16_NegZero(t *testing.T) {
	// -0.0: bits = 0x8000
	data := []byte{0xB8, 0x00, 0x80}
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.Equal(t, math.Signbit(toks[0].Data.(float64)), true)
}

func TestSpec_Float16_PosInf(t *testing.T) {
	// +inf: exp=0x1F, mant=0 → bits = 0x7C00
	data := []byte{0xB8, 0x00, 0x7C}
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.True(t, math.IsInf(toks[0].Data.(float64), 1))
}

func TestSpec_Float16_NaN(t *testing.T) {
	// NaN: exp=0x1F, mant≠0 → bits = 0x7E00
	data := []byte{0xB8, 0x00, 0x7E}
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.True(t, math.IsNaN(toks[0].Data.(float64)))
}

func TestSpec_Float16_Subnormal(t *testing.T) {
	// Smallest positive subnormal: bits = 0x0001
	data := []byte{0xB8, 0x01, 0x00}
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	v := toks[0].Data.(float64)
	assert.Greater(t, v, 0.0)
}

func TestSpec_Float16_Normal(t *testing.T) {
	// 1.0 in float16: bits = 0x3C00
	data := []byte{0xB8, 0x00, 0x3C}
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.InDelta(t, 1.0, toks[0].Data.(float64), 1e-4)
}

func TestSpec_Float16_Negative(t *testing.T) {
	// -2.0 in float16: bits = 0xC000
	data := []byte{0xB8, 0x00, 0xC0}
	toks := tokens(t, data)
	require.Len(t, toks, 1)
	assert.InDelta(t, -2.0, toks[0].Data.(float64), 1e-4)
}

// ---------------------------------------------------------------------------
// 16. writeDictIntKey — all integer key types
// ---------------------------------------------------------------------------

func TestSpec_DictKey_AllIntTypes(t *testing.T) {
	cases := []struct {
		name    string
		m       any
		keyByte byte
	}{
		{"uint8", map[uint8]string{1: "a"}, typeUint8},
		{"int8", map[int8]string{1: "a"}, typeInt8},
		{"int16", map[int16]string{1: "a"}, typeInt16},
		{"uint16", map[uint16]string{1: "a"}, typeUint16},
		{"int32", map[int32]string{1: "a"}, typeInt32},
		{"uint32", map[uint32]string{1: "a"}, typeUint32},
		{"int64", map[int64]string{1: "a"}, typeInt64},
		{"uint64", map[uint64]string{1: "a"}, typeUint64},
		{"int", map[int]string{1: "a"}, 0xBB},
		{"uint", map[uint]string{1: "a"}, 0xBB},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := encode(t, tc.m)
			assert.Equal(t, byte(dictStart), data[0])
			assert.Equal(t, tc.keyByte, data[1], "first int key must have correct type prefix")

			// round-trip through Decoder
			d := NewDecoder(data)
			v, err := d.Decode()
			require.NoError(t, err)
			result := v.(map[any]any)
			assert.Len(t, result, 1)
		})
	}
}

func TestSpec_DictKey_MultipleIntKeys_SLEB128(t *testing.T) {
	// map[int] uses 0xBB SLEB128; test multiple keys via Decoder
	enc := &Encoder{Deterministic: true}
	m := map[int]string{10: "a", 20: "b", 30: "c"}
	data := encodeWith(t, enc, m)

	d := NewDecoder(data)
	v, err := d.Decode()
	require.NoError(t, err)
	result := v.(map[any]any)
	assert.Len(t, result, 3)
	assert.Equal(t, "a", result[10])
	assert.Equal(t, "b", result[20])
	assert.Equal(t, "c", result[30])
}

// ---------------------------------------------------------------------------
// 17. writeMap error paths
// ---------------------------------------------------------------------------

func TestSpec_WriteMap_UnsupportedKeyType(t *testing.T) {
	// float64 keys are not allowed — must return error
	m := map[float64]string{1.5: "x"}
	var buf bytes.Buffer
	var enc Encoder
	err := enc.Write(&buf, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dict keys must be string or integer")
}

func TestSpec_WriteMap_MixedKeyTypes(t *testing.T) {
	// map[any]any with mixed string/int keys must error
	m := map[any]any{
		"string_key": "v1",
		42:           "v2",
	}
	var buf bytes.Buffer
	var enc Encoder
	err := enc.Write(&buf, m)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// 18. WriteChunkedTypedArray — non-slice error
// ---------------------------------------------------------------------------

func TestSpec_WriteChunkedTypedArray_NonSliceError(t *testing.T) {
	var buf bytes.Buffer
	var enc Encoder
	err := enc.WriteChunkedTypedArray(&buf, typeInt32, 42) // not a slice
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a slice")
}

// ---------------------------------------------------------------------------
// 19. Decoder edge cases
// ---------------------------------------------------------------------------

func TestSpec_Decoder_EmptyDict(t *testing.T) {
	data := []byte{dictStart, dictEnd}
	d := NewDecoder(data)
	v, err := d.Decode()
	require.NoError(t, err)
	assert.Equal(t, map[string]any{}, v)
}

func TestSpec_Decoder_UnexpectedToken(t *testing.T) {
	// Feed a TokenDictEnd where a value is expected
	data := []byte{dictEnd}
	d := NewDecoder(data)
	_, err := d.Decode()
	assert.Error(t, err)
}

func TestSpec_Decoder_Float(t *testing.T) {
	for _, v := range []float64{-1.5, 0.0, 100.5} {
		data := encode(t, v)
		d := NewDecoder(data)
		result, err := d.Decode()
		require.NoError(t, err)
		assert.InDelta(t, v, result.(float64), 1e-10)
	}
}

func TestSpec_Decoder_TypedArrayAllTypes(t *testing.T) {
	cases := []struct {
		v any
	}{
		{[]int8{1, 2, 3}},
		{[]int16{1, 2, 3}},
		{[]uint16{1, 2, 3}},
		{[]uint32{1, 2, 3}},
		{[]uint64{1, 2, 3}},
		{[]float32{1.5, 2.5}},
	}
	for _, tc := range cases {
		data := encode(t, tc.v)
		d := NewDecoder(data)
		v, err := d.Decode()
		require.NoError(t, err)
		assert.Equal(t, tc.v, v)
	}
}

// ---------------------------------------------------------------------------
// 20. writeStruct — unexported fields are skipped
// ---------------------------------------------------------------------------

func TestSpec_Struct_UnexportedFieldsSkipped(t *testing.T) {
	type S struct {
		Exported   string
		unexported string //nolint
	}
	toks := tokens(t, encode(t, S{Exported: "yes", unexported: "no"}))
	// dictStart + "exported" + "yes" + dictEnd — no "unexported"
	assert.Equal(t, TokenDictStart, toks[0].A)
	assert.Equal(t, "exported", toks[1].Data)
	// verify "unexported" is not present
	for _, tok := range toks {
		if tok.A == TokenString {
			assert.NotEqual(t, "unexported", tok.Data)
		}
	}
}

// ---------------------------------------------------------------------------
// 21. LRU size cap
// ---------------------------------------------------------------------------

func TestSpec_LRU_SizeCap(t *testing.T) {
	// LRU must not grow beyond 512 entries
	enc := &Encoder{LRU: true}
	var buf bytes.Buffer
	for i := 0; i < 600; i++ {
		s := string([]byte{byte(i / 256), byte(i % 256), 0x41}) // unique 3-byte strings
		require.NoError(t, enc.Write(&buf, s))
	}
	assert.LessOrEqual(t, len(enc.lru), lruMaxSize)
}
