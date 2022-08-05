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
)
