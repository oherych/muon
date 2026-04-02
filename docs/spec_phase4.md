# Phase 4 Spec: Tags (magic, count, padding)

## Magic signature — `0x8F 0xB5 0x30 0x31`

Optional 4-byte header at the start of a file or stream.  
Allows readers/tools to reliably detect muon format.

**Write:** `Encoder.WriteWithMagic(w, value)` — writes magic then encodes the value.  
**Read:** `Next()` consumes the magic bytes and returns `Token{A: TokenMagic}` — does NOT skip silently, so callers can detect it.

## Count tag — `0x8A` + ULEB128(n)

Optional hint that the following Dict/List/String contains `n` elements (not bytes).  
Allows pre-allocating buffers without scanning.

**Write:** not exposed in this phase (complex to compose with existing write methods — deferred).  
**Read:** `Next()` consumes `0x8A` + ULEB128, returns `Token{A: TokenCount, Data: uint64(n)}`.

## Padding — `0xFF`

Zero or more `0xFF` bytes. Used as keepalive or alignment.

**Write:** `Encoder.WritePadding(w, n)` — writes `n` bytes of `0xFF`.  
**Read:** `Next()` skips all consecutive `0xFF` bytes silently (no token returned — loops internally).

## New token types

| Token | Value |
|---|---|
| `TokenMagic` | `"magic"` |
| `TokenCount` | `"count"` |
