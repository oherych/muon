package muon

import "reflect"

var (
	KindToMuonType = map[reflect.Kind]byte{
		reflect.Int:     TypeInt64,
		reflect.Int8:    TypeInt8,
		reflect.Int16:   TypeInt16,
		reflect.Int32:   TypeInt32,
		reflect.Int64:   TypeInt64,
		reflect.Uint:    TypeUint64,
		reflect.Uint8:   TypeUint8,
		reflect.Uint16:  TypeUint16,
		reflect.Uint32:  TypeUint32,
		reflect.Uint64:  TypeUint64,
		reflect.Float32: TypeFloat32,
		reflect.Float64: TypeFloat64,
	}

	MuonTypeToType = map[byte]reflect.Type{
		TypeInt8:    reflect.TypeOf(int8(1)),
		TypeInt16:   reflect.TypeOf(int16(1)),
		TypeInt32:   reflect.TypeOf(int32(1)),
		TypeInt64:   reflect.TypeOf(int64(1)),
		TypeUint8:   reflect.TypeOf(uint8(1)),
		TypeUint16:  reflect.TypeOf(uint16(1)),
		TypeUint32:  reflect.TypeOf(uint32(1)),
		TypeUint64:  reflect.TypeOf(uint64(1)),
		TypeFloat16: reflect.TypeOf(float32(1)),
		TypeFloat32: reflect.TypeOf(float32(1)),
		TypeFloat64: reflect.TypeOf(float64(1)),
	}
)
