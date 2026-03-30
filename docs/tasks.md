# Implementation Tasks

Full plan: `/Users/oherych/.claude/plans/effervescent-conjuring-diffie.md`  
Spec reference: https://github.com/vshymanskyy/muon

---

## Phase 1 — Fix existing + floats ✅ DONE

- [x] Create `docs/spec_phase1.md`
- [x] `consts.go`: replace `stringStart = 0x82` → `tagSize = 0x8B`, add `floatF64 = 0xBA`
- [x] `token.go`: add `TokenFloat`
- [x] `writer.go`: fix `writeString()` — `0x8B` + ULEB128(len) + raw bytes (no null)
- [x] `writer.go`: implement `writeFloat()` — `0xBA` + 8 bytes LE IEEE 754
- [x] `writer.go`: fix `writeStruct()` — use `internal.ParseTags()`, respect skip and name
- [x] `writer.go`: fix `writeUint()` — use SLEB128 (spec: `0xBB` is always Signed LEB128)
- [x] `reader.go`: handle `0x8B` → ULEB128 + N bytes → TokenString
- [x] `reader.go`: handle `0xBA`/`0xB9` → TokenFloat; NaN/±Inf → TokenFloat with math values
- [x] `reader.go`: fix empty string (`0x00` as first byte)
- [x] `reader.go`: handle `0xBB` → Signed LEB128 → tokenInt
- [x] `writer_test.go`: update all expected bytes + add tokens for all int/uint cases
- [x] `writer_test.go`: add float64 test cases (pi, neg, zero, nan, ±inf)
- [x] `writer_test.go` / `reader_test.go`: add struct tag test cases

## Phase 2 — Typed integers & TypedArray

- [ ] `writer.go`: typed integer write (`0xB0..0xB7`, LE) for use inside TypedArray
- [ ] `writer.go`: `writeList()` for typed slices → `0x84` + type byte + ULEB128 count + packed values
- [ ] `reader.go`: handle `0x84` → tokenTypedArray
- [ ] `token.go`: add `tokenTypedArray`
- [ ] `writer_test.go`: activate commented-out `typed_array` test

## Phase 3 — Dict completeness

- [ ] `writer.go`: `writeMap()` — validate all keys same type, return error if mixed
- [ ] `writer.go`: integer key dict — first key with type prefix, subsequent without
- [ ] `writer_test.go`: activate commented-out `map` test; add `map[int]string` test

## Phase 4 — Tags (magic, count, size, padding)

- [ ] `consts.go`: add `tagMagic`, `tagCount = 0x8A`, `tagPadding = 0xFF`
- [ ] `writer.go`: `WriteWithMagic()` method
- [ ] `writer.go`: `WritePadding(w, n)` utility
- [ ] `reader.go`: skip `0xFF` padding in `Next()`
- [ ] `reader.go`: handle `0x8F` magic and `0x8A` count tag

## Phase 5 — LRU String References

- [ ] `writer.go`: `Encoder` gets `lru []string` (max 512), write `0x81` + index if found
- [ ] `reader.go`: `Reader` gets `lru []string`, handle `0x8C` and `0x81`

## Phase 6 — Advanced

- [ ] Deterministic encoding mode (`Encoder{Deterministic: true}`)
- [ ] Chunked TypedArray (`0x85`)
- [ ] Streaming/Chaining decoder