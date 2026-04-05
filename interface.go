package muon

import "io"

// Marshaler is implemented by types that can encode themselves to muon bytes.
// MarshalMuon must return a complete, valid muon-encoded value.
type Marshaler interface {
	MarshalMuon() ([]byte, error)
}

// MarshalerStream is implemented by types that write their muon encoding
// directly to a writer. It is preferred over [Marshaler] for large values to
// avoid allocating an intermediate byte slice.
type MarshalerStream interface {
	MarshalMuon(w io.Writer) error
}
