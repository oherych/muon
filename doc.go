// Package muon implements encoding and decoding of the µON (muon) binary
// serialization format.
//
// µON is a compact, self-describing, schemaless format that exploits gaps in
// the UTF-8 encoding space for control bytes, making every null-terminated
// UTF-8 string a valid muon document at the same time.
//
// # Encoding
//
// Use [Encoder] to serialize Go values:
//
//	var buf bytes.Buffer
//	enc := muon.Encoder{}
//	enc.Write(&buf, map[string]interface{}{
//	    "name": "Alice",
//	    "age":  30,
//	})
//
// Structs are encoded as dicts. Field names default to strings.ToLower of the
// Go field name; use the `muon` struct tag to override or skip a field:
//
//	type User struct {
//	    Name string `muon:"name"`
//	    Pass string `muon:"-"` // skipped
//	}
//
// # Decoding
//
// [Unmarshal] decodes into a typed Go value:
//
//	var u User
//	err := muon.Unmarshal(data, &u)
//
// [NewDecoder] provides a streaming API that handles multiple concatenated
// objects (chaining):
//
//	d := muon.NewDecoder(data)
//	for {
//	    v, err := d.Decode()
//	    if err == io.EOF {
//	        break
//	    }
//	    fmt.Println(v)
//	}
//
// # Type mapping (Decoder / Unmarshal)
//
//	muon type       → Go value
//	─────────────────────────────────────────────────────
//	string          → string
//	integer         → int, int64, or uint64
//	float           → float64
//	true / false    → bool
//	null            → nil
//	TypedArray      → []int8, []float64, etc.
//	list            → []interface{}
//	dict (str keys) → map[string]interface{}
//	dict (int keys) → map[interface{}]interface{}
//
// # Options
//
// Set [Encoder.LRU] to enable string back-references, reducing the encoded
// size of documents with repeated strings. Reuse the same [Encoder] across
// multiple [Encoder.Write] calls to share the deduplication table.
//
// Set [Encoder.Deterministic] to produce canonical output — same input always
// yields identical bytes. This sorts dict keys and disables LRU string refs.
//
// # Errors
//
// [Unmarshal] and [Decoder.Unmarshal] may return [MuonError] with a Code field:
// [ErrCodeInvalidTarget], [ErrCodeTypeMismatch], or [ErrCodeUnexpectedToken].
// Streaming APIs may also return standard errors such as io.EOF.
//
// # Specification
//
// https://github.com/vshymanskyy/muon
package muon
