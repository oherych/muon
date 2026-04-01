# Phase 3 Spec: Dict Completeness

## Dict format

```
0x92 | key₁ | value₁ | key₂ | value₂ | ... | 0x93
```

**Key order within each pair:** key first, then value.

> Current code has a bug: writes value before key — must be fixed.

## Constraints (validated at encode time)

1. All keys must be the same type: all strings OR all integers (mixing → error)
2. Duplicate keys are not allowed (not validated in this phase — deferred to Phase 6)
3. `special` integer encoding (`0xA0..0xA9`) must **not** be used for keys

## String-keyed dict

Keys encoded as regular strings (null-terminated or size-tagged). No special rules.

**Example:** `map[string]string{"a": "b"}`
```
92 | 61 00 | 62 00 | 93
   ^key "a" ^val "b"
```

## Integer-keyed dict

- **First key:** full typed integer encoding — `0xBB` + SLEB128 (for `int`/`int64`) or `0xB0..0xB7` + LE bytes (for sized types)
- **Subsequent keys:** same byte width, but **type specifier byte is omitted**
- `0xA0..0xA9` (inline 0-9) must **not** be used for keys

**Example:** `map[int64]string{10: "a", 20: "b"}` (assume iteration order 10, 20)
```
92 | BB 0A | 61 00 | 14 | 62 00 | 93
   ^first key: 0xBB + SLEB128(10)
               ^val "a"
                     ^second key: SLEB128(20) only, no 0xBB prefix
                        ^val "b"
```

## Reader

Dict tokens are already handled by the generic tokenMapping (`0x92`/`0x93`). No reader changes needed in this phase — keys are decoded as regular tokens (string or int).

> Note: integer keys after the first one will be decoded as strings by the current reader (the bytes `0x0A` etc. look like string bytes). Full integer-key dict decoding requires tracking dict context — deferred to a later phase.
