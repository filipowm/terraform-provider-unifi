package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateValidator(t *testing.T) {
	t.Parallel()
	var testValidator validator = "test"
	testCases := []struct {
		params             []string
		expectedValidation string
	}{
		{[]string{"vpn", "802.1x", "custom"}, "test=vpn 802.1x custom"},
		{[]string{"vpn"}, "test=vpn"},
		{[]string{}, "test"},
		{[]string{"0"}, "test=0"},
		{[]string{"2-2"}, "test=2-2"},
		{[]string{""}, "test"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedValidation, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			v := validation{testValidator, tc.params}
			result := createValidator(v.v, v.params...)
			a.Equal(tc.expectedValidation, result)
		})
	}
}

func TestCreateValidations(t *testing.T) {
	t.Parallel()

	var testValidator validator = "test"
	testCases := []struct {
		params              [][]string
		expectedValidations string
	}{
		{[][]string{{"1", "2", "3"}}, "test=1 2 3"},
		{[][]string{{"1", "2", "3"}, {"4", "5", "6"}}, "test=1 2 3,test=4 5 6"},
		{[][]string{{}}, "test"},
		{[][]string{{}, {}}, "test,test"},
		{[][]string{{}, {"1"}, {}}, "test,test=1,test"},
		{[][]string{}, ""},
	}

	for _, c := range testCases {
		var expectedValidationTag string
		if len(c.params) == 0 {
			expectedValidationTag = ""
		} else {
			expectedValidationTag = fmt.Sprintf("validate:\"omitempty,%s\"", c.expectedValidations)
		}
		t.Run(expectedValidationTag, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			var validations []validation
			for _, params := range c.params {
				validations = append(validations, validation{testValidator, params})
			}
			result := createValidations(validations...)
			a.Equal(expectedValidationTag, result)
		})
	}
}

func TestDefineValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		validationComment, expected string
	}{
		{"a|b", "oneof=a b"},
		{"1|2", "oneof=1 2"},
		{".{1,2}", "gte=1,lte=2"},
		{".{1}", "len=1"},
		{".{1}", "len=1"},
		{"[\\d\\w]+", "w_regex"},
		{"[\\d\\w]*", "w_regex"},
		{"[\\w]+", "w_regex"},
		{"[\\w]*", "w_regex"},
		{"a", ""},
		{".{1}|.{5,6}", ""},
		{"a|.{5,6}", ""},
	}

	for _, c := range testCases {
		var fullExpected string
		if c.expected == "" {
			fullExpected = ""
		} else {
			fullExpected = fmt.Sprintf("validate:\"omitempty,%s\"", c.expected)
		}
		t.Run(c.expected, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			result := defineFieldValidation(c.validationComment)
			a.Equal(fullExpected, result)
		})
	}
}

func testValidationCommentCheck(t *testing.T, testCases []struct {
	validationComment validationComment
	expected          bool
}, fn func(validationComment) bool,
) {
	t.Helper()
	for _, c := range testCases {
		t.Run(fmt.Sprintf("%s-%t", c.validationComment, c.expected), func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			trimmed := trimWrappers(string(c.validationComment))
			// trimmed := string(c.validationComment) //trimWrappers(string(c.validationComment))
			result := fn(validationComment(trimmed))
			a.Equal(c.expected, result)
		})
	}
}

func TestIsOneOfValidation(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		validationComment validationComment
		expected          bool
	}{
		{"", false},
		{"a", false},
		{"a|b", true},
		{"^(a|b)$", true},
		{"1-2|2-3", true},
		{"1_2|2_3", true},
		{"1|2", true},
		{"%|#", true},
		{".|b", true},
		{".|.", true},
		{"a|.", true},
		{"a|.", true},
		{"^a|b", false},
		{"a|b$", false},
		{"(a)|b", false},
		{"[a]|b", false},
		{"[a]|b", false},
		{"a+|b", false},
		{"a*|b", false},
		{"[a]*|b", false},
		{"a?|b", false},
		{"\\w|b", false},
		{"{a}|b", false},
		{".{0,32}", false},
	}
	testValidationCommentCheck(t, testCases, func(v validationComment) bool { return v.IsOneOf() })
}

func TestStringLengthValidation(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		validationComment validationComment
		expected          bool
	}{
		{"", false},
		{".{1,2}", true},
		{".{0,9999}", true},
		{".{9999}", true},
		{"{9999}", false},
		{"{9999", false},
		{"9999}", false},
		{".{}", false},
		{".{1,2,3}", false},
		{"a", false},
		{"a,b", false},
		{"1,2", false},
		{"1", false},
		{".{1,b}", false},
		{".{a,2}", false},
		{".{a,b}", false},
		{".{1-2}", false},
		{".{1_2,2_3}", false},
		{".{%,#}", false},
		{".{^1,2}", false},
		{".{1,2$", false},
		{".{(1),2}", false},
		{".{[1],2}", false},
		{".{[1],2}", false},
		{".{1+,2}", false},
		{".{1*,2}", false},
		{".{[1]*,2}", false},
		{".{1?,2}", false},
		{".{\\w,2}", false},
		{".{{1},2}", false},
		{".{.,2}", false},
	}
	testValidationCommentCheck(t, testCases, func(v validationComment) bool { return v.HasDefinedLength() })
}

func TestIsWRegexValidation(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		validationComment validationComment
		expected          bool
	}{
		{"[\\d\\w]+", true},
		{"[\\d\\w]*", true},
		{"[\\w]+", true},
		{"[\\w]*", true},
		{"", false},
		{"a", false},
		{"[\\d]+", false},
		{"[\\d]*", false},
		{"[\\s]+", false},
		{"[\\s]*", false},
	}
	testValidationCommentCheck(t, testCases, func(v validationComment) bool { return v.IsWRegex() })
}

func TestIsMACValidation(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		validationComment validationComment
		expected          bool
	}{
		{"([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})", true},
		{"([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})$", true},
		{"^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})$", true},
		{"^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})$", true},
		{"^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})", true},
		{"^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})", true},
		{"^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})|^$", true},
		{"^$|^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})|^$", true},
		{"^$|^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})", true},
		{"(^$|^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})|^$)", true},

		{"[0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2}", false},
		{"[0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})", false},
		{"([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2}", false},
		{"^[0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2}$", false},
		{"^$|[0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})", false},
		{"^$|[0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})|^$", false},
		{"(^$|[0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})|^$)", false},
		{"^$|[0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})|^$)", false},
		{"(^$|[0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})|^$", false},
		{"", false},
		{"a", false},
		{"([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})|(0-9)", false},
		{"[0-9]|([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})", false},
		{"[0-9]|([0-9A-Fa-f]{2}:){5}", false},
	}
	testValidationCommentCheck(t, testCases, func(v validationComment) bool { return v.IsMAC() })
}

func (vc validationComment) mutate(prefix, suffix string) validationComment {
	return validationComment(fmt.Sprintf("%s%s%s", prefix, string(vc), suffix))
}

func generateTestCasesForFixedRegex(regex string) []struct {
	validationComment validationComment
	expected          bool
} {
	base := validationComment(regex)
	return []struct {
		validationComment validationComment
		expected          bool
	}{
		{base, true},
		{base.mutate("^", "$"), true},
		{base.mutate("^", ""), true},
		{base.mutate("", "$"), true},
		{base.mutate("(^", "$)"), true},
		{base.mutate("^", "$)"), true},
		{base.mutate("^(", "$"), true},
		{base.mutate("^$|", "|^$"), true},
		{base.mutate("^$|", ""), true},
		{base.mutate("", "|^$"), true},
		{base.mutate("(^$|", "|^$)"), true},
		{base.mutate("(^$|", ""), true},
		{base.mutate("", "|^$)"), true},

		// FIXME how to handle these cases? current implementation is dumb and quick, and might fail..
		//{base.mutate("^test", ""), false},
		//{base.mutate("^test", "test$"), false},
		//{base.mutate("", "test$"), false},
		//{base.mutate("test", "test"), false},
		//{base.mutate("", "test"), false},
		//{base.mutate("test", ""), false},
		{base.mutate("^test$|", "|^test$"), false},
		{base.mutate("^test|", "|^test$"), false},
		{base.mutate("test$|", "|^test$"), false},
		{base.mutate("test$|", "|^test$"), false},
		{base.mutate("^test$|", "|test$"), false},
		{base.mutate("^test|", "|^test"), false},
		{base.mutate("test|", "|test"), false},
		{base.mutate("(test|", "|test)"), false},
		{"test", false},
		{"", false},
	}
}

func TestIsIPv4Validation(t *testing.T) {
	t.Parallel()
	testCases := generateTestCasesForFixedRegex(ipv4Regex)
	testValidationCommentCheck(t, testCases, func(v validationComment) bool { return v.IsIPv4() })
}

func TestIsIPv6Validation(t *testing.T) {
	t.Parallel()
	testCases := generateTestCasesForFixedRegex(ipv6Regex)
	testValidationCommentCheck(t, testCases, func(v validationComment) bool { return v.IsIPv6() })
}

func TestIsIPValidation(t *testing.T) {
	t.Parallel()
	testCases := generateTestCasesForFixedRegex(ipv4Regex + "|" + ipv6Regex)
	testCases = append(testCases, generateTestCasesForFixedRegex(ipv6Regex+"|"+ipv4Regex)...)
	testCases = append(testCases, generateTestCasesForFixedRegex("("+ipv6Regex+")|("+ipv4Regex+")")...)
	testCases = append(testCases, generateTestCasesForFixedRegex("("+ipv4Regex+")|("+ipv6Regex+")")...)
	testValidationCommentCheck(t, testCases, func(v validationComment) bool { return v.IsIP() })
}

func TestIsNumericNonZeroValidation(t *testing.T) {
	t.Parallel()
	base := validationComment(numericNonZeroRegex)
	testCases := []struct {
		validationComment validationComment
		expected          bool
	}{
		{base, true},
		{base.mutate("^$|", "|^$"), true},
		{base.mutate("^$|", ""), true},
		{base.mutate("", "|^$"), true},
		{base.mutate("(^$|", "|^$)"), true},
		{base.mutate("(^$|", ""), true},
		{base.mutate("", "|^$)"), true},

		{base.mutate("(^", "$)"), false},
		{base.mutate("^", "$)"), false},
		{base.mutate("^(", "$"), false},
		{base.mutate("^test", ""), false},
		{base.mutate("^test", "test$"), false},
		{base.mutate("", "test$"), false},
		{base.mutate("test", "test"), false},
		{base.mutate("", "test"), false},
		{base.mutate("test", ""), false},
		{base.mutate("^test$|", "|^test$"), false},
		{base.mutate("^test|", "|^test$"), false},
		{base.mutate("test$|", "|^test$"), false},
		{base.mutate("test$|", "|^test$"), false},
		{base.mutate("^test$|", "|test$"), false},
		{base.mutate("^test|", "|^test"), false},
		{base.mutate("test|", "|test"), false},
		{base.mutate("(test|", "|test)"), false},
		{"test", false},
		{"", false},
	}
	testValidationCommentCheck(t, testCases, func(v validationComment) bool { return v.IsNumericNonZeroBased() })
}
