package muon

import (
	"bytes"
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

type G string

var (
	testString = "test"

	tests = map[string]struct {
		golang  interface{}
		encoded []byte
		tokens  []Token
	}{
		"nil": {
			golang:  nil,
			encoded: []byte{nilValue},
			tokens:  []Token{{A: TokenNil}},
		},
		"string_empty": {
			golang:  "",
			encoded: []byte{stringEnd},
			tokens:  []Token{{A: TokenString, Data: ""}},
		},
		"string": {
			golang:  testString,
			encoded: []byte{0x74, 0x65, 0x73, 0x74, stringEnd},
			tokens:  []Token{{A: TokenString, Data: testString}},
		},
		"string_with_zero": {
			golang:  "te" + string([]byte{stringEnd}) + "st",
			encoded: []byte{tagSize, 0x05, 0x74, 0x65, 0x0, 0x73, 0x74},
			tokens:  []Token{{A: TokenString, Data: "te" + string([]byte{stringEnd}) + "st"}},
		},
		"long_string": {
			golang:  "test Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce mi mauris, fringilla a gravida ac, vulputate vitae dui. Proin rhoncus ante vitae purus mollis, id hendrerit tellus tempor. Aliquam ut ex nibh. Aenean quis quam eu purus scelerisque viverra ac consequat justo. Sed lobortis interdum facilisis. Sed euismod est magna, at iaculis nisi mollis a. Maecenas nec diam augue. Phasellus volutpat mattis nisi, eu sagittis enim tempor vitae. Aliquam sit amet ante finibus, bibendum lorem et, porta libero. Sed eu.",
			tokens:  []Token{{A: TokenString, Data: "test Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce mi mauris, fringilla a gravida ac, vulputate vitae dui. Proin rhoncus ante vitae purus mollis, id hendrerit tellus tempor. Aliquam ut ex nibh. Aenean quis quam eu purus scelerisque viverra ac consequat justo. Sed lobortis interdum facilisis. Sed euismod est magna, at iaculis nisi mollis a. Maecenas nec diam augue. Phasellus volutpat mattis nisi, eu sagittis enim tempor vitae. Aliquam sit amet ante finibus, bibendum lorem et, porta libero. Sed eu."}},
			encoded: []byte{tagSize, 0x86, 0x04, 0x74, 0x65, 0x73, 0x74, 0x20, 0x4c, 0x6f, 0x72, 0x65, 0x6d, 0x20, 0x69, 0x70, 0x73, 0x75, 0x6d, 0x20, 0x64, 0x6f, 0x6c, 0x6f, 0x72, 0x20, 0x73, 0x69, 0x74, 0x20, 0x61, 0x6d, 0x65, 0x74, 0x2c, 0x20, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x63, 0x74, 0x65, 0x74, 0x75, 0x72, 0x20, 0x61, 0x64, 0x69, 0x70, 0x69, 0x73, 0x63, 0x69, 0x6e, 0x67, 0x20, 0x65, 0x6c, 0x69, 0x74, 0x2e, 0x20, 0x46, 0x75, 0x73, 0x63, 0x65, 0x20, 0x6d, 0x69, 0x20, 0x6d, 0x61, 0x75, 0x72, 0x69, 0x73, 0x2c, 0x20, 0x66, 0x72, 0x69, 0x6e, 0x67, 0x69, 0x6c, 0x6c, 0x61, 0x20, 0x61, 0x20, 0x67, 0x72, 0x61, 0x76, 0x69, 0x64, 0x61, 0x20, 0x61, 0x63, 0x2c, 0x20, 0x76, 0x75, 0x6c, 0x70, 0x75, 0x74, 0x61, 0x74, 0x65, 0x20, 0x76, 0x69, 0x74, 0x61, 0x65, 0x20, 0x64, 0x75, 0x69, 0x2e, 0x20, 0x50, 0x72, 0x6f, 0x69, 0x6e, 0x20, 0x72, 0x68, 0x6f, 0x6e, 0x63, 0x75, 0x73, 0x20, 0x61, 0x6e, 0x74, 0x65, 0x20, 0x76, 0x69, 0x74, 0x61, 0x65, 0x20, 0x70, 0x75, 0x72, 0x75, 0x73, 0x20, 0x6d, 0x6f, 0x6c, 0x6c, 0x69, 0x73, 0x2c, 0x20, 0x69, 0x64, 0x20, 0x68, 0x65, 0x6e, 0x64, 0x72, 0x65, 0x72, 0x69, 0x74, 0x20, 0x74, 0x65, 0x6c, 0x6c, 0x75, 0x73, 0x20, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x2e, 0x20, 0x41, 0x6c, 0x69, 0x71, 0x75, 0x61, 0x6d, 0x20, 0x75, 0x74, 0x20, 0x65, 0x78, 0x20, 0x6e, 0x69, 0x62, 0x68, 0x2e, 0x20, 0x41, 0x65, 0x6e, 0x65, 0x61, 0x6e, 0x20, 0x71, 0x75, 0x69, 0x73, 0x20, 0x71, 0x75, 0x61, 0x6d, 0x20, 0x65, 0x75, 0x20, 0x70, 0x75, 0x72, 0x75, 0x73, 0x20, 0x73, 0x63, 0x65, 0x6c, 0x65, 0x72, 0x69, 0x73, 0x71, 0x75, 0x65, 0x20, 0x76, 0x69, 0x76, 0x65, 0x72, 0x72, 0x61, 0x20, 0x61, 0x63, 0x20, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x71, 0x75, 0x61, 0x74, 0x20, 0x6a, 0x75, 0x73, 0x74, 0x6f, 0x2e, 0x20, 0x53, 0x65, 0x64, 0x20, 0x6c, 0x6f, 0x62, 0x6f, 0x72, 0x74, 0x69, 0x73, 0x20, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x64, 0x75, 0x6d, 0x20, 0x66, 0x61, 0x63, 0x69, 0x6c, 0x69, 0x73, 0x69, 0x73, 0x2e, 0x20, 0x53, 0x65, 0x64, 0x20, 0x65, 0x75, 0x69, 0x73, 0x6d, 0x6f, 0x64, 0x20, 0x65, 0x73, 0x74, 0x20, 0x6d, 0x61, 0x67, 0x6e, 0x61, 0x2c, 0x20, 0x61, 0x74, 0x20, 0x69, 0x61, 0x63, 0x75, 0x6c, 0x69, 0x73, 0x20, 0x6e, 0x69, 0x73, 0x69, 0x20, 0x6d, 0x6f, 0x6c, 0x6c, 0x69, 0x73, 0x20, 0x61, 0x2e, 0x20, 0x4d, 0x61, 0x65, 0x63, 0x65, 0x6e, 0x61, 0x73, 0x20, 0x6e, 0x65, 0x63, 0x20, 0x64, 0x69, 0x61, 0x6d, 0x20, 0x61, 0x75, 0x67, 0x75, 0x65, 0x2e, 0x20, 0x50, 0x68, 0x61, 0x73, 0x65, 0x6c, 0x6c, 0x75, 0x73, 0x20, 0x76, 0x6f, 0x6c, 0x75, 0x74, 0x70, 0x61, 0x74, 0x20, 0x6d, 0x61, 0x74, 0x74, 0x69, 0x73, 0x20, 0x6e, 0x69, 0x73, 0x69, 0x2c, 0x20, 0x65, 0x75, 0x20, 0x73, 0x61, 0x67, 0x69, 0x74, 0x74, 0x69, 0x73, 0x20, 0x65, 0x6e, 0x69, 0x6d, 0x20, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x20, 0x76, 0x69, 0x74, 0x61, 0x65, 0x2e, 0x20, 0x41, 0x6c, 0x69, 0x71, 0x75, 0x61, 0x6d, 0x20, 0x73, 0x69, 0x74, 0x20, 0x61, 0x6d, 0x65, 0x74, 0x20, 0x61, 0x6e, 0x74, 0x65, 0x20, 0x66, 0x69, 0x6e, 0x69, 0x62, 0x75, 0x73, 0x2c, 0x20, 0x62, 0x69, 0x62, 0x65, 0x6e, 0x64, 0x75, 0x6d, 0x20, 0x6c, 0x6f, 0x72, 0x65, 0x6d, 0x20, 0x65, 0x74, 0x2c, 0x20, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x20, 0x6c, 0x69, 0x62, 0x65, 0x72, 0x6f, 0x2e, 0x20, 0x53, 0x65, 0x64, 0x20, 0x65, 0x75, 0x2e},
		},
		"kind_string": {
			golang:  G(testString),
			encoded: []byte{0x74, 0x65, 0x73, 0x74, stringEnd},
			tokens:  []Token{{A: TokenString, Data: testString}},
		},

		"pointer": {
			golang:  &testString,
			encoded: []byte{0x74, 0x65, 0x73, 0x74, stringEnd},
			tokens:  []Token{{A: TokenString, Data: testString}},
		},

		"true": {
			golang:  true,
			encoded: []byte{boolTrue},
			tokens:  []Token{{A: TokenTrue}},
		},
		"false": {
			golang:  false,
			encoded: []byte{boolFalse},
			tokens:  []Token{{A: TokenFalse}},
		},

		"int": {
			golang:  5,
			encoded: []byte{0xa5},
			tokens:  []Token{{A: TokenInt, Data: 5}},
		},
		"int8":  {golang: int8(8), encoded: []byte{0xa8}, tokens: []Token{{A: TokenInt, Data: 8}}},
		"int16": {golang: int16(16), encoded: []byte{0xbb, 0x10}, tokens: []Token{{A: TokenInt, Data: 16}}},
		"int32": {golang: int32(32), encoded: []byte{0xbb, 0x20}, tokens: []Token{{A: TokenInt, Data: 32}}},
		"int64": {golang: int64(64), encoded: []byte{0xbb, 0xc0, 0x0}, tokens: []Token{{A: TokenInt, Data: 64}}},

		"uint": {
			golang:  uint(5),
			encoded: []byte{0xa5},
			tokens:  []Token{{A: TokenInt, Data: 5}},
		},
		"uint8":  {golang: uint8(8), encoded: []byte{0xa8}, tokens: []Token{{A: TokenInt, Data: 8}}},
		"uint16": {golang: uint16(16), encoded: []byte{0xbb, 0x10}, tokens: []Token{{A: TokenInt, Data: 16}}},
		"uint32": {golang: uint32(32), encoded: []byte{0xbb, 0x20}, tokens: []Token{{A: TokenInt, Data: 32}}},
		"uint64": {golang: uint64(64), encoded: []byte{0xbb, 0xc0, 0x00}, tokens: []Token{{A: TokenInt, Data: 64}}},

		"slice": {
			golang:  []interface{}{testString, true},
			encoded: []byte{listStart, 0x74, 0x65, 0x73, 0x74, stringEnd, boolTrue, listEnd},
			tokens:  []Token{{A: TokenListStart}, {A: TokenString, Data: testString}, {A: TokenTrue}, {A: TokenListEnd}},
		},
		//"typed_array": {
		//	golang: []int{10, 500},
		//},
		"byte_slice": {
			golang:  []byte{1, 30},
			encoded: []byte{typedArray, typeUint8, 0x02, 0x01, 0x1e},
			tokens:  []Token{{A: TokenTypedArray, Data: []uint8{1, 30}}},
		},

		"typed_array_int32": {
			golang:  []int32{10, 500},
			encoded: []byte{typedArray, typeInt32, 0x02, 0x0a, 0x00, 0x00, 0x00, 0xf4, 0x01, 0x00, 0x00},
			tokens:  []Token{{A: TokenTypedArray, Data: []int32{10, 500}}},
		},
		"typed_array_float64": {
			golang:  []float64{1.5},
			encoded: []byte{typedArray, typeFloat64, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf8, 0x3f},
			tokens:  []Token{{A: TokenTypedArray, Data: []float64{1.5}}},
		},

		"array": {
			golang:  [2]bool{false, true},
			encoded: []byte{listStart, boolFalse, boolTrue, listEnd},
			tokens:  []Token{{A: TokenListStart}, {A: TokenFalse}, {A: TokenTrue}, {A: TokenListEnd}},
		},

		"map_string": {
			golang:  map[string]string{"a": "b"},
			encoded: []byte{dictStart, 'a', stringEnd, 'b', stringEnd, dictEnd},
			tokens:  []Token{{A: TokenDictStart}, {A: TokenString, Data: "a"}, {A: TokenString, Data: "b"}, {A: TokenDictEnd}},
		},
		"map_int64_key": {
			golang:  map[int64]string{10: "z"},
			encoded: []byte{dictStart, typeInt64, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 'z', stringEnd, dictEnd},
			tokens:  []Token{{A: TokenDictStart}, {A: TokenInt, Data: int64(10)}, {A: TokenString, Data: "z"}, {A: TokenDictEnd}},
		},

		"float64_pi": {
			golang:  math.Pi,
			encoded: []byte{floatF64, 0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40},
			tokens:  []Token{{A: TokenFloat, Data: math.Pi}},
		},
		"float64_neg": {
			golang:  -1.0,
			encoded: []byte{floatF64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0xbf},
			tokens:  []Token{{A: TokenFloat, Data: -1.0}},
		},
		"float64_zero": {
			golang:  0.0,
			encoded: []byte{floatF64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			tokens:  []Token{{A: TokenFloat, Data: 0.0}},
		},
		"float64_nan": {
			golang:  math.NaN(),
			encoded: []byte{nanValue},
			// NaN != NaN in Go, reader test skipped (tokens == nil)
		},
		"float64_pos_inf": {
			golang:  math.Inf(1),
			encoded: []byte{positiveInfValue},
			tokens:  []Token{{A: TokenFloat, Data: math.Inf(1)}},
		},
		"float64_neg_inf": {
			golang:  math.Inf(-1),
			encoded: []byte{negativeInfValue},
			tokens:  []Token{{A: TokenFloat, Data: math.Inf(-1)}},
		},

		"struct_tags": {
			golang: struct {
				Name   string `muon:"name"`
				Age    int    `muon:"age"`
				Hidden string `muon:"-"`
				NoTag  bool
			}{Name: "alice", Age: 3, Hidden: "x", NoTag: true},
			encoded: []byte{
				dictStart,
				'n', 'a', 'm', 'e', stringEnd,
				'a', 'l', 'i', 'c', 'e', stringEnd,
				'a', 'g', 'e', stringEnd,
				0xa3,
				'n', 'o', 't', 'a', 'g', stringEnd,
				boolTrue,
				dictEnd,
			},
			tokens: []Token{
				{A: TokenDictStart},
				{A: TokenString, Data: "name"},
				{A: TokenString, Data: "alice"},
				{A: TokenString, Data: "age"},
				{A: TokenInt, Data: 3},
				{A: TokenString, Data: "notag"},
				{A: TokenTrue},
				{A: TokenDictEnd},
			},
		},
	}
)

func TestWrite(t *testing.T) {
	for testCase, tt := range tests {
		t.Run(testCase, func(t *testing.T) {
			var writer bytes.Buffer

			var enc Encoder
			err := enc.Write(&writer, tt.golang)

			assert.Equal(t, tt.encoded, writer.Bytes())
			assert.Nil(t, err)
		})
	}
}

func TestWriteMagic(t *testing.T) {
	var buf bytes.Buffer
	var enc Encoder
	err := enc.WriteWithMagic(&buf, true)
	assert.Nil(t, err)
	assert.Equal(t, []byte{tagMagicByte, 0xB5, 0x30, 0x31, boolTrue}, buf.Bytes())
}

func TestWritePadding(t *testing.T) {
	var buf bytes.Buffer
	var enc Encoder
	err := enc.WritePadding(&buf, 3)
	assert.Nil(t, err)
	assert.Equal(t, []byte{tagPadding, tagPadding, tagPadding}, buf.Bytes())
}

func TestLRUStringRefs(t *testing.T) {
	var buf bytes.Buffer
	enc := Encoder{LRU: true}

	// write two values referencing the same strings
	err := enc.Write(&buf, []interface{}{"foo", "bar", "foo"})
	assert.Nil(t, err)

	// first "foo": 0x8C tag + "foo\x00"
	// "bar": 0x8C tag + "bar\x00"
	// second "foo": 0x81 + ULEB128(1)  (foo is at index 1: bar=0, foo=1)
	expected := []byte{
		listStart,
		tagRefString, 'f', 'o', 'o', stringEnd,
		tagRefString, 'b', 'a', 'r', stringEnd,
		stringRef, 0x01,
		listEnd,
	}
	assert.Equal(t, expected, buf.Bytes())

	// reader should reconstruct original strings
	r := NewByteReader(buf.Bytes())
	tokens := []Token{}
	for {
		tok, err := r.Next()
		if err != nil {
			break
		}
		tokens = append(tokens, tok)
	}
	assert.Equal(t, []Token{
		{A: TokenListStart},
		{A: TokenString, Data: "foo"},
		{A: TokenString, Data: "bar"},
		{A: TokenString, Data: "foo"},
		{A: TokenListEnd},
	}, tokens)
}

func TestChunkedTypedArray(t *testing.T) {
	var buf bytes.Buffer
	var enc Encoder
	err := enc.WriteChunkedTypedArray(&buf, typeInt32, []int32{1, 2}, []int32{3, 4})
	assert.Nil(t, err)

	expected := []byte{
		typedArrayChunk, typeInt32,
		0x02, 0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00,
		0x02, 0x03, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00,
		0x00,
	}
	assert.Equal(t, expected, buf.Bytes())

	r := NewByteReader(buf.Bytes())
	tok, err := r.Next()
	assert.Nil(t, err)
	assert.Equal(t, TokenTypedArray, tok.A)
	assert.Equal(t, []int32{1, 2, 3, 4}, tok.Data)
}

func TestReaderMagicAndPadding(t *testing.T) {
	// padding + magic + value
	in := []byte{tagPadding, tagPadding, tagMagicByte, 0xB5, 0x30, 0x31, boolTrue}
	r := NewByteReader(in)

	tok, err := r.Next()
	assert.Nil(t, err)
	assert.Equal(t, TokenMagic, tok.A)

	tok, err = r.Next()
	assert.Nil(t, err)
	assert.Equal(t, TokenTrue, tok.A)
}

func TestReaderCountTag(t *testing.T) {
	// count(3) before a list
	in := []byte{tagCount, 0x03, listStart, boolTrue, boolFalse, nilValue, listEnd}
	r := NewByteReader(in)

	tok, err := r.Next()
	assert.Nil(t, err)
	assert.Equal(t, TokenCount, tok.A)
	assert.Equal(t, uint64(3), tok.Data)

	tok, err = r.Next()
	assert.Nil(t, err)
	assert.Equal(t, TokenListStart, tok.A)
}

func TestDeterministicMapOrdering(t *testing.T) {
	enc := Encoder{Deterministic: true}

	// string keys must be sorted alphabetically
	m := map[string]int{"c": 3, "a": 1, "b": 2}
	var buf bytes.Buffer
	assert.Nil(t, enc.Write(&buf, m))

	r := NewByteReader(buf.Bytes())
	var keys []string
	tok, _ := r.Next() // dictStart
	assert.Equal(t, TokenDictStart, tok.A)
	for {
		tok, err := r.Next()
		assert.Nil(t, err)
		if tok.A == TokenDictEnd {
			break
		}
		assert.Equal(t, TokenString, tok.A)
		keys = append(keys, tok.Data.(string))
		r.Next() // value
	}
	assert.Equal(t, []string{"a", "b", "c"}, keys)
}

func TestDeterministicDisablesLRU(t *testing.T) {
	enc := Encoder{LRU: true, Deterministic: true}
	var buf bytes.Buffer
	// write same string twice — must NOT emit 0x8C or 0x81 references
	type S struct {
		A string
		B string
	}
	assert.Nil(t, enc.Write(&buf, S{A: "hello", B: "hello"}))
	b := buf.Bytes()
	for _, by := range b {
		assert.NotEqual(t, tagRefString, by, "0x8C must not appear in deterministic mode")
		assert.NotEqual(t, stringRef, by, "0x81 must not appear in deterministic mode")
	}
}

func TestDecoder(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var buf bytes.Buffer
		var enc Encoder
		enc.Write(&buf, nil)
		d := NewDecoder(buf.Bytes())
		v, err := d.Decode()
		assert.Nil(t, err)
		assert.Nil(t, v)
	})

	t.Run("bool", func(t *testing.T) {
		var buf bytes.Buffer
		var enc Encoder
		enc.Write(&buf, true)
		d := NewDecoder(buf.Bytes())
		v, err := d.Decode()
		assert.Nil(t, err)
		assert.Equal(t, true, v)
	})

	t.Run("int", func(t *testing.T) {
		var buf bytes.Buffer
		var enc Encoder
		enc.Write(&buf, 42)
		d := NewDecoder(buf.Bytes())
		v, err := d.Decode()
		assert.Nil(t, err)
		assert.Equal(t, 42, v)
	})

	t.Run("string", func(t *testing.T) {
		var buf bytes.Buffer
		var enc Encoder
		enc.Write(&buf, "hello")
		d := NewDecoder(buf.Bytes())
		v, err := d.Decode()
		assert.Nil(t, err)
		assert.Equal(t, "hello", v)
	})

	t.Run("float", func(t *testing.T) {
		var buf bytes.Buffer
		var enc Encoder
		enc.Write(&buf, math.Pi)
		d := NewDecoder(buf.Bytes())
		v, err := d.Decode()
		assert.Nil(t, err)
		assert.InDelta(t, math.Pi, v.(float64), 1e-10)
	})

	t.Run("list", func(t *testing.T) {
		var buf bytes.Buffer
		var enc Encoder
		// use []interface{} to avoid TypedArray path
		enc.Write(&buf, []interface{}{"a", "b"})
		d := NewDecoder(buf.Bytes())
		v, err := d.Decode()
		assert.Nil(t, err)
		assert.Equal(t, []interface{}{"a", "b"}, v)
	})

	t.Run("dict", func(t *testing.T) {
		var buf bytes.Buffer
		enc := Encoder{Deterministic: true}
		enc.Write(&buf, map[string]interface{}{"x": 1})
		d := NewDecoder(buf.Bytes())
		v, err := d.Decode()
		assert.Nil(t, err)
		assert.Equal(t, map[string]interface{}{"x": 1}, v)
	})

	t.Run("chaining", func(t *testing.T) {
		var buf bytes.Buffer
		var enc Encoder
		enc.Write(&buf, 1)
		enc.Write(&buf, 2)
		enc.Write(&buf, 3)
		d := NewDecoder(buf.Bytes())
		for _, expected := range []int{1, 2, 3} {
			v, err := d.Decode()
			assert.Nil(t, err)
			assert.Equal(t, expected, v)
		}
		_, err := d.Decode()
		assert.Equal(t, io.EOF, err)
	})

	t.Run("magic_skipped", func(t *testing.T) {
		var buf bytes.Buffer
		var enc Encoder
		enc.WriteWithMagic(&buf, "value")
		d := NewDecoder(buf.Bytes())
		v, err := d.Decode()
		assert.Nil(t, err)
		assert.Equal(t, "value", v)
	})
}

func BenchmarkWrite(b *testing.B) {
	for testCase, tt := range tests {
		var writer DummyWriter
		var encoder Encoder

		b.Run(testCase, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				if err := encoder.Write(&writer, tt.golang); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

type DummyWriter struct{}

func (d DummyWriter) Write(_ []byte) (int, error) {
	return 0, nil
}
