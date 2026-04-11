package muon

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal_Primitives(t *testing.T) {
	t.Run("bool_true", func(t *testing.T) {
		var v bool
		require.NoError(t, Unmarshal(encode(t, true), &v))
		assert.True(t, v)
	})
	t.Run("bool_false", func(t *testing.T) {
		var v bool
		require.NoError(t, Unmarshal(encode(t, false), &v))
		assert.False(t, v)
	})
	t.Run("int", func(t *testing.T) {
		var v int
		require.NoError(t, Unmarshal(encode(t, 42), &v))
		assert.Equal(t, 42, v)
	})
	t.Run("int8", func(t *testing.T) {
		var v int8
		require.NoError(t, Unmarshal(encode(t, int8(-5)), &v))
		assert.Equal(t, int8(-5), v)
	})
	t.Run("int64", func(t *testing.T) {
		var v int64
		require.NoError(t, Unmarshal(encode(t, int64(1<<40)), &v))
		assert.Equal(t, int64(1<<40), v)
	})
	t.Run("uint32", func(t *testing.T) {
		var v uint32
		require.NoError(t, Unmarshal(encode(t, uint32(1000)), &v))
		assert.Equal(t, uint32(1000), v)
	})
	t.Run("float64", func(t *testing.T) {
		var v float64
		require.NoError(t, Unmarshal(encode(t, math.Pi), &v))
		assert.InDelta(t, math.Pi, v, 1e-12)
	})
	t.Run("float32", func(t *testing.T) {
		var v float32
		require.NoError(t, Unmarshal(encode(t, float32(1.5)), &v))
		assert.InDelta(t, 1.5, float64(v), 1e-4)
	})
	t.Run("string", func(t *testing.T) {
		var v string
		require.NoError(t, Unmarshal(encode(t, "hello"), &v))
		assert.Equal(t, "hello", v)
	})
	t.Run("nil_into_interface", func(t *testing.T) {
		var v any
		require.NoError(t, Unmarshal(encode(t, nil), &v))
		assert.Nil(t, v)
	})
	t.Run("nil_zeros_pointer", func(t *testing.T) {
		s := "before"
		p := &s
		require.NoError(t, Unmarshal(encode(t, nil), &p))
		assert.Nil(t, p)
	})
}

func TestUnmarshal_Struct(t *testing.T) {
	type Person struct {
		Name string `muon:"name"`
		Age  int    `muon:"age"`
		Skip string `muon:"-"`
	}

	enc := &Encoder{Deterministic: true}
	data := encodeWith(t, enc, Person{Name: "Alice", Age: 30, Skip: "ignored"})

	var out Person
	require.NoError(t, Unmarshal(data, &out))
	assert.Equal(t, "Alice", out.Name)
	assert.Equal(t, 30, out.Age)
	assert.Equal(t, "", out.Skip)
}

func TestUnmarshal_Struct_DefaultFieldName(t *testing.T) {
	type S struct {
		Hello string
	}
	data := encode(t, S{Hello: "world"})
	var out S
	require.NoError(t, Unmarshal(data, &out))
	assert.Equal(t, "world", out.Hello)
}

func TestUnmarshal_Struct_DefaultFieldNameAcronym(t *testing.T) {
	type S struct {
		URL        string
		HTTPStatus int
	}
	data := encode(t, S{URL: "https://example.com", HTTPStatus: 200})
	var out S
	require.NoError(t, Unmarshal(data, &out))
	assert.Equal(t, "https://example.com", out.URL)
	assert.Equal(t, 200, out.HTTPStatus)
}

func TestUnmarshal_Struct_UnknownFieldSkipped(t *testing.T) {
	// encode struct with extra field, decode into smaller struct
	type Full struct {
		Name  string `muon:"name"`
		Extra string `muon:"extra"`
	}
	type Partial struct {
		Name string `muon:"name"`
	}
	enc := &Encoder{Deterministic: true}
	data := encodeWith(t, enc, Full{Name: "Alice", Extra: "ignored"})
	var out Partial
	require.NoError(t, Unmarshal(data, &out))
	assert.Equal(t, "Alice", out.Name)
}

func TestUnmarshal_Map(t *testing.T) {
	enc := &Encoder{Deterministic: true}
	data := encodeWith(t, enc, map[string]int{"a": 1, "b": 2})
	var out map[string]int
	require.NoError(t, Unmarshal(data, &out))
	assert.Equal(t, map[string]int{"a": 1, "b": 2}, out)
}

func TestUnmarshal_Slice(t *testing.T) {
	data := encode(t, []any{"x", "y", "z"})
	var out []string
	require.NoError(t, Unmarshal(data, &out))
	assert.Equal(t, []string{"x", "y", "z"}, out)
}

func TestUnmarshal_TypedArray(t *testing.T) {
	data := encode(t, []int32{1, 2, 3})
	var out []int32
	require.NoError(t, Unmarshal(data, &out))
	assert.Equal(t, []int32{1, 2, 3}, out)
}

func TestUnmarshal_Array(t *testing.T) {
	data := encode(t, []int32{10, 20, 30})
	var out [3]int32
	require.NoError(t, Unmarshal(data, &out))
	assert.Equal(t, [3]int32{10, 20, 30}, out)
}

func TestUnmarshal_Pointer(t *testing.T) {
	data := encode(t, "pointed")
	var s string
	p := &s
	require.NoError(t, Unmarshal(data, &p))
	assert.Equal(t, "pointed", *p)
}

func TestUnmarshal_Interface(t *testing.T) {
	data := encode(t, []any{"a", 1, true})
	var out any
	require.NoError(t, Unmarshal(data, &out))
	assert.Equal(t, []any{"a", 1, true}, out)
}

func TestUnmarshal_MagicTransparent(t *testing.T) {
	var buf []byte
	{
		var b [4]byte
		b[0] = tagMagicByte
		b[1] = 0xB5
		b[2] = 0x30
		b[3] = 0x31
		buf = append(b[:], encode(t, "hi")...)
	}
	var out string
	require.NoError(t, Unmarshal(buf, &out))
	assert.Equal(t, "hi", out)
}

func TestUnmarshal_InvalidTarget(t *testing.T) {
	err := Unmarshal(encode(t, 1), nil)
	assert.Error(t, err)
	me, ok := err.(MuonError)
	require.True(t, ok)
	assert.Equal(t, ErrCodeInvalidTarget, me.Code)
}

func TestUnmarshal_TypeMismatch(t *testing.T) {
	var v int
	err := Unmarshal(encode(t, "string"), &v)
	assert.Error(t, err)
	me, ok := err.(MuonError)
	require.True(t, ok)
	assert.Equal(t, ErrCodeTypeMismatch, me.Code)
}

func TestUnmarshal_MapIntKeys(t *testing.T) {
	enc := &Encoder{Deterministic: true}
	data := encodeWith(t, enc, map[int64]string{1: "a", 2: "b"})
	var out map[int64]string
	require.NoError(t, Unmarshal(data, &out))
	assert.Equal(t, map[int64]string{1: "a", 2: "b"}, out)
}

func TestUnmarshal_SkipNestedValue(t *testing.T) {
	// Struct with extra fields whose values are nested list and nested dict.
	// skipValue must correctly consume the entire nested structure.
	type Full struct {
		Name string   `muon:"name"`
		Tags []string `muon:"tags"`
		Meta struct {
			K string `muon:"k"`
		} `muon:"meta"`
	}
	type Partial struct {
		Name string `muon:"name"`
	}
	enc := &Encoder{Deterministic: true}
	data := encodeWith(t, enc, Full{
		Name: "test",
		Tags: []string{"a", "b"},
		Meta: struct {
			K string `muon:"k"`
		}{K: "v"},
	})
	var out Partial
	require.NoError(t, Unmarshal(data, &out))
	assert.Equal(t, "test", out.Name)
}

func TestUnmarshal_EmptySlice(t *testing.T) {
	var out []string
	require.NoError(t, Unmarshal(encode(t, []any{}), &out))
	assert.Empty(t, out)
}

func TestUnmarshal_NonPointerTarget(t *testing.T) {
	var v int
	err := Unmarshal(encode(t, 1), v) // passing non-pointer
	require.Error(t, err)
	me, ok := err.(MuonError)
	require.True(t, ok)
	assert.Equal(t, ErrCodeInvalidTarget, me.Code)
}

func TestUnmarshal_DecoderUnmarshal(t *testing.T) {
	// Decoder.Unmarshal used directly (vs package-level Unmarshal shortcut).
	data := encode(t, "direct")
	d := NewDecoder(data)
	var out string
	require.NoError(t, d.Unmarshal(&out))
	assert.Equal(t, "direct", out)
}

func TestUnmarshal_RoundTrip(t *testing.T) {
	type Item struct {
		ID     int64   `muon:"id"`
		Name   string  `muon:"name"`
		Score  float64 `muon:"score"`
		Active bool    `muon:"active"`
	}
	in := Item{ID: 42, Name: "hello", Score: 3.14, Active: true}
	enc := &Encoder{Deterministic: true}
	data := encodeWith(t, enc, in)

	var out Item
	require.NoError(t, Unmarshal(data, &out))
	assert.Equal(t, in.ID, out.ID)
	assert.Equal(t, in.Name, out.Name)
	assert.InDelta(t, in.Score, out.Score, 1e-10)
	assert.Equal(t, in.Active, out.Active)
}
