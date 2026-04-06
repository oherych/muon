# muon

Go implementation of the [µON (muon)](https://github.com/vshymanskyy/muon) binary serialization format — a compact, self-describing, schemaless notation that uses gaps in the UTF-8 encoding space to encode structured data.

## Installation

```bash
go get oherych/muon
```

## Quick start

### Encoding

```go
var buf bytes.Buffer
enc := muon.Encoder{}
enc.Write(&buf, map[string]interface{}{
    "name": "Alice",
    "age":  30,
})
```

Structs are encoded as dicts. Use the `muon` tag to control field names:

```go
type User struct {
    Name string `muon:"name"`
    Pass string `muon:"-"` // skipped
}
```

### Decoding

```go
d := muon.NewDecoder(data)
for {
    v, err := d.Decode()
    if err == io.EOF {
        break
    }
    fmt.Println(v)
}
```

| muon type | Go value |
|---|---|
| string | `string` |
| integer | `int`, `int64`, or `uint64` |
| float | `float64` |
| true / false | `bool` |
| null | `nil` |
| TypedArray | `[]int8`, `[]float64`, etc. |
| list | `[]interface{}` |
| dict (string keys) | `map[string]interface{}` |
| dict (integer keys) | `map[interface{}]interface{}` |

### Low-level token reader

```go
r := muon.NewByteReader(data)
for {
    tok, err := r.Next()
    if err == io.EOF {
        break
    }
    fmt.Println(tok.A, tok.Data)
}
```

## Options

### LRU string deduplication

Repeated strings are written as back-references instead of full UTF-8 sequences. Reuse the same `Encoder` instance across writes to share the table.

```go
enc := muon.Encoder{LRU: true}
enc.Write(&buf, "status")
enc.Write(&buf, "status") // written as a 2-byte reference
```

### Deterministic encoding

Same input always produces identical bytes — dict keys are sorted, LRU is disabled.

```go
enc := muon.Encoder{Deterministic: true}
enc.Write(&buf, m)
```

### File signature

```go
enc.WriteWithMagic(&buf, value) // prepends 0x8F µ01
```

### Padding / keep-alive

```go
enc.WritePadding(&buf, 4) // writes 4 × 0xFF
```

## Custom marshaling

Implement `Marshaler` or `MarshalerStream` for custom encoding:

```go
type MyType struct{ ... }

func (m MyType) MarshalMuon() ([]byte, error) {
    // return raw muon bytes
}
```

## Specification

Full format specification: https://github.com/vshymanskyy/muon
