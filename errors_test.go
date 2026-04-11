package muon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMuonError_Error(t *testing.T) {
	tests := []struct {
		code int
		msg  string
		want string
	}{
		{ErrCodeInvalidTarget, "nil pointer", "muon error 1: nil pointer"},
		{ErrCodeTypeMismatch, "type mismatch", "muon error 2: type mismatch"},
		{ErrCodeUnexpectedToken, "bad token", "muon error 3: bad token"},
	}
	for _, tc := range tests {
		e := MuonError{Code: tc.code, Msg: tc.msg}
		assert.Equal(t, tc.want, e.Error())
	}
}

func TestMuonError_Codes(t *testing.T) {
	assert.Equal(t, 1, ErrCodeInvalidTarget)
	assert.Equal(t, 2, ErrCodeTypeMismatch)
	assert.Equal(t, 3, ErrCodeUnexpectedToken)
}

func TestMuonError_IsError(t *testing.T) {
	var err error = MuonError{Code: ErrCodeTypeMismatch, Msg: "test"}
	require.Error(t, err)
	assert.Contains(t, err.Error(), "muon error")
}
