package muon

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"reflect"
	"testing"
)

func TestDecoder_Unmarshal(t *testing.T) {
	for testCase, tt := range tests {
		t.Run(testCase, func(t *testing.T) {
			if tt.skipReading {
				t.SkipNow()
			}

			for targetName, in := range tt.unmarshal {
				t.Run(targetName, func(t *testing.T) {
					dr := bytes.NewReader(tt.encoded)
					decoder := NewDecoder(dr)

					err := decoder.Unmarshal(in.ptr)
					if err != nil {
						t.Fatal(err)
					}

					assert.Equal(t, in.exp, reflect.ValueOf(in.ptr).Elem().Interface())
				})
			}
		})
	}
}

func TestDecoder_Next(t *testing.T) {
	for testCase, tt := range tests {
		t.Run(testCase, func(t *testing.T) {
			if tt.skipReading {
				t.SkipNow()
			}

			dr := bytes.NewReader(tt.encoded)
			decoder := NewDecoder(dr)

			result := make([]Token, 0)
			for {
				token, err := decoder.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}

				result = append(result, token)
			}

			assert.Equal(t, tt.tokens, result)
		})
	}
}

func BenchmarkDecoder(b *testing.B) {
	for testCase, tt := range tests {
		r := bytes.NewReader(tt.encoded)
		decoder := NewDecoder(r)

		b.Run(testCase, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				for {
					_, err := decoder.Next()
					if err == io.EOF {
						break
					}
					if err != nil {
						b.Fatal(err)
					}
				}
			}

		})
	}
}
