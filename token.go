package muon

import "math"

const (
	TokenSignature  TokenEnum = "signature"
	TokenString     TokenEnum = "string"
	tokenNil        TokenEnum = "nil"
	tokenFalse      TokenEnum = "false"
	tokenTrue       TokenEnum = "true"
	tokenListStart  TokenEnum = "list_start"
	tokenListEnd    TokenEnum = "list_end"
	tokenDictStart  TokenEnum = "dict_start"
	tokenDictEnd    TokenEnum = "dict_end"
	TokenNumber     TokenEnum = "number"
	TokenTypedArray TokenEnum = "typed_array"
)

var (
	tokenMapping = map[byte]Token{
		listStart:        {A: tokenListStart},
		listEnd:          {A: tokenListEnd},
		dictStart:        {A: tokenDictStart},
		dictEnd:          {A: tokenDictEnd},
		boolFalse:        {A: tokenFalse},
		boolTrue:         {A: tokenTrue},
		nilValue:         {A: tokenNil},
		nanValue:         {A: TokenNumber, D: math.NaN()},
		negativeInfValue: {A: TokenNumber, D: math.Inf(-1)},
		positiveInfValue: {A: TokenNumber, D: math.Inf(1)},
	}
)

type TokenEnum string

type Token struct {
	A TokenEnum
	D interface{}
}
