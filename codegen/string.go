package codegen

import (
	"strings"
	"unicode"

	"github.com/u-robot/go-tdlib/tlparser"
)

func firstUpper(str string) string {
	for i, r := range str {
		return tlparser.FixAAA(string(unicode.ToUpper(r)) + str[i+1:])
	}

	return tlparser.FixAAA(str)
}

func firstLower(str string) string {
	for i, r := range str {
		return tlparser.FixAAA(string(unicode.ToLower(r)) + str[i+1:])
	}

	return tlparser.FixAAA(str)
}

func underscoreToCamelCase(str string) string {
	result := strings.Replace(strings.Title(strings.Replace(strings.ToLower(str), "_", " ", -1)), " ", "", -1)
	return tlparser.FixAAA(result)
}
