package muon

import (
	"io"
)

type Reader struct {
	in    []byte
	scanp int
}

type Token struct {
	A    TokenEnum
	Data interface{}
}

func NewByteReader(in []byte) Reader {
	return Reader{in: in}
}

func (r *Reader) Next() (Token, error) {
	if r.scanp >= len(r.in) {
		return Token{}, io.EOF
	}

	first := r.in[r.scanp]

	r.scanp += 1

	if token, ok := tokenMapping[first]; ok {
		return Token{A: token}, nil
	}

	if first >= 0xA0 && first <= 0xA0+9 {
		return Token{A: tokenInt, Data: int(first - 0xA0)}, nil
	}

	// strings
	from := r.scanp - 1
	to := r.scanp
	for ; to < len(r.in); to++ {
		if r.in[to] == stringEnd {
			r.scanp = to + 1

			return Token{A: TokenString, Data: string(r.in[from:to])}, nil
		}
	}

	return Token{}, io.EOF
}
