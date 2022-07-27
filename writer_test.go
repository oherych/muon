package muon

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testString = "test"

	tests = map[string]struct {
		golang  interface{}
		encoded []byte
	}{
		"nil": {golang: nil, encoded: []byte{nilValue}},

		"string": {golang: testString, encoded: []byte{0x74, 0x65, 0x73, 0x74, stringEnd}},

		"pointer": {golang: &testString, encoded: []byte{0x74, 0x65, 0x73, 0x74, stringEnd}},

		"true":  {golang: true, encoded: []byte{boolTrue}},
		"false": {golang: false, encoded: []byte{boolFalse}},

		"int":   {golang: 5, encoded: []byte{0xa5}},
		"int8":  {golang: int8(8), encoded: []byte{0xa8}},
		"int16": {golang: int16(16), encoded: []byte{0x10}},
		"int32": {golang: int32(32), encoded: []byte{0x20}},
		"int64": {golang: int64(64), encoded: []byte{0xc0, 0x0}},

		"uint":   {golang: uint(5), encoded: []byte{0xa5}},
		"uint8":  {golang: uint8(8), encoded: []byte{0xa8}},
		"uint16": {golang: uint16(16), encoded: []byte{0x10}},
		"uint32": {golang: uint32(32), encoded: []byte{0x20}},
		"uint64": {golang: uint64(64), encoded: []byte{0x40}},

		"slice":      {golang: []interface{}{testString, true}, encoded: []byte{listStart, 0x74, 0x65, 0x73, 0x74, stringEnd, boolTrue, listEnd}},
		"byte_slice": {golang: []byte{1, 30}, encoded: []byte{listStart, 0xa1, 0x1e, listEnd}},

		"array": {golang: [2]bool{false, true}, encoded: []byte{listStart, boolFalse, boolTrue, listEnd}},

		"map": {golang: map[string]string{"a": "b"}, encoded: []byte{dictStart, 0x62, stringEnd, 0x61, stringEnd, dictEnd}},
	}
)

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
