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
// Structs are encoded as dicts. Field names default to the lowercase field
// name; use the `muon` struct tag to override:
//
//	type User struct {
//	    Name string `muon:"name"`
//	    Pass string `muon:"-"` // skipped
//	}
//
// # Decoding
//
// [NewDecoder] provides a high-level API that reconstructs Go values:
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
// For lower-level control, use [NewByteReader] and call [Reader.Next] to
// iterate over tokens one at a time.
//
// # Type mapping (Decoder)
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
// # Deterministic encoding
//
// Set [Encoder.Deterministic] to produce canonical output — same input always
// yields identical bytes. This sorts dict keys and disables LRU string refs.
//
// # Specification
//
// https://github.com/vshymanskyy/muon
package muon
