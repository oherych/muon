package muon

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	for testCase, tt := range tests {
		t.Run(testCase, func(t *testing.T) {
			var writer bytes.Buffer

			encoder := NewEncoder(&writer)
			encoder.WithSignature = tt.withSignature

			err := encoder.Marshal(tt.golang)

			assert.Equal(t, tt.encoded, writer.Bytes())
			assert.Nil(t, err)
		})
	}
}

func BenchmarkEncoder(b *testing.B) {
	for testCase, tt := range tests {
		var writer DummyWriter

		encoder := NewEncoder(&writer)
		encoder.WithSignature = tt.withSignature

		b.Run(testCase, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				if err := encoder.Marshal(tt.golang); err != nil {
					b.Fatal(err)
				}
			}

		})
	}
}

type DummyWriter struct{}

func (d DummyWriter) Write(in []byte) (int, error) {
	return len(in), nil
}
