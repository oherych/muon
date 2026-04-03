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

## Phase 2 — Typed integers & TypedArray ✅ DONE

- [x] `consts.go`: add type bytes `typeInt8..typeFloat64` (`0xB0..0xBA`)
- [x] `token.go`: add `TokenTypedArray`
- [x] `writer.go`: replace `kindToType` with `elemKindToTypeByte` (checks element kind, not slice kind)
- [x] `writer.go`: `writeList()` dispatches to `writeTypedArray()` for supported element types
- [x] `writer.go`: `writeTypedArray()` + `writeTypedElem()` — packed LE bytes
- [x] `reader.go`: handle `0x84` → `readTypedElems()` → `TokenTypedArray` with typed Go slice
- [x] `writer_test.go`: `byte_slice` updated to TypedArray encoding; added `typed_array_int32`, `typed_array_float64`

## Phase 3 — Dict completeness ✅ DONE

- [x] `writer.go`: fix key-value order in `writeMap()` (was value-first, now key-first per spec)
- [x] `writer.go`: validate all keys same kind-class (string or integer), return error if mixed
- [x] `writer.go`: `writeDictIntKey()` — first key with type prefix (`0xB0..0xB7` or `0xBB`), subsequent without
- [x] `reader.go`: add typed LE integer decoding (`0xB0..0xB7`) → tokenInt
- [x] `writer_test.go`: replace commented `map` test with `map_string`; add `map_int64_key`

## Phase 4 — Tags (magic, count, padding) ✅ DONE

- [x] `consts.go`: add `tagMagicByte = 0x8F`, `tagCount = 0x8A`, `tagPadding = 0xFF`
- [x] `token.go`: add `TokenMagic`, `TokenCount`
- [x] `writer.go`: `WriteWithMagic()` — writes `0x8F 0xB5 0x30 0x31` then value
- [x] `writer.go`: `WritePadding(w, n)` — writes n bytes of `0xFF`
- [x] `reader.go`: skip leading `0xFF` padding bytes before each token
- [x] `reader.go`: handle `0x8F` magic → `TokenMagic`; `0x8A` count tag → `TokenCount`

## Phase 5 — LRU String References + Chunked TypedArray ✅ DONE

- [x] `consts.go`: add `tagRefString = 0x8C`, `stringRef = 0x81`, `typedArrayChunk = 0x85`
- [x] `writer.go`: `Encoder` becomes stateful with pointer receivers; `LRU bool` field
- [x] `writer.go`: `writeString()` checks LRU — writes `0x8C` + string on first occurrence, `0x81` + index on repeat
- [x] `writer.go`: `WriteChunkedTypedArray(w, typeByte, chunks...)` — `0x85` + type + chunks + `0x00`
- [x] `reader.go`: `Reader` gets `lru []string`; handles `0x8C` (add to LRU), `0x81` (lookup by index)
- [x] `reader.go`: handles `0x85` chunked TypedArray — aggregates all chunks into single typed slice

## Phase 6 — Advanced

- [ ] Deterministic encoding mode (`Encoder{Deterministic: true}`)
- [ ] Streaming/Chaining decoder