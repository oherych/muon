//go:build go1.18

package muon

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func FuzzString(f *testing.F) {
	f.Add("test")
	f.Add("")
	f.Add(testStringWithZero)
	f.Add(testLongString)
	f.Fuzz(func(t *testing.T, in string) { fuzzNew(t, in) })
}

func FuzzInt64(f *testing.F) {
	f.Add(int64(10))
	f.Fuzz(func(t *testing.T, in int64) { fuzzNew(t, in) })
}

func FuzzBool(f *testing.F) {
	f.Add(true)
	f.Fuzz(func(t *testing.T, in bool) { fuzzNew(t, in) })
}

func fuzzNew[T any](t *testing.T, in T) {
	var writer bytes.Buffer
	encoder := NewEncoder(&writer, Config{})
	decoder := NewDecoder(&writer)

	if err := encoder.Write(in); err != nil {
		t.Fatal(err)
	}

	var target T
	if err := decoder.Unmarshal(&target); err != nil {
		t.Error(err)
	}

	assert.Equal(t, in, target)
}
