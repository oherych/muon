package muon

const (
	StringEnd         = 0x00
	TypedArray        = 0x84
	CountTag          = 0x8A
	SizeTag           = 0x8B
	ReferencedString  = 0x8C
	SignatureStart    = 0x8f
	ListStart         = 0x90
	ListEnd           = 0x91
	DictStart         = 0x92
	DictEnd           = 0x93
	ZeroNumber        = 0xA0
	BoolFalse         = 0xAA
	BoolTrue          = 0xAB
	NullValue         = 0xAC
	NanValue          = 0xAD
	NegativeInfValue  = 0xAE // -Inf
	PositiveInfValue  = 0xAF // +Inf
	TypeInt8          = 0xB0
	TypeInt16         = 0xB1
	TypeInt32         = 0xB2
	TypeInt64         = 0xB3
	TypeUint8         = 0xB4
	TypeUint16        = 0xB5
	TypeUint32        = 0xB6
	TypeUint64        = 0xB7
	TypeFloat16       = 0xB8
	TypeFloat32       = 0xB9
	TypeFloat64       = 0xBA
	SignatureVersion1 = 0x31
)

var (
	Signature = []byte{SignatureStart, 0xB5, 0x30, SignatureVersion1}
)
