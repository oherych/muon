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

	stringStart = 0x82
	typedArray  = 0x84

	typeInt8    = 0xB0
	typeInt16   = 0xB1
	typeInt32   = 0xB2
	typeInt64   = 0xB3
	typeUint8   = 0xB4
	typeUint16  = 0xB5
	typeUint32  = 0xB6
	typeUint64  = 0xB7
	typeFloat16 = 0xB8
	typeFloat32 = 0xB9
	typeFloat64 = 0xBA
)
