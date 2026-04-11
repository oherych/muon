package muon_test

import (
	"bytes"
	"fmt"
	"io"

	"github.com/oherych/muon"
)

// Encode a struct and decode it back with Decoder.
func Example() {
	type Point struct {
		X float64 `muon:"x"`
		Y float64 `muon:"y"`
	}

	var buf bytes.Buffer
	enc := muon.Encoder{}
	enc.Write(&buf, Point{X: 1.5, Y: 2.5})

	d := muon.NewDecoder(buf.Bytes())
	v, _ := d.Decode()
	m := v.(map[string]interface{})
	fmt.Printf("x=%.1f y=%.1f\n", m["x"], m["y"])
	// Output:
	// x=1.5 y=2.5
}

// Write and read back a string value.
func ExampleEncoder_Write() {
	var buf bytes.Buffer
	enc := muon.Encoder{}
	enc.Write(&buf, "hello")

	d := muon.NewDecoder(buf.Bytes())
	v, _ := d.Decode()
	fmt.Println(v)
	// Output:
	// hello
}

// WriteWithMagic prepends the muon file signature so readers can detect the format.
func ExampleEncoder_WriteWithMagic() {
	var buf bytes.Buffer
	enc := muon.Encoder{}
	enc.WriteWithMagic(&buf, 42)

	d := muon.NewDecoder(buf.Bytes())
	v, _ := d.Decode() // magic token is skipped automatically
	fmt.Println(v)
	// Output:
	// 42
}

// Decode multiple concatenated muon objects from a single stream.
func ExampleDecoder_Decode_chaining() {
	var buf bytes.Buffer
	enc := muon.Encoder{}
	enc.Write(&buf, "first")
	enc.Write(&buf, "second")
	enc.Write(&buf, "third")

	d := muon.NewDecoder(buf.Bytes())
	for {
		v, err := d.Decode()
		if err == io.EOF {
			break
		}
		fmt.Println(v)
	}
	// Output:
	// first
	// second
	// third
}

// LRU string deduplication reduces repeated strings to a back-reference.
func ExampleEncoder_lru() {
	enc := muon.Encoder{LRU: true}

	var buf bytes.Buffer
	// Reuse the same encoder so the LRU table is shared across writes.
	enc.Write(&buf, "status")
	enc.Write(&buf, "status") // written as 0x81 index, not full string
	enc.Write(&buf, "status")

	fmt.Println("encoded bytes:", len(buf.Bytes()))

	d := muon.NewDecoder(buf.Bytes())
	for {
		v, err := d.Decode()
		if err == io.EOF {
			break
		}
		fmt.Println(v)
	}
	// Output:
	// encoded bytes: 12
	// status
	// status
	// status
}

// Deterministic encoding sorts dict keys so the same input always produces
// the same bytes.
func ExampleEncoder_Deterministic() {
	enc := muon.Encoder{Deterministic: true}

	var b1, b2 bytes.Buffer
	m := map[string]int{"c": 3, "a": 1, "b": 2}
	enc.Write(&b1, m)

	enc2 := muon.Encoder{Deterministic: true}
	enc2.Write(&b2, m)

	fmt.Println("equal:", bytes.Equal(b1.Bytes(), b2.Bytes()))
	// Output:
	// equal: true
}

// Use Reader for low-level token-by-token processing.
func ExampleReader_Next() {
	var buf bytes.Buffer
	enc := muon.Encoder{}
	enc.Write(&buf, []interface{}{"a", "b"})

	r := muon.NewByteReader(buf.Bytes())
	for {
		tok, err := r.Next()
		if err == io.EOF {
			break
		}
		fmt.Println(tok.A, tok.Data)
	}
	// Output:
	// list_start <nil>
	// string a
	// string b
	// list_end <nil>
}
