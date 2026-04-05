package muon

// TokenEnum identifies the kind of a [Token] returned by [Reader.Next].
type TokenEnum string

const (
	// TokenString is a UTF-8 string value.
	TokenString TokenEnum = "string"
	// TokenFloat is a float64 value (covers float16, float32, float64, NaN, ±Inf).
	TokenFloat TokenEnum = "float"
	// TokenTypedArray is a packed array of a single numeric type.
	// Token.Data holds a typed Go slice: []int8, []float64, etc.
	TokenTypedArray TokenEnum = "typed_array"
	// TokenMagic is the optional muon file signature (0x8F µ01).
	// Returned by [Reader.Next]; transparently skipped by [Decoder.Decode].
	TokenMagic TokenEnum = "magic"
	// TokenCount is the optional count tag (0x8A) preceding a list, dict, or string.
	// Token.Data holds the count as uint64.
	// Returned by [Reader.Next]; transparently skipped by [Decoder.Decode].
	TokenCount TokenEnum = "count"

	tokenNil       TokenEnum = "nil"
	tokenFalse     TokenEnum = "false"
	tokenTrue      TokenEnum = "true"
	tokenListStart TokenEnum = "list_start"
	tokenListEnd   TokenEnum = "list_end"
	tokenDictStart TokenEnum = "dict_start"
	tokenDictEnd   TokenEnum = "dict_end"
	tokenInt       TokenEnum = "int"
)

var (
	tokenMapping = map[byte]TokenEnum{
		listStart: tokenListStart,
		listEnd:   tokenListEnd,
		dictStart: tokenDictStart,
		dictEnd:   tokenDictEnd,
		boolFalse: tokenFalse,
		boolTrue:  tokenTrue,
		nilValue:  tokenNil,
	}
)
