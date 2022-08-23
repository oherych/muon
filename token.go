package muon

import "math"

const (
	TokenSignature  TokenEnum = "signature"
	tokenListStart  TokenEnum = "list_start"
	tokenListEnd    TokenEnum = "list_end"
	tokenDictStart  TokenEnum = "dict_start"
	tokenDictEnd    TokenEnum = "dict_end"
	TokenTypedArray TokenEnum = "typed_array"
	TokenLiteral    TokenEnum = "literal"
)

var (
	tokenMapping = map[byte]Token{
		listStart:        {A: tokenListStart},
		listEnd:          {A: tokenListEnd},
		dictStart:        {A: tokenDictStart},
		dictEnd:          {A: tokenDictEnd},
		boolFalse:        {A: TokenLiteral, D: false},
		boolTrue:         {A: TokenLiteral, D: true},
		nilValue:         {A: TokenLiteral, D: nil},
		nanValue:         {A: TokenLiteral, D: math.NaN()},
		negativeInfValue: {A: TokenLiteral, D: math.Inf(-1)},
		positiveInfValue: {A: TokenLiteral, D: math.Inf(1)},
	}
)

type TokenEnum string

type Token struct {
	A TokenEnum
	D interface{}
}
