package utils

import (
	"strconv"
	"strings"
)

func MarkdownValueList[T any](strMapper func(T) string, values []T) string {
	switch {
	case len(values) == 0:
		return ""
	case len(values) == 1:
		return "`" + strMapper(values[0]) + "`"
	default:
		s := ""
		var sSb13 strings.Builder
		for i := range len(values) - 1 {
			sSb13.WriteString("`" + strMapper(values[i]) + "`, ")
		}
		s += sSb13.String()
		s += " and `" + strMapper(values[len(values)-1]) + "`"
		return s
	}
}

func MarkdownValueListInt(values []int) string {
	return MarkdownValueList(func(i int) string { return strconv.Itoa(i) }, values)
}

func MarkdownValueListString(values []string) string {
	return MarkdownValueList(func(s string) string { return s }, values)
}
