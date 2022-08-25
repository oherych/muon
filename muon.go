package muon

import "io"

type Marshaler interface {
	MarshalMuon() ([]byte, error)
}

type MarshalerStream interface {
	MarshalMuon(w io.Writer) error
}
