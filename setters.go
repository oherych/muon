package muon

import (
	"fmt"
	"github.com/oherych/muon/internal"
	"reflect"
)

type setter struct {
	Decoder *Decoder
}

func (s setter) Read(target reflect.Value) error {
	token, err := s.Decoder.Next()
	if err != nil {
		return err
	}

	if token.A == TokenSignature {
		return s.Read(target)
	}

	return s.Set(token, target)
}

func (s setter) Set(token Token, target reflect.Value) error {
	kind := target.Kind()

	if token.A == TokenLiteral && token.D == nil {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	if (kind >= reflect.Bool && kind <= reflect.Float64) || kind == reflect.String {
		return s.SetLiteral(token, target)
	}

	if kind == reflect.Struct {
		return s.SetStruct(token, target)
	}

	if kind == reflect.Interface {
		return s.SetInterface(token, target)
	}

	if kind == reflect.Slice || kind == reflect.Array {
		return s.SetSlice(token, target)
	}

	panic("unexpected type")
}

func (s setter) SetStruct(token Token, target reflect.Value) error {
	if token.A != tokenDictStart {
		panic("wrong type")
	}

	targetType := target.Type()

	fields := make(map[string]int, targetType.NumField())
	for i := 0; i < targetType.NumField(); i++ {
		tf := targetType.Field(i)

		tagInfo := internal.ParseTags(tf)
		if tagInfo.Skip {
			continue
		}

		fields[tagInfo.Name] = i
	}

	for {
		cur, err := s.Decoder.Next()
		if err != nil {
			return err
		}
		if cur.A == tokenDictEnd {
			break
		}

		var key string
		if err := s.SetLiteral(cur, reflect.ValueOf(&key).Elem()); err != nil {
			return err
		}

		i, ok := fields[key]
		if !ok {
			continue
		}

		if err := s.Read(target.Field(i)); err != nil {
			return err
		}
	}

	return nil
}

func (s setter) SetSlice(token Token, target reflect.Value) error {
	tmp := target
	elementType := target.Type().Elem()

	if token.A == tokenListStart {
		var i int
		for {
			cur, err := s.Decoder.Next()
			if err != nil {
				return err
			}
			if cur.A == tokenListEnd {
				break
			}

			newValue := reflect.New(elementType).Elem()
			if err := s.Set(cur, newValue); err != nil {
				return err
			}

			if tmp.Kind() == reflect.Slice {
				tmp.Set(reflect.Append(tmp, newValue))
			} else {
				tmp.Index(i).Set(newValue)
				i++
			}

		}

		target.Set(tmp)

		return nil
	}

	if token.A == TokenTypedArray {
		fmt.Println(token)
		return newError(ErrCodeNotImplemented, "%s", token.A)
	}

	return newError(ErrCodeInvalidType, "can't apply %s to %s", token.A, target.Type())
}

func (s setter) SetInterface(token Token, target reflect.Value) error {
	var newType reflect.Type
	switch token.A {
	case TokenLiteral:
		newType = reflect.TypeOf(token.D)
	case tokenDictStart:
		newType = reflect.TypeOf(map[interface{}]interface{}{})
	case tokenListStart:
		newType = reflect.TypeOf([]interface{}{})
	default:
		panic("sds")
	}

	newValue := reflect.New(newType).Elem()

	if err := s.Set(token, newValue); err != nil {
		return err
	}

	target.Set(newValue)

	return nil
}

func (s setter) SetLiteral(token Token, target reflect.Value) error {
	if token.A != TokenLiteral {
		panic("Literal only")
	}

	tk := reflect.ValueOf(token.D)

	if target.Kind() == tk.Kind() {
		setValue(target, token.D)

		return nil
	}

	if !tk.CanConvert(target.Type()) {
		panic("wrong type")
	}

	nv := tk.Convert(target.Type())

	target.Set(nv)

	return nil
}
