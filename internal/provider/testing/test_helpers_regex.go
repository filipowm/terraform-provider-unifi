package testing

import (
	"fmt"
	"regexp"
)

func MissingArgumentErrorRegex(arg string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`%q is required`, arg))
}
