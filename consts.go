package muon

const (
	// string
	stringEnd = 0x00

	// list
	listStart = 0x90
	listEnd   = 0x91

	// dict
	dictStart = 0x92
	dictEnd   = 0x93

	// special values
	boolFalse        = 0xAA
	boolTrue         = 0xAB
	nilValue         = 0xAC
	nanValue         = 0xAD
	negativeInfValue = 0xAE // -Inf
	positiveInfValue = 0xAF // +Inf

	tagSize         = 0x8B
	tagCount        = 0x8A
	tagPadding      = 0xFF
	tagRefString    = 0x8C
	stringRef       = 0x81
	typedArray      = 0x84
	typedArrayChunk = 0x85
	floatF64        = 0xBA

	// magic signature bytes (after the 0x8F tag byte)
	tagMagicByte = 0x8F

	// typed integer/float type bytes (used in TypedArray)
	typeInt8    = 0xB0
	typeInt16   = 0xB1
	typeInt32   = 0xB2
	typeInt64   = 0xB3
	typeUint8   = 0xB4
	typeUint16  = 0xB5
	typeUint32  = 0xB6
	typeUint64  = 0xB7
	typeFloat32 = 0xB9
	typeFloat64 = 0xBA // same as floatF64
)

// TypeByte* constants identify the numeric element type within a TypedArray.
// Pass one of these to [Encoder.WriteChunkedTypedArray] to specify the wire type.
const (
	TypeByteInt8    byte = typeInt8
	TypeByteInt16   byte = typeInt16
	TypeByteInt32   byte = typeInt32
	TypeByteInt64   byte = typeInt64
	TypeByteUint8   byte = typeUint8
	TypeByteUint16  byte = typeUint16
	TypeByteUint32  byte = typeUint32
	TypeByteUint64  byte = typeUint64
	TypeByteFloat32 byte = typeFloat32
	TypeByteFloat64 byte = typeFloat64
)
