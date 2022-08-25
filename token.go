package muon

const (
	TokenSignature  TokenEnum = "signature"
	TokenListStart  TokenEnum = "list_start"
	TokenListEnd    TokenEnum = "list_end"
	TokenDictStart  TokenEnum = "dict_start"
	TokenDictEnd    TokenEnum = "dict_end"
	TokenTypedArray TokenEnum = "typed_array"
	TokenLiteral    TokenEnum = "literal"
)

type TokenEnum string

type Token struct {
	A TokenEnum
	D interface{}
}
