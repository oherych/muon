package muon

import (
	"math"
)

type G string

var (
	testString         = "test"
	testLongString     = "test Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce mi mauris, fringilla a gravida ac, vulputate vitae dui. Proin rhoncus ante vitae purus mollis, id hendrerit tellus tempor. Aliquam ut ex nibh. Aenean quis quam eu purus scelerisque viverra ac consequat justo. Sed lobortis interdum facilisis. Sed euismod est magna, at iaculis nisi mollis a. Maecenas nec diam augue. Phasellus volutpat mattis nisi, eu sagittis enim tempor vitae. Aliquam sit amet ante finibus, bibendum lorem et, porta libero. Sed eu."
	testStringWithZero = "te" + string([]byte{StringEnd}) + "st"

	tests = map[string]struct {
		golang        interface{}
		encoded       []byte
		tokens        []Token
		unmarshal     map[string]unmarshalTest
		skipReading   bool
		withSignature bool
	}{
		"nil": {
			golang:  nil,
			encoded: []byte{NullValue},
			tokens:  []Token{{A: TokenLiteral, D: nil}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: nil},
				"slice":           {ptr: new([]int), exp: []int(nil)},
				"map":             {ptr: new(map[string]int), exp: map[string]int(nil)},
			},
		},
		"string_empty": {
			golang:  "",
			encoded: []byte{StringEnd},
			tokens:  []Token{{A: TokenLiteral, D: ""}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: ""},
				"string":          {ptr: new(string), exp: ""},
			},
		},
		"string": {
			golang:  testString,
			encoded: []byte{0x74, 0x65, 0x73, 0x74, StringEnd},
			tokens:  []Token{{A: TokenLiteral, D: testString}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: testString},
				"string":          {ptr: new(string), exp: testString},
			},
		},
		"string_with_zero": {
			golang:  testStringWithZero,
			encoded: []byte{CountTag, 0x5, 0x74, 0x65, 0x0, 0x73, 0x74},
			tokens:  []Token{{A: TokenLiteral, D: testStringWithZero}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: testStringWithZero},
				"string":          {ptr: new(string), exp: testStringWithZero},
			},
		},
		"long_string": {
			golang:  testLongString,
			encoded: []byte{CountTag, 0x86, 0x4, 0x74, 0x65, 0x73, 0x74, 0x20, 0x4c, 0x6f, 0x72, 0x65, 0x6d, 0x20, 0x69, 0x70, 0x73, 0x75, 0x6d, 0x20, 0x64, 0x6f, 0x6c, 0x6f, 0x72, 0x20, 0x73, 0x69, 0x74, 0x20, 0x61, 0x6d, 0x65, 0x74, 0x2c, 0x20, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x63, 0x74, 0x65, 0x74, 0x75, 0x72, 0x20, 0x61, 0x64, 0x69, 0x70, 0x69, 0x73, 0x63, 0x69, 0x6e, 0x67, 0x20, 0x65, 0x6c, 0x69, 0x74, 0x2e, 0x20, 0x46, 0x75, 0x73, 0x63, 0x65, 0x20, 0x6d, 0x69, 0x20, 0x6d, 0x61, 0x75, 0x72, 0x69, 0x73, 0x2c, 0x20, 0x66, 0x72, 0x69, 0x6e, 0x67, 0x69, 0x6c, 0x6c, 0x61, 0x20, 0x61, 0x20, 0x67, 0x72, 0x61, 0x76, 0x69, 0x64, 0x61, 0x20, 0x61, 0x63, 0x2c, 0x20, 0x76, 0x75, 0x6c, 0x70, 0x75, 0x74, 0x61, 0x74, 0x65, 0x20, 0x76, 0x69, 0x74, 0x61, 0x65, 0x20, 0x64, 0x75, 0x69, 0x2e, 0x20, 0x50, 0x72, 0x6f, 0x69, 0x6e, 0x20, 0x72, 0x68, 0x6f, 0x6e, 0x63, 0x75, 0x73, 0x20, 0x61, 0x6e, 0x74, 0x65, 0x20, 0x76, 0x69, 0x74, 0x61, 0x65, 0x20, 0x70, 0x75, 0x72, 0x75, 0x73, 0x20, 0x6d, 0x6f, 0x6c, 0x6c, 0x69, 0x73, 0x2c, 0x20, 0x69, 0x64, 0x20, 0x68, 0x65, 0x6e, 0x64, 0x72, 0x65, 0x72, 0x69, 0x74, 0x20, 0x74, 0x65, 0x6c, 0x6c, 0x75, 0x73, 0x20, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x2e, 0x20, 0x41, 0x6c, 0x69, 0x71, 0x75, 0x61, 0x6d, 0x20, 0x75, 0x74, 0x20, 0x65, 0x78, 0x20, 0x6e, 0x69, 0x62, 0x68, 0x2e, 0x20, 0x41, 0x65, 0x6e, 0x65, 0x61, 0x6e, 0x20, 0x71, 0x75, 0x69, 0x73, 0x20, 0x71, 0x75, 0x61, 0x6d, 0x20, 0x65, 0x75, 0x20, 0x70, 0x75, 0x72, 0x75, 0x73, 0x20, 0x73, 0x63, 0x65, 0x6c, 0x65, 0x72, 0x69, 0x73, 0x71, 0x75, 0x65, 0x20, 0x76, 0x69, 0x76, 0x65, 0x72, 0x72, 0x61, 0x20, 0x61, 0x63, 0x20, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x71, 0x75, 0x61, 0x74, 0x20, 0x6a, 0x75, 0x73, 0x74, 0x6f, 0x2e, 0x20, 0x53, 0x65, 0x64, 0x20, 0x6c, 0x6f, 0x62, 0x6f, 0x72, 0x74, 0x69, 0x73, 0x20, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x64, 0x75, 0x6d, 0x20, 0x66, 0x61, 0x63, 0x69, 0x6c, 0x69, 0x73, 0x69, 0x73, 0x2e, 0x20, 0x53, 0x65, 0x64, 0x20, 0x65, 0x75, 0x69, 0x73, 0x6d, 0x6f, 0x64, 0x20, 0x65, 0x73, 0x74, 0x20, 0x6d, 0x61, 0x67, 0x6e, 0x61, 0x2c, 0x20, 0x61, 0x74, 0x20, 0x69, 0x61, 0x63, 0x75, 0x6c, 0x69, 0x73, 0x20, 0x6e, 0x69, 0x73, 0x69, 0x20, 0x6d, 0x6f, 0x6c, 0x6c, 0x69, 0x73, 0x20, 0x61, 0x2e, 0x20, 0x4d, 0x61, 0x65, 0x63, 0x65, 0x6e, 0x61, 0x73, 0x20, 0x6e, 0x65, 0x63, 0x20, 0x64, 0x69, 0x61, 0x6d, 0x20, 0x61, 0x75, 0x67, 0x75, 0x65, 0x2e, 0x20, 0x50, 0x68, 0x61, 0x73, 0x65, 0x6c, 0x6c, 0x75, 0x73, 0x20, 0x76, 0x6f, 0x6c, 0x75, 0x74, 0x70, 0x61, 0x74, 0x20, 0x6d, 0x61, 0x74, 0x74, 0x69, 0x73, 0x20, 0x6e, 0x69, 0x73, 0x69, 0x2c, 0x20, 0x65, 0x75, 0x20, 0x73, 0x61, 0x67, 0x69, 0x74, 0x74, 0x69, 0x73, 0x20, 0x65, 0x6e, 0x69, 0x6d, 0x20, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x20, 0x76, 0x69, 0x74, 0x61, 0x65, 0x2e, 0x20, 0x41, 0x6c, 0x69, 0x71, 0x75, 0x61, 0x6d, 0x20, 0x73, 0x69, 0x74, 0x20, 0x61, 0x6d, 0x65, 0x74, 0x20, 0x61, 0x6e, 0x74, 0x65, 0x20, 0x66, 0x69, 0x6e, 0x69, 0x62, 0x75, 0x73, 0x2c, 0x20, 0x62, 0x69, 0x62, 0x65, 0x6e, 0x64, 0x75, 0x6d, 0x20, 0x6c, 0x6f, 0x72, 0x65, 0x6d, 0x20, 0x65, 0x74, 0x2c, 0x20, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x20, 0x6c, 0x69, 0x62, 0x65, 0x72, 0x6f, 0x2e, 0x20, 0x53, 0x65, 0x64, 0x20, 0x65, 0x75, 0x2e},
			tokens:  []Token{{A: TokenLiteral, D: testLongString}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: testLongString},
				"string":          {ptr: new(string), exp: testLongString},
			},
		},
		"kind_string": {
			golang:  G(testString),
			encoded: []byte{0x74, 0x65, 0x73, 0x74, StringEnd},
			tokens:  []Token{{A: TokenLiteral, D: testString}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: testString},
				"kind_string":     {ptr: new(G), exp: G(testString)},
				"string":          {ptr: new(string), exp: testString},
			},
		},

		"pointer": {
			golang:  &testString,
			encoded: []byte{0x74, 0x65, 0x73, 0x74, StringEnd},
			tokens:  []Token{{A: TokenLiteral, D: testString}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: testString},
				"string":          {ptr: new(string), exp: testString},
			},
		},

		"true": {
			golang:  true,
			encoded: []byte{BoolTrue},
			tokens:  []Token{{A: TokenLiteral, D: true}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: true},
				"bool":            {ptr: new(bool), exp: true},
			},
		},
		"false": {
			golang:  false,
			encoded: []byte{BoolFalse},
			tokens:  []Token{{A: TokenLiteral, D: false}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: false},
				"bool":            {ptr: new(bool), exp: false},
			},
		},

		"int": {
			golang:  math.MaxInt64,
			encoded: []byte{0xb3, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
			tokens:  []Token{{A: TokenLiteral, D: int64(math.MaxInt64)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: int64(math.MaxInt64)},
				"int64":           {ptr: new(int64), exp: int64(math.MaxInt64)},
			},
		},
		"int8": {
			golang:  int8(math.MaxInt8),
			encoded: []byte{TypeInt8, 0x7f},
			tokens:  []Token{{A: TokenLiteral, D: int8(math.MaxInt8)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: int8(math.MaxInt8)},
				"int8":            {ptr: new(int8), exp: int8(math.MaxInt8)},
				"int16":           {ptr: new(int16), exp: int16(math.MaxInt8)},
				"int32":           {ptr: new(int32), exp: int32(math.MaxInt8)},
				"int64":           {ptr: new(int64), exp: int64(math.MaxInt8)},
			},
		},
		"int16": {
			golang:  int16(math.MaxInt16),
			encoded: []byte{TypeInt16, 0xff, 0x7f},
			tokens:  []Token{{A: TokenLiteral, D: int16(math.MaxInt16)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: int16(math.MaxInt16)},
				"int16":           {ptr: new(int16), exp: int16(math.MaxInt16)},
				"int32":           {ptr: new(int32), exp: int32(math.MaxInt16)},
				"int64":           {ptr: new(int64), exp: int64(math.MaxInt16)},
			},
		},
		"int32": {
			golang:  int32(math.MaxInt32),
			encoded: []byte{TypeInt32, 0xff, 0xff, 0xff, 0x7f},
			tokens:  []Token{{A: TokenLiteral, D: int32(math.MaxInt32)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: int32(math.MaxInt32)},
				"int32":           {ptr: new(int32), exp: int32(math.MaxInt32)},
				"int64":           {ptr: new(int64), exp: int64(math.MaxInt32)},
			},
		},
		"int64": {
			golang:  math.MaxInt64,
			encoded: []byte{0xb3, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
			tokens:  []Token{{A: TokenLiteral, D: int64(math.MaxInt64)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: int64(math.MaxInt64)},
				"int64":           {ptr: new(int64), exp: int64(math.MaxInt64)},
			},
		},

		"int_negative": {
			golang:  math.MinInt,
			encoded: []byte{TypeInt64, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80},
			tokens:  []Token{{A: TokenLiteral, D: int64(math.MinInt)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: int64(math.MinInt)},
				"int64":           {ptr: new(int64), exp: int64(math.MinInt)},
			},
		},
		"int8_negative": {
			golang:  int8(math.MinInt8),
			encoded: []byte{TypeInt8, 0x80},
			tokens:  []Token{{A: TokenLiteral, D: int8(math.MinInt8)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: int8(math.MinInt8)},
				"int8":            {ptr: new(int8), exp: int8(math.MinInt8)},
				"int16":           {ptr: new(int16), exp: int16(math.MinInt8)},
				"int32":           {ptr: new(int32), exp: int32(math.MinInt8)},
				"int64":           {ptr: new(int64), exp: int64(math.MinInt8)},
			},
		},
		"int16_negative": {
			golang:  int16(math.MinInt16),
			encoded: []byte{TypeInt16, 0x0, 0x80},
			tokens:  []Token{{A: TokenLiteral, D: int16(math.MinInt16)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: int16(math.MinInt16)},
				"int16":           {ptr: new(int16), exp: int16(math.MinInt16)},
				"int32":           {ptr: new(int32), exp: int32(math.MinInt16)},
				"int64":           {ptr: new(int64), exp: int64(math.MinInt16)},
			},
		},
		"int32_negative": {
			golang:  int32(math.MinInt32),
			encoded: []byte{TypeInt32, 0x0, 0x0, 0x0, 0x80},
			tokens:  []Token{{A: TokenLiteral, D: int32(math.MinInt32)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: int32(math.MinInt32)},
				"int32":           {ptr: new(int32), exp: int32(math.MinInt32)},
				"int64":           {ptr: new(int64), exp: int64(math.MinInt32)},
			},
		},
		"int64_negative": {
			golang:  int64(math.MinInt64),
			encoded: []byte{TypeInt64, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80},
			tokens:  []Token{{A: TokenLiteral, D: int64(math.MinInt64)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: int64(math.MinInt64)},
				"int64":           {ptr: new(int64), exp: int64(math.MinInt64)},
			},
		},

		"uint": {
			golang:  uint64(math.MaxUint64),
			encoded: []byte{TypeUint64, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			tokens:  []Token{{A: TokenLiteral, D: uint64(math.MaxUint64)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: uint64(math.MaxUint64)},
				"uint64":          {ptr: new(uint64), exp: uint64(math.MaxUint64)},
			},
		},
		"uint8": {
			golang:  uint8(math.MaxUint8),
			encoded: []byte{TypeUint8, 0xff},
			tokens:  []Token{{A: TokenLiteral, D: uint8(math.MaxUint8)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: uint8(math.MaxUint8)},
				"uint8":           {ptr: new(uint8), exp: uint8(math.MaxUint8)},
				"uint16":          {ptr: new(uint16), exp: uint16(math.MaxUint8)},
				"uint32":          {ptr: new(uint32), exp: uint32(math.MaxUint8)},
				"uint64":          {ptr: new(uint64), exp: uint64(math.MaxUint8)},
			},
		},
		"uint16": {
			golang:  uint16(math.MaxUint16),
			encoded: []byte{TypeUint16, 0xff, 0xff},
			tokens:  []Token{{A: TokenLiteral, D: uint16(math.MaxUint16)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: uint16(math.MaxUint16)},
				"uint16":          {ptr: new(uint16), exp: uint16(math.MaxUint16)},
				"uint32":          {ptr: new(uint32), exp: uint32(math.MaxUint16)},
				"uint64":          {ptr: new(uint64), exp: uint64(math.MaxUint16)},
			},
		},
		"uint32": {
			golang:  uint32(math.MaxUint32),
			encoded: []byte{TypeUint32, 0xff, 0xff, 0xff, 0xff},
			tokens:  []Token{{A: TokenLiteral, D: uint32(math.MaxUint32)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: uint32(math.MaxUint32)},
				"uint32":          {ptr: new(uint32), exp: uint32(math.MaxUint32)},
				"uint64":          {ptr: new(uint64), exp: uint64(math.MaxUint32)},
			},
		},
		"uint64": {
			golang:  uint64(math.MaxUint64),
			encoded: []byte{TypeUint64, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			tokens:  []Token{{A: TokenLiteral, D: uint64(math.MaxUint64)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: uint64(math.MaxUint64)},
				"uint64":          {ptr: new(uint64), exp: uint64(math.MaxUint64)},
			},
		},

		"float32": {
			golang:  float32(math.MaxFloat32),
			encoded: []byte{TypeFloat32, 0xff, 0xff, 0x7f, 0x7f},
			tokens:  []Token{{A: TokenLiteral, D: float32(math.MaxFloat32)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: float32(math.MaxFloat32)},
				"float32":         {ptr: new(float32), exp: float32(math.MaxFloat32)},
				"float64":         {ptr: new(float64), exp: float64(math.MaxFloat32)},
			},
		},

		"float64": {
			golang:  float64(math.MaxFloat64),
			encoded: []byte{TypeFloat64, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xef, 0x7f},
			tokens:  []Token{{A: TokenLiteral, D: float64(math.MaxFloat64)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: float64(math.MaxFloat64)},
				"float64":         {ptr: new(float64), exp: float64(math.MaxFloat64)},
			},
		},

		"float32_negative": {
			golang:  float32(math.SmallestNonzeroFloat32),
			encoded: []byte{TypeFloat32, 0x1, 0x0, 0x0, 0x0},
			tokens:  []Token{{A: TokenLiteral, D: float32(math.SmallestNonzeroFloat32)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: float32(math.SmallestNonzeroFloat32)},
				"float32":         {ptr: new(float32), exp: float32(math.SmallestNonzeroFloat32)},
				"float64":         {ptr: new(float64), exp: float64(math.SmallestNonzeroFloat32)},
			},
		},

		"float64_negative": {
			golang:  float64(math.SmallestNonzeroFloat64),
			encoded: []byte{TypeFloat64, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			tokens:  []Token{{A: TokenLiteral, D: float64(math.SmallestNonzeroFloat64)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: float64(math.SmallestNonzeroFloat64)},
				"float64":         {ptr: new(float64), exp: float64(math.SmallestNonzeroFloat64)},
			},
		},

		"flot_nan": {
			golang:      math.NaN(),
			encoded:     []byte{NanValue},
			tokens:      []Token{{A: TokenLiteral, D: math.NaN()}},
			skipReading: true,
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: math.NaN()},
				"float32":         {ptr: new(float32), exp: math.NaN()},
				"float64":         {ptr: new(float64), exp: math.NaN()},
			},
		},
		"flot_-Inf": {
			golang:  math.Inf(-1),
			encoded: []byte{NegativeInfValue},
			tokens:  []Token{{A: TokenLiteral, D: math.Inf(-1)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: math.Inf(-1)},
				"float32":         {ptr: new(float32), exp: float32(math.Inf(-1))},
				"float64":         {ptr: new(float64), exp: math.Inf(-1)},
			},
		},
		"flot_+Inf": {
			golang:  math.Inf(1),
			encoded: []byte{PositiveInfValue},
			tokens:  []Token{{A: TokenLiteral, D: math.Inf(1)}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: math.Inf(1)},
				"float32":         {ptr: new(float32), exp: float32(math.Inf(1))},
				"float64":         {ptr: new(float64), exp: math.Inf(1)},
			},
		},

		"slice": {
			golang:  []interface{}{testString, true},
			encoded: []byte{ListStart, 0x74, 0x65, 0x73, 0x74, StringEnd, BoolTrue, ListEnd},
			tokens:  []Token{{A: TokenListStart}, {A: TokenLiteral, D: testString}, {A: TokenLiteral, D: true}, {A: TokenListEnd}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: []interface{}{testString, true}},
				"array":           {ptr: new([2]interface{}), exp: [2]interface{}{testString, true}},
			},
		},
		"slice_bytes": {
			golang:    []byte{0x1, 0x30},
			encoded:   []byte{TypedArray, TypeUint8, 2, 0x1, 0x30},
			tokens:    []Token{{A: TokenTypedArray, D: []byte{0x1, 0x30}}},
			unmarshal: map[string]unmarshalTest{
				// interface
				//"array": {ptr: new([]byte), exp: []byte{0x1, 0x30}},
			},
		},
		"byte_int": {
			golang:  []int{5, 16},
			encoded: []byte{TypedArray, TypeInt64, 0x2, 0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x10, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			tokens:  []Token{{A: TokenTypedArray, D: []int64{5, 16}}},
		},
		"byte_uint": {
			golang:  []uint{5, 16},
			encoded: []byte{TypedArray, TypeUint64, 0x2, 0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x10, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			tokens:  []Token{{A: TokenTypedArray, D: []uint64{5, 16}}},
		},

		"array": {
			golang:  [2]bool{false, true},
			encoded: []byte{ListStart, BoolFalse, BoolTrue, ListEnd},
			tokens:  []Token{{A: TokenListStart}, {A: TokenLiteral, D: false}, {A: TokenLiteral, D: true}, {A: TokenListEnd}},
		},
		"map": {
			golang:  map[string]string{"a": "b"},
			encoded: []byte{DictStart, 0x61, StringEnd, 0x62, StringEnd, DictEnd},
			tokens:  []Token{{A: TokenDictStart}, {A: TokenLiteral, D: "a"}, {A: TokenLiteral, D: "b"}, {A: TokenDictEnd}},
		},
		"simple_structure": {
			golang:  SimpleStructure{B: "b_value", SkipMy: "skip_me"},
			encoded: []byte{DictStart, 0x6d, 0x79, 0x5f, 0x62, StringEnd, 0x62, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, StringEnd, DictEnd},
			tokens:  []Token{{A: TokenDictStart}, {A: TokenLiteral, D: "my_b"}, {A: TokenLiteral, D: "b_value"}, {A: TokenDictEnd}},
			unmarshal: map[string]unmarshalTest{
				"struct": {ptr: new(SimpleStructure), exp: SimpleStructure{B: "b_value"}},
			},
		},

		"with_signature": {
			withSignature: true,
			golang:        testString,
			encoded:       []byte{SignatureStart, 0xb5, 0x30, 0x31, 0x74, 0x65, 0x73, 0x74, StringEnd},
			tokens:        []Token{{A: TokenSignature}, {A: TokenLiteral, D: testString}},
			unmarshal: map[string]unmarshalTest{
				"empty_interface": {ptr: new(interface{}), exp: testString},
				"string":          {ptr: new(string), exp: testString},
			},
		},
	}
)

type unmarshalTest struct {
	ptr interface{}
	exp interface{}
}

type SimpleStructure struct {
	B      string `muon:"my_b"`
	SkipMy string `muon:"-"`
}
