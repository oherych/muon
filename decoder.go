package muon

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/oherych/muon/internal"
	"io"
	"reflect"
)

type Decoder struct {
	TagName string

	reader Reader
}

func NewDecoder(r io.Reader) Decoder {
	reader := Reader{B: bufio.NewReader(r)}
	reader.Reset()

	return Decoder{
		TagName: defaultTag,

		reader: reader,
	}
}

func (r *Decoder) Unmarshal(target interface{}) (err error) {
	rv := reflect.ValueOf(target)

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("parameter should be pointer")
	}

	return r.process(rv.Elem())
}

func (r *Decoder) Next() (Token, error) {
	return r.reader.Next()
}

func (r Decoder) One(token Token, target reflect.Value) error {
	kind := target.Kind()

	if token.A == TokenLiteral && token.D == nil {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	if (kind >= reflect.Bool && kind <= reflect.Float64) || kind == reflect.String {
		return r.setLiteral(token, target)
	}

	if kind == reflect.Struct {
		return r.setStruct(token, target)
	}

	if kind == reflect.Map {
		return r.setMap(token, target)
	}

	if kind == reflect.Interface {
		return r.setInterface(token, target)
	}

	if kind == reflect.Slice || kind == reflect.Array {
		return r.setSlice(token, target)
	}

	panic("unexpected type")
}

func (r Decoder) process(target reflect.Value) error {
	token, err := r.Next()
	if err != nil {
		return err
	}

	if token.A == TokenSignature {
		return r.process(target)
	}

	return r.One(token, target)
}

func (r Decoder) setMap(token Token, target reflect.Value) error {
	if token.A != TokenDictStart {
		panic("wrong type")
	}

	panic("implement me")
}

func (r Decoder) setStruct(token Token, target reflect.Value) error {
	if token.A != TokenDictStart {
		panic("wrong type")
	}

	targetType := target.Type()

	fields := make(map[string]int, targetType.NumField())
	for i := 0; i < targetType.NumField(); i++ {
		tf := targetType.Field(i)

		tagInfo := internal.ParseTags(r.TagName, tf)
		if tagInfo.Skip {
			continue
		}

		fields[tagInfo.Name] = i
	}

	for {
		cur, err := r.Next()
		if err != nil {
			return err
		}
		if cur.A == TokenDictEnd {
			break
		}

		var key string
		if err := r.setLiteral(cur, reflect.ValueOf(&key).Elem()); err != nil {
			return err
		}

		i, ok := fields[key]
		if !ok {
			continue
		}

		if err := r.process(target.Field(i)); err != nil {
			return err
		}
	}

	return nil
}

func (r Decoder) setSlice(token Token, target reflect.Value) error {
	tmp := target
	elementType := target.Type().Elem()

	if token.A == TokenListStart {
		var i int
		for {
			cur, err := r.Next()
			if err != nil {
				return err
			}
			if cur.A == TokenListEnd {
				break
			}

			newValue := reflect.New(elementType).Elem()
			if err := r.One(cur, newValue); err != nil {
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

func (r Decoder) setInterface(token Token, target reflect.Value) error {
	var newType reflect.Type
	switch token.A {
	case TokenLiteral:
		newType = reflect.TypeOf(token.D)
	case TokenDictStart:
		newType = reflect.TypeOf(map[interface{}]interface{}{})
	case TokenListStart:
		newType = reflect.TypeOf([]interface{}{})
	default:
		panic("sds")
	}

	newValue := reflect.New(newType).Elem()

	if err := r.One(token, newValue); err != nil {
		return err
	}

	target.Set(newValue)

	return nil
}

func (r Decoder) setLiteral(token Token, target reflect.Value) error {
	if token.A != TokenLiteral {
		panic("Literal only")
	}

	tk := reflect.ValueOf(token.D)

	if target.Kind() == tk.Kind() {
		r.setValue(target, token.D)

		return nil
	}

	if !tk.CanConvert(target.Type()) {
		panic("wrong type")
	}

	nv := tk.Convert(target.Type())

	target.Set(nv)

	return nil
}

func (r Decoder) setValue(target reflect.Value, v interface{}) {
	var n = reflect.ValueOf(v)
	if target.Type() != reflect.TypeOf(v) {
		n = n.Convert(target.Type())
	}

	target.Set(n)
}
