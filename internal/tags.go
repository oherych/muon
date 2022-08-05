package internal

import (
	"reflect"
	"strings"
)

const (
	tag = "muon"
)

type TagInfo struct {
	Name string
	Skip bool
}

func ParseTags(field reflect.StructField) TagInfo {
	val := field.Tag.Get(tag)
	parts := strings.Split(val, ",")

	skip := parts[0] == "-"

	if parts[0] == "" {
		parts[0] = strings.ToLower(field.Name)
	}

	return TagInfo{
		Name: parts[0],
		Skip: skip,
	}
}
