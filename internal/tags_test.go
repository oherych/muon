package internal

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTags(t *testing.T) {
	type AA struct {
		WithoutName string
		Name        string `muon:"my_name"`
		EmptyName   string `muon:"-"`
	}

	exp := map[string]TagInfo{
		"WithoutName": {Name: "without_name", Skip: false},
		"Name":        {Name: "my_name", Skip: false},
		"EmptyName":   {Name: "-", Skip: true},
	}

	typ := reflect.TypeOf(AA{})

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)

		t.Run(f.Name, func(t *testing.T) {

			got := ParseTags("muon", f)

			assert.Equal(t, exp[f.Name], got)
		})
	}
}
