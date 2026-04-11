package muon

import "fmt"

// Error codes returned in [MuonError.Code].
const (
	// ErrCodeInvalidTarget is returned when Unmarshal receives a nil or non-pointer target.
	ErrCodeInvalidTarget = iota + 1
	// ErrCodeTypeMismatch is returned when the muon token cannot be assigned to the target type.
	ErrCodeTypeMismatch
	// ErrCodeUnexpectedToken is returned when an unexpected token is encountered during decoding.
	ErrCodeUnexpectedToken
)

// MuonError is a structured error returned by Unmarshal and other typed decode
// paths when the target is invalid, a token cannot be assigned to the target
// type, or an unexpected token is encountered.
//
// Other APIs may still return standard errors such as io.EOF or fmt errors.
type MuonError struct {
	Code int
	Msg  string
}

func (e MuonError) Error() string {
	return fmt.Sprintf("muon error %d: %s", e.Code, e.Msg)
}

func errInvalidTarget(msg string) error {
	return MuonError{Code: ErrCodeInvalidTarget, Msg: msg}
}

func errTypeMismatch(token TokenEnum, target interface{}) error {
	return MuonError{Code: ErrCodeTypeMismatch, Msg: fmt.Sprintf("cannot assign %s token to %T", token, target)}
}

func errUnexpectedToken(token TokenEnum) error {
	return MuonError{Code: ErrCodeUnexpectedToken, Msg: fmt.Sprintf("unexpected token: %s", token)}
}
