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

	// TokenNil is a null/nil value.
	TokenNil TokenEnum = "nil"
	// TokenFalse is a boolean false value.
	TokenFalse TokenEnum = "false"
	// TokenTrue is a boolean true value.
	TokenTrue TokenEnum = "true"
	// TokenListStart marks the beginning of a list.
	TokenListStart TokenEnum = "list_start"
	// TokenListEnd marks the end of a list.
	TokenListEnd TokenEnum = "list_end"
	// TokenDictStart marks the beginning of a dict (map or struct).
	TokenDictStart TokenEnum = "dict_start"
	// TokenDictEnd marks the end of a dict.
	TokenDictEnd TokenEnum = "dict_end"
	// TokenInt is an integer value. Token.Data holds int, int64, or uint64.
	TokenInt TokenEnum = "int"
)

var (
	tokenMapping = map[byte]TokenEnum{
		listStart: TokenListStart,
		listEnd:   TokenListEnd,
		dictStart: TokenDictStart,
		dictEnd:   TokenDictEnd,
		boolFalse: TokenFalse,
		boolTrue:  TokenTrue,
		nilValue:  TokenNil,
	}
)
