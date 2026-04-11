package muon

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReader(t *testing.T) {
	for testCase, tt := range tests {
		t.Run(testCase, func(t *testing.T) {
			result := make([]Token, 0)
			r := NewByteReader(tt.encoded)

			for {
				token, err := r.Next()
				if err == io.EOF {
					break
				}

				result = append(result, token)
			}

			if tt.tokens != nil {
				assert.Equal(t, tt.tokens, result)
			}
		})
	}

}
