package utils

import "strconv"

func MarkdownValueList[T any](strMapper func(T) string, values []T) string {
	switch {
	case len(values) == 0:
		return ""
	case len(values) == 1:
		return "`" + strMapper(values[0]) + "`"
	default:
		s := ""
		for i := 0; i < len(values)-1; i++ {
			s += "`" + strMapper(values[i]) + "`, "
		}
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
