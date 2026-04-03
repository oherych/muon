# Phase 5 Spec: LRU String References & Chunked TypedArray

---

## LRU String References

### Encoding side

`Encoder` gets an opt-in `LRU bool` field. When `LRU: false` (default), behaviour is unchanged.

When `LRU: true`:
- Maintain an ordered list of up to 512 strings (index 0 = most recently added)
- When writing a string **not** in LRU: write `0x8C` tag + null-terminated/size-tagged string, prepend to LRU
- When writing a string **already** in LRU: write `0x81` + ULEB128(index), do **not** re-add to LRU
- Never add the same string twice
- When LRU reaches 512 entries, drop the oldest (last) entry before prepending a new one

### Byte formats

```
0x8C | <regular string encoding>    — "remember this string"
0x81 | ULEB128(index)               — "use string at index N from LRU"
```

### Decoding side

`Reader` gets `lru []string` (zero-value = LRU disabled, works transparently).

- On `0x8C`: consume the tag, read the following string normally, prepend to `lru`, return `Token{A: TokenString, Data: s}`
- On `0x81`: read ULEB128 index, return `Token{A: TokenString, Data: lru[index]}`

---

## Chunked TypedArray

### Format

```
0x85 | type_byte | ULEB128(n₁) | n₁×LE_bytes | ULEB128(n₂) | n₂×LE_bytes | ... | ULEB128(0)
```

Zero-length chunk (`ULEB128(0)`) terminates the sequence.

### Write API

```go
func (e Encoder) WriteChunkedTypedArray(w io.Writer, typeByte byte, chunks ...interface{}) error
```

Each element of `chunks` must be a slice of the type matching `typeByte`. Writes each slice as one chunk, then writes terminating `ULEB128(0)`.

### Read

On `0x85`: read type byte, then loop reading ULEB128 count + values until count == 0. Concatenate all chunks. Return single `Token{A: TokenTypedArray, Data: <combined slice>}`.
