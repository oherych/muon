package muon

import (
	"reflect"
)

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
	zeroNumber       = 0xA0
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

var (
	kindToMuonType = map[reflect.Kind]byte{
		reflect.Int:     typeInt64,
		reflect.Int8:    typeInt8,
		reflect.Int16:   typeInt16,
		reflect.Int32:   typeInt32,
		reflect.Int64:   typeInt64,
		reflect.Uint:    typeUint64,
		reflect.Uint8:   typeUint8,
		reflect.Uint16:  typeUint16,
		reflect.Uint32:  typeUint32,
		reflect.Uint64:  typeUint64,
		reflect.Float32: typeFloat32,
		reflect.Float64: typeFloat64,
	}

	muonTypeToType = map[byte]reflect.Type{
		typeInt8:    reflect.TypeOf(int8(1)),
		typeInt16:   reflect.TypeOf(int16(1)),
		typeInt32:   reflect.TypeOf(int32(1)),
		typeInt64:   reflect.TypeOf(int64(1)),
		typeUint8:   reflect.TypeOf(uint8(1)),
		typeUint16:  reflect.TypeOf(uint16(1)),
		typeUint32:  reflect.TypeOf(uint32(1)),
		typeUint64:  reflect.TypeOf(uint64(1)),
		typeFloat16: reflect.TypeOf(float32(1)),
		typeFloat32: reflect.TypeOf(float32(1)),
		typeFloat64: reflect.TypeOf(float64(1)),
	}
)
