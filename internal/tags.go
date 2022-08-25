package internal

import (
	"reflect"
	"regexp"
	"strings"
)

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

type TagInfo struct {
	Name string
	Skip bool
}

func ParseTags(tagName string, field reflect.StructField) TagInfo {
	val := field.Tag.Get(tagName)
	parts := strings.Split(val, ",")

	skip := parts[0] == "-"

	if parts[0] == "" {
		parts[0] = toSnakeCase(field.Name)
	}

	return TagInfo{
		Name: parts[0],
		Skip: skip,
	}
}

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
