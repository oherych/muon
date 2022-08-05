package internal

import (
	"reflect"
	"testing"
)

func TestParseTags(t *testing.T) {
	type AA struct {
		Without   string
		Name      string `muon:"my_name"`
		EmptyName string `muon:"-"`
	}

	typ := reflect.TypeOf(AA{})

	for i := 0; i < typ.NumField(); i++ {
		got := ParseTags(typ.Field(i))

		t.Log(got)
	}
}
