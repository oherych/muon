package muon

import "io"

// Decoder reconstructs complete Go values from a muon byte stream.
// Handles multiple concatenated objects (chaining) — call Decode in a loop
// until io.EOF is returned.
type Decoder struct {
	r Reader
}

func NewDecoder(data []byte) *Decoder {
	return &Decoder{r: NewByteReader(data)}
}

// Decode reads the next value from the stream and returns it as a Go value.
// Returns io.EOF when the stream is exhausted.
func (d *Decoder) Decode() (interface{}, error) {
	tok, err := d.r.Next()
	if err != nil {
		return nil, err
	}
	return d.tokenToValue(tok)
}

func (d *Decoder) tokenToValue(tok Token) (interface{}, error) {
	switch tok.A {
	case TokenMagic, TokenCount:
		// skip transparent tokens and read the actual value
		return d.Decode()

	case TokenString:
		return tok.Data.(string), nil

	case tokenInt:
		return tok.Data, nil

	case TokenFloat:
		return tok.Data.(float64), nil

	case tokenTrue:
		return true, nil

	case tokenFalse:
		return false, nil

	case tokenNil:
		return nil, nil

	case TokenTypedArray:
		return tok.Data, nil

	case tokenListStart:
		return d.readList()

	case tokenDictStart:
		return d.readDict()

	default:
		return nil, io.ErrUnexpectedEOF
	}
}

func (d *Decoder) readList() ([]interface{}, error) {
	var out []interface{}
	for {
		tok, err := d.r.Next()
		if err != nil {
			return nil, err
		}
		if tok.A == tokenListEnd {
			return out, nil
		}
		v, err := d.tokenToValue(tok)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
}

func (d *Decoder) readDict() (map[string]interface{}, error) {
	out := make(map[string]interface{})
	for {
		// read key
		keyTok, err := d.r.Next()
		if err != nil {
			return nil, err
		}
		if keyTok.A == tokenDictEnd {
			return out, nil
		}
		if keyTok.A != TokenString {
			return nil, io.ErrUnexpectedEOF
		}
		key := keyTok.Data.(string)

		// read value
		valTok, err := d.r.Next()
		if err != nil {
			return nil, err
		}
		val, err := d.tokenToValue(valTok)
		if err != nil {
			return nil, err
		}
		out[key] = val
	}
}
