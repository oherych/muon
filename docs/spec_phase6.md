# Phase 6 Spec: Deterministic Encoding & Streaming Decoder

---

## Deterministic Encoding

`Encoder{Deterministic: true}` produces a canonical byte output:
same input always → same bytes.

**Rules (from muon spec):**
- LRU is disabled (0x8C/0x81 must not be used), even if `LRU: true`
- Dict keys are **sorted**: strings alphabetically, integers numerically
- Count and padding tags must not be emitted (already the case)
- Size tag only for strings (already the case)
- Integers 0–9 use inline encoding; others use SLEB128 (already the case)
- NaN/±Inf use special bytes; all other floats use float64 0xBA (already the case)

The only behavioural change from current code: **map key sorting**.

---

## Streaming Decoder

A high-level `Decoder` that reconstructs complete Go values from a byte stream.
Handles multiple concatenated objects (chaining).

```go
type Decoder struct { ... }

func NewDecoder(data []byte) *Decoder
func (d *Decoder) Decode() (interface{}, error)
```

**Type mapping (token → Go value):**

| Token | Go value |
|---|---|
| `TokenString` | `string` |
| `tokenInt` | `int` or `int64` or `uint64` (matches token Data type) |
| `TokenFloat` | `float64` |
| `tokenTrue` / `tokenFalse` | `bool` |
| `tokenNil` | `nil` |
| `TokenTypedArray` | typed slice (`[]int32`, `[]float64`, etc.) |
| `tokenListStart` | `[]interface{}` — reads until `tokenListEnd` |
| `tokenDictStart` | `map[string]interface{}` — reads key-value pairs until `tokenDictEnd` |
| `TokenMagic` | skipped, reads next value |
| `TokenCount` | skipped, reads next value |

**Chaining:** call `Decode()` in a loop; returns `io.EOF` when stream is exhausted.
