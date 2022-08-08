package muon

import (
	"bufio"
	"github.com/go-interpreter/wagon/wasm/leb128"
	"io"
)

type Reader struct {
	b *bufio.Reader
}

func NewReader(r io.Reader) Reader {
	return Reader{
		b: bufio.NewReader(r),
	}
}

func (r *Reader) Next() (Token, error) {
	first, err := r.b.ReadByte()
	if err != nil {
		return Token{}, err
	}

	if token, ok := tokenMapping[first]; ok {
		return Token{A: token}, nil
	}

	if first >= 0xA0 && first <= 0xA0+9 {
		return Token{A: TokenNumber, Data: int(first - 0xA0)}, nil
	}

	if first == stringEnd {
		return Token{A: TokenString, Data: ""}, nil
	}

	if first == stringStart {
		size, err := r.readCount()
		if err != nil {
			return Token{}, err
		}

		buf := make([]byte, size)
		if _, err := r.b.Read(buf); err != nil {
			return Token{}, err
		}

		return Token{A: TokenString, Data: string(buf)}, nil
	}

	if err := r.b.UnreadByte(); err != nil {
		return Token{}, err
	}

	str, err := r.b.ReadString(stringEnd)
	if err != nil {
		return Token{}, err
	}

	return Token{A: TokenString, Data: str[:len(str)-1]}, nil
}

func (r Reader) readCount() (int, error) {
	v, err := leb128.ReadVarint64(r.b)

	return int(v), err
}
