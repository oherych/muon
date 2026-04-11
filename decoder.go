package muon

import "io"

// Decoder reconstructs complete Go values from a muon byte stream.
// Handles multiple concatenated objects (chaining) — call Decode in a loop
// until io.EOF is returned.
type Decoder struct {
	r Reader
}

// NewDecoder creates a Decoder that reads from data.
func NewDecoder(data []byte) *Decoder {
	return &Decoder{r: NewByteReader(data)}
}

// Decode reads the next value from the stream and returns it as a Go value.
// Returns io.EOF when the stream is exhausted.
func (d *Decoder) Decode() (any, error) {
	tok, err := d.r.Next()
	if err != nil {
		return nil, err
	}
	return d.tokenToValue(tok)
}

func (d *Decoder) tokenToValue(tok Token) (any, error) {
	switch tok.A {
	case TokenMagic, TokenCount:
		// skip transparent tokens and read the actual value
		return d.Decode()

	case TokenString:
		return tok.Data.(string), nil

	case TokenInt:
		return tok.Data, nil

	case TokenFloat:
		return tok.Data.(float64), nil

	case TokenTrue:
		return true, nil

	case TokenFalse:
		return false, nil

	case TokenNil:
		return nil, nil

	case TokenTypedArray:
		return tok.Data, nil

	case TokenListStart:
		return d.readList()

	case TokenDictStart:
		return d.readDict()

	default:
		return nil, io.ErrUnexpectedEOF
	}
}

func (d *Decoder) readList() ([]any, error) {
	var out []any
	for {
		tok, err := d.r.Next()
		if err != nil {
			return nil, err
		}
		if tok.A == TokenListEnd {
			return out, nil
		}
		v, err := d.tokenToValue(tok)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
}

func (d *Decoder) readDict() (any, error) {
	// peek at first key to decide string vs integer dict
	keyTok, err := d.r.Next()
	if err != nil {
		return nil, err
	}
	if keyTok.A == TokenDictEnd {
		return map[string]any{}, nil
	}

	if keyTok.A == TokenString {
		return d.readStringDict(keyTok)
	}
	if keyTok.A == TokenInt {
		return d.readIntDict(keyTok)
	}
	return nil, io.ErrUnexpectedEOF
}

func (d *Decoder) readStringDict(firstKey Token) (map[string]any, error) {
	out := make(map[string]any)
	keyTok := firstKey
	for {
		key := keyTok.Data.(string)
		valTok, err := d.r.Next()
		if err != nil {
			return nil, err
		}
		val, err := d.tokenToValue(valTok)
		if err != nil {
			return nil, err
		}
		out[key] = val

		keyTok, err = d.r.Next()
		if err != nil {
			return nil, err
		}
		if keyTok.A == TokenDictEnd {
			return out, nil
		}
	}
}

func (d *Decoder) readIntDict(firstKey Token) (map[any]any, error) {
	out := make(map[any]any)
	intKeyType := d.r.lastIntKeyType
	keyTok := firstKey
	for {
		key := keyTok.Data
		valTok, err := d.r.Next()
		if err != nil {
			return nil, err
		}
		val, err := d.tokenToValue(valTok)
		if err != nil {
			return nil, err
		}
		out[key] = val

		// subsequent keys have no type prefix — use the stored type byte
		keyTok, err = d.r.NextIntKey(intKeyType)
		if err != nil {
			return nil, err
		}
		if keyTok.A == TokenDictEnd {
			return out, nil
		}
	}
}
