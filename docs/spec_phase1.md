# Phase 1 Spec: Foundation Correctness

## 1. Fixed-Length String (bug fix)

**Trigger:** string length >= 512 bytes OR string contains `0x00` byte.

**Encoding:**
```
0x8B | ULEB128(len_bytes) | utf8_bytes
```
- `0x8B` = Size tag
- `len_bytes` = byte length of the string (not character count)
- `utf8_bytes` = raw UTF-8, NOT null-terminated

**Decoding:**
- On byte `0x8B`: read ULEB128 → N, read next N bytes → Token{A: TokenString, Data: string(bytes)}

**Examples:**

| Input | Encoded bytes |
|---|---|
| `"te\x00st"` (5 bytes) | `8B 05 74 65 00 73 74` |
| 512-byte string | `8B 80 04 <512 bytes>` |

> **Previous behavior (bug):** used `0x82` prefix with no length — not in spec.

---

## 2. Short String (unchanged)

**Trigger:** length < 512 bytes AND no `0x00` bytes.

**Encoding:**
```
utf8_bytes | 0x00
```

**Decoding:** read bytes until `0x00` → Token{A: TokenString, Data: string(bytes_before_zero)}

---

## 3. Float64

**Trigger:** `float64` or `float32` Go value that is NOT NaN, -Inf, or +Inf.

**Encoding:**
```
0xBA | 8_bytes_little_endian_ieee754
```

**Decoding:**
- On byte `0xBA`: read 8 bytes LE → interpret as float64 → Token{A: tokenFloat, Data: float64}
- `0xB9` (float32): read 4 bytes LE → Token{A: tokenFloat, Data: float64(value)}
- `0xB8` (float16): not in Phase 1 (no Go native type)

**Special values (already implemented, unchanged):**
- NaN → `0xAD`
- -Inf → `0xAE`
- +Inf → `0xAF`

**Examples:**

| Input | Encoded bytes |
|---|---|
| `3.14` (float64) | `BA 1F 85 EB 51 B8 09 40` |
| `-1.0` (float64) | `BA 00 00 00 00 00 00 F0 BF` |
| `0.0` (float64) | `BA 00 00 00 00 00 00 00 00` |
| NaN | `AD` |
| +Inf | `AF` |
| -Inf | `AE` |

---

## 4. Struct Tags

**Tag format:** `` `muon:"name"` `` or `` `muon:"-"` ``

**Rules:**
- If tag is `"-"` → skip field entirely
- If tag name is set → use that name as dict key
- If no tag → use lowercase field name (current behavior uses original name — also a bug)
- Unexported fields → skip

**Examples:**

```go
type Example struct {
    Name    string `muon:"name"`   // key: "name"
    Age     int    `muon:"age"`    // key: "age"
    Secret  string `muon:"-"`      // skipped
    NoTag   bool                   // key: "notag" (lowercase)
}
```

Encoded as Dict with keys `"name"`, `"age"`, `"notag"` (Secret is absent).
