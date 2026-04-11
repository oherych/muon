//go:build go1.18

package muon

import (
	"bytes"
	"testing"
)

func FuzzString(f *testing.F) {
	f.Add("test")
	f.Add("")
	f.Add(testString)
	f.Fuzz(func(t *testing.T, in string) { fuzzRoundTrip(t, in) })
}

func FuzzInt64(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(10))
	f.Add(int64(-1))
	f.Add(int64(1 << 32))
	f.Fuzz(func(t *testing.T, in int64) { fuzzRoundTrip(t, in) })
}

func FuzzBool(f *testing.F) {
	f.Add(true)
	f.Add(false)
	f.Fuzz(func(t *testing.T, in bool) { fuzzRoundTrip(t, in) })
}

func FuzzFloat64(f *testing.F) {
	f.Add(0.0)
	f.Add(1.5)
	f.Add(-1.5)
	f.Fuzz(func(t *testing.T, in float64) { fuzzRoundTrip(t, in) })
}

func fuzzRoundTrip[T any](t *testing.T, in T) {
	t.Helper()
	var buf bytes.Buffer
	var enc Encoder
	if err := enc.Write(&buf, in); err != nil {
		t.Fatal("encode:", err)
	}
	var out T
	if err := Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatal("unmarshal:", err)
	}
	// NaN != NaN by IEEE 754, skip equality check for floats containing NaN
	gotBytes, _ := marshalBytes(out)
	wantBytes, _ := marshalBytes(in)
	if !bytes.Equal(gotBytes, wantBytes) {
		t.Errorf("round-trip mismatch: in=%v out=%v", in, out)
	}
}

func marshalBytes(v any) ([]byte, error) {
	var buf bytes.Buffer
	var enc Encoder
	err := enc.Write(&buf, v)
	return buf.Bytes(), err
}
