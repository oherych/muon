package muon

const (
	TokenString    TokenEnum = "string"
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

type TokenEnum string
