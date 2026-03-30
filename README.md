# muon

Go implementation of the [µON (muon)](https://github.com/vshymanskyy/muon) binary serialization format — a compact, self-describing, schemaless notation that uses gaps in the UTF-8 encoding space to encode structured data.




# tokens




| Functionality    |     | Write | Read |
|------------------|-----|-------|------|
| string with 0x00 |     | Y     | X    |
| string2          |     | X     | X    |
| string3          |     | X     | X    |
| special True     |     | Y     | X    |
| special False    |     | Y     | X    |
| special Null     |     | Y     | X    |
| special NaN      |     | Y     | X    |
| special -Inf     |     | Y     | X    |
| special +Inf     |     | Y     | X    |
| special array    |     | Y     | X    |
| typed value      |     | X     | X    |
| list             |     | Y     | X    |
| dict             |     | Y     | X    |



