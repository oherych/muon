package muon

import (
	"fmt"
	"reflect"
)

// Unmarshal decodes the muon-encoded data into the value pointed to by target.
// target must be a non-nil pointer. Supported target types mirror the encoding
// side: bool, all int/uint/float sizes, string, slice, array, map, struct, and
// pointer. Use *interface{} to decode any value without a known schema.
func Unmarshal(data []byte, target interface{}) error {
	d := NewDecoder(data)
	return d.Unmarshal(target)
}

// Unmarshal reads the next value from the stream and stores it into target.
// target must be a non-nil pointer.
func (d *Decoder) Unmarshal(target interface{}) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errInvalidTarget("target must be a non-nil pointer")
	}
	tok, err := d.r.Next()
	if err != nil {
		return err
	}
	return d.unmarshalToken(tok, rv.Elem())
}

func (d *Decoder) unmarshalToken(tok Token, v reflect.Value) error {
	// transparently skip magic and count tags
	if tok.A == TokenMagic || tok.A == TokenCount {
		next, err := d.r.Next()
		if err != nil {
			return err
		}
		return d.unmarshalToken(next, v)
	}

	// nil token → zero the target
	if tok.A == tokenNil {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	// pointer: allocate if nil, then unwrap
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return d.unmarshalToken(tok, v.Elem())
	}

	// interface{}: use the high-level tokenToValue path
	if v.Kind() == reflect.Interface {
		val, err := d.tokenToValue(tok)
		if err != nil {
			return err
		}
		if val == nil {
			v.Set(reflect.Zero(v.Type()))
		} else {
			v.Set(reflect.ValueOf(val))
		}
		return nil
	}

	switch tok.A {
	case tokenTrue, tokenFalse:
		return d.unmarshalBool(tok, v)
	case tokenInt:
		return d.unmarshalInt(tok, v)
	case TokenFloat:
		return d.unmarshalFloat(tok, v)
	case TokenString:
		return d.unmarshalString(tok, v)
	case TokenTypedArray:
		return d.unmarshalTypedArray(tok, v)
	case tokenListStart:
		return d.unmarshalList(v)
	case tokenDictStart:
		return d.unmarshalDict(v)
	default:
		return errUnexpectedToken(tok.A)
	}
}

func (d *Decoder) unmarshalBool(tok Token, v reflect.Value) error {
	if v.Kind() != reflect.Bool {
		return errTypeMismatch(tok.A, v.Interface())
	}
	v.SetBool(tok.A == tokenTrue)
	return nil
}

func (d *Decoder) unmarshalInt(tok Token, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := toInt64(tok.Data)
		if err != nil {
			return err
		}
		v.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := toUint64(tok.Data)
		if err != nil {
			return err
		}
		v.SetUint(n)
		return nil
	}
	return errTypeMismatch(tok.A, v.Interface())
}

func (d *Decoder) unmarshalFloat(tok Token, v reflect.Value) error {
	if v.Kind() != reflect.Float32 && v.Kind() != reflect.Float64 {
		return errTypeMismatch(tok.A, v.Interface())
	}
	v.SetFloat(tok.Data.(float64))
	return nil
}

func (d *Decoder) unmarshalString(tok Token, v reflect.Value) error {
	if v.Kind() != reflect.String {
		return errTypeMismatch(tok.A, v.Interface())
	}
	v.SetString(tok.Data.(string))
	return nil
}

func (d *Decoder) unmarshalTypedArray(tok Token, v reflect.Value) error {
	src := reflect.ValueOf(tok.Data)
	switch v.Kind() {
	case reflect.Slice:
		if !src.Type().AssignableTo(v.Type()) {
			// attempt element-wise conversion
			out := reflect.MakeSlice(v.Type(), src.Len(), src.Len())
			for i := 0; i < src.Len(); i++ {
				elem := src.Index(i)
				if !elem.Type().ConvertibleTo(v.Type().Elem()) {
					return errTypeMismatch(tok.A, v.Interface())
				}
				out.Index(i).Set(elem.Convert(v.Type().Elem()))
			}
			v.Set(out)
			return nil
		}
		v.Set(src)
		return nil
	case reflect.Array:
		n := v.Len()
		if src.Len() < n {
			n = src.Len()
		}
		for i := 0; i < n; i++ {
			elem := src.Index(i)
			if !elem.Type().ConvertibleTo(v.Type().Elem()) {
				return errTypeMismatch(tok.A, v.Interface())
			}
			v.Index(i).Set(elem.Convert(v.Type().Elem()))
		}
		return nil
	}
	return errTypeMismatch(tok.A, v.Interface())
}

func (d *Decoder) unmarshalList(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Slice:
		elemType := v.Type().Elem()
		for {
			tok, err := d.r.Next()
			if err != nil {
				return err
			}
			if tok.A == tokenListEnd {
				break
			}
			elem := reflect.New(elemType).Elem()
			if err := d.unmarshalToken(tok, elem); err != nil {
				return err
			}
			v.Set(reflect.Append(v, elem))
		}
		return nil
	case reflect.Array:
		i := 0
		for {
			tok, err := d.r.Next()
			if err != nil {
				return err
			}
			if tok.A == tokenListEnd {
				break
			}
			if i < v.Len() {
				if err := d.unmarshalToken(tok, v.Index(i)); err != nil {
					return err
				}
			}
			i++
		}
		return nil
	}
	return errTypeMismatch(tokenListStart, v.Interface())
}

func (d *Decoder) unmarshalDict(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Struct:
		return d.unmarshalStruct(v)
	case reflect.Map:
		return d.unmarshalMap(v)
	}
	return errTypeMismatch(tokenDictStart, v.Interface())
}

func (d *Decoder) unmarshalStruct(v reflect.Value) error {
	// build field name → index map using muon tags
	t := v.Type()
	fields := make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		if !tf.IsExported() {
			continue
		}
		name := tf.Name
		if tag, ok := tf.Tag.Lookup("muon"); ok {
			if tag == "-" {
				continue
			}
			if tag != "" {
				name = tag
			}
		} else {
			// default: lowercase field name (mirrors writeStruct behaviour)
			if len(name) > 0 {
				name = string([]byte{name[0] | 0x20}) + name[1:]
			}
		}
		fields[name] = i
	}

	for {
		keyTok, err := d.r.Next()
		if err != nil {
			return err
		}
		if keyTok.A == tokenDictEnd {
			return nil
		}
		if keyTok.A != TokenString {
			return errUnexpectedToken(keyTok.A)
		}
		key := keyTok.Data.(string)

		idx, ok := fields[key]
		if !ok {
			// unknown field: read and discard the value
			if err := d.skipValue(); err != nil {
				return err
			}
			continue
		}
		valTok, err := d.r.Next()
		if err != nil {
			return err
		}
		if err := d.unmarshalToken(valTok, v.Field(idx)); err != nil {
			return err
		}
	}
}

func (d *Decoder) unmarshalMap(v reflect.Value) error {
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}
	keyType := v.Type().Key()
	elemType := v.Type().Elem()

	intKeyType := byte(0)
	first := true

	for {
		var keyTok Token
		var err error
		if first {
			keyTok, err = d.r.Next()
			first = false
		} else if intKeyType != 0 {
			keyTok, err = d.r.NextIntKey(intKeyType)
		} else {
			keyTok, err = d.r.Next()
		}
		if err != nil {
			return err
		}
		if keyTok.A == tokenDictEnd {
			return nil
		}

		// remember int key type for subsequent keys
		if keyTok.A == tokenInt && intKeyType == 0 {
			intKeyType = d.r.lastIntKeyType
		}

		keyVal := reflect.New(keyType).Elem()
		if err := d.unmarshalToken(keyTok, keyVal); err != nil {
			return err
		}

		valTok, err := d.r.Next()
		if err != nil {
			return err
		}
		elemVal := reflect.New(elemType).Elem()
		if err := d.unmarshalToken(valTok, elemVal); err != nil {
			return err
		}
		v.SetMapIndex(keyVal, elemVal)
	}
}

// skipValue reads and discards the next complete value (including nested structures).
func (d *Decoder) skipValue() error {
	tok, err := d.r.Next()
	if err != nil {
		return err
	}
	switch tok.A {
	case tokenListStart:
		depth := 1
		for depth > 0 {
			t, err := d.r.Next()
			if err != nil {
				return err
			}
			if t.A == tokenListStart {
				depth++
			} else if t.A == tokenListEnd {
				depth--
			}
		}
	case tokenDictStart:
		depth := 1
		for depth > 0 {
			t, err := d.r.Next()
			if err != nil {
				return err
			}
			if t.A == tokenDictStart {
				depth++
			} else if t.A == tokenDictEnd {
				depth--
			}
		}
	}
	return nil
}

func toInt64(v interface{}) (int64, error) {
	switch n := v.(type) {
	case int:
		return int64(n), nil
	case int64:
		return n, nil
	case uint64:
		return int64(n), nil
	}
	return 0, fmt.Errorf("cannot convert %T to int64", v)
}

func toUint64(v interface{}) (uint64, error) {
	switch n := v.(type) {
	case int:
		return uint64(n), nil
	case int64:
		return uint64(n), nil
	case uint64:
		return n, nil
	}
	return 0, fmt.Errorf("cannot convert %T to uint64", v)
}
