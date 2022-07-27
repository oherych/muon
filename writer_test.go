package muon

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var tests = map[string]struct {
	golang  interface{}
	encoded []byte
}{
	"string": {
		golang:  "test",
		encoded: []byte{0x74, 0x65, 0x73, 0x74, stringEnd},
	},
	"nil": {
		golang:  nil,
		encoded: []byte{nilValue},
	},
	"true": {
		golang:  true,
		encoded: []byte{boolTrue},
	},
	"false": {
		golang:  false,
		encoded: []byte{boolFalse},
	},
	"slice": {
		golang:  []interface{}{"test", true},
		encoded: []byte{listStart, 0x74, 0x65, 0x73, 0x74, stringEnd, boolTrue, listEnd},
	},
	"array": {
		golang:  [2]bool{false, true},
		encoded: []byte{listStart, boolFalse, boolTrue, listEnd},
	},
}

func TestWrite(t *testing.T) {
	for testCase, tt := range tests {
		t.Run(testCase, func(t *testing.T) {
			var writer bytes.Buffer

			err := Write(&writer, tt.golang)

			assert.Equal(t, tt.encoded, writer.Bytes())
			assert.Nil(t, err)
		})
	}
}

func TestExampleWrite(t *testing.T) {
	obj := struct {
		String string
		Int    int
		Int64  int64
	}{}

	var writer bytes.Buffer
	if err := Write(&writer, obj); err != nil {
		panic(err)
	}

	fmt.Printf("%v \n", writer.Bytes())
}

func BenchmarkWrite(b *testing.B) {
	for testCase, tt := range tests {
		var writer DummyWriter

		b.Run(testCase, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				if err := Write(&writer, tt.golang); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

type DummyWriter struct{}

func (d DummyWriter) Write(_ []byte) (int, error) {
	return 0, nil
}
