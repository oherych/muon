# Phase 2 Spec: Typed Integers & TypedArray

## Type bytes for LE-encoded values

| Go type    | Type byte | Size  |
|------------|-----------|-------|
| `int8`     | `0xB0`    | 1 byte |
| `int16`    | `0xB1`    | 2 bytes LE |
| `int32`    | `0xB2`    | 4 bytes LE |
| `int64`    | `0xB3`    | 8 bytes LE |
| `uint8`    | `0xB4`    | 1 byte |
| `uint16`   | `0xB5`    | 2 bytes LE |
| `uint32`   | `0xB6`    | 4 bytes LE |
| `uint64`   | `0xB7`    | 8 bytes LE |
| `float32`  | `0xB9`    | 4 bytes LE |
| `float64`  | `0xBA`    | 8 bytes LE |

`int` / `uint` (platform-dependent) are NOT TypedArray — they fall through to generic List.

---

## TypedArray encoding

```
0x84 | type_byte | ULEB128(count) | count × LE_bytes(element)
```

**Examples:**

| Go value | Encoded |
|---|---|
| `[]int32{10, 500}` | `84 B2 02 0A 00 00 00 F4 01 00 00` |
| `[]uint8{1, 30}` | `84 B4 02 01 1E` |
| `[]float64{1.5}` | `84 BA 01 00 00 00 00 00 00 F8 3F` |

Note: `[]byte` = `[]uint8` — encoded as TypedArray (type `0xB4`), not as generic List.

---

## TypedArray decoding

- On byte `0x84`: read type byte, read ULEB128 count, read `count × sizeof(type)` bytes
- Return `Token{A: tokenTypedArray, Data: <typed Go slice>}`
- Data type matches the element type: `[]int32`, `[]uint8`, `[]float64`, etc.
