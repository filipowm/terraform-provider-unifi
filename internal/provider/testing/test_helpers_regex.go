package testing

import (
	"fmt"
	"regexp"
)

func MissingArgumentErrorRegex(arg string) *regexp.Regexp {
	r, _ := regexp.Compile(fmt.Sprintf(`%q is required`, arg))
	return r
}
