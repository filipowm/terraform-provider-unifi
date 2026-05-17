package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type validator string

type validation struct {
	v      validator
	params []string
}

type validationComment string

type regexSpecialChars string

const (
	validateTag               = "validate"
	mac             validator = "mac"
	ip              validator = "ip"
	ipv4            validator = "ipv4"
	ipv6            validator = "ipv6"
	httpUrl         validator = "http_url"
	oneOf           validator = "oneof"
	cidr            validator = "cidr"
	omitempty       validator = "omitempty"
	length          validator = "len"
	gte             validator = "gte"
	lte             validator = "lte"
	w_regex         validator = "w_regex"
	numeric_nonzero validator = "numeric_nonzero"

	regexChars regexSpecialChars = "^$*+?()[]{}\\|."
)

func createValidations(validations ...validation) string {
	if len(validations) == 0 {
		return ""
	}
	validators := make([]string, len(validations)+1)
	validators[0] = createValidator(omitempty)
	for i, v := range validations {
		validators[i+1] = createValidator(v.v, v.params...)
	}
	joinedValidators := strings.Join(validators, ",")
	return fmt.Sprintf("%s:\"%s\"", validateTag, joinedValidators)
}

func createValidator(v validator, params ...string) string {
	var filteredParams []string
	for _, p := range params {
		if p != "" {
			filteredParams = append(filteredParams, p)
		}
	}
	if len(filteredParams) == 0 {
		return string(v)
	}
	return fmt.Sprintf("%s=%s", v, strings.Join(filteredParams, " "))
}

func (r regexSpecialChars) In(s string, excludedChars string) bool {
	for _, c := range r {
		if strings.ContainsRune(s, c) && !strings.ContainsRune(excludedChars, c) {
			return true
		}
	}
	return false
}

func (r regexSpecialChars) NotIn(s string, excludedChars string) bool {
	return !r.In(s, excludedChars)
}

func (vc validationComment) HasDefinedLength() bool {
	s := string(vc)
	formatOk := strings.HasPrefix(s, ".{") && strings.HasSuffix(s, "}") && regexChars.NotIn(s, ".{}")
	if formatOk {
		sub := s[2 : len(s)-1]
		bounds := strings.Split(sub, ",")
		if len(bounds) < 1 || len(bounds) > 2 {
			return false
		}
		for _, b := range bounds {
			if _, err := strconv.Atoi(b); err != nil {
				return false
			}
		}
		return true
	}
	return false
}

func (vc validationComment) IsOneOf() bool {
	s := string(vc)
	trimmed := strings.TrimPrefix(strings.TrimSuffix(s, ")$"), "^(")
	return strings.Contains(trimmed, "|") && regexChars.NotIn(trimmed, "|.")
}

func (vc validationComment) IsWRegex() bool {
	s := string(vc)
	return slices.Contains([]string{"[\\d\\w]+", "[\\d\\w]*", "[\\w]+", "[\\w]*"}, s)
}

func (vc validationComment) IsMAC() bool {
	s := string(vc)
	// there are validations present in both notations, so we need to check for both
	return (strings.Contains(s, "([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})") || strings.Contains(s, "([0-9A-Fa-f]{2}[:]){5}([0-9A-Fa-f]{2})")) && regexChars.NotIn(s, "(){}[]^$")
}

func (vc validationComment) IsIPv4() bool {
	s := string(vc)
	return strings.Contains(s, ipv4Regex) && strings.Count(s, "|") == ipv4RegexGroupsCount // last is sanity check if there are no more validation groups than expected
}

func (vc validationComment) IsIPv6() bool {
	s := string(vc)
	return strings.Contains(s, ipv6Regex) && strings.Count(s, "|") == ipv6RegexGroupsCount // last is sanity check if there are no more validation groups than expected
}

func (vc validationComment) IsIP() bool {
	s := string(vc)
	return strings.Contains(s, ipv4Regex) && strings.Contains(s, ipv6Regex) && strings.Count(s, "|") == (ipv4RegexGroupsCount+ipv6RegexGroupsCount+1)
}

func (vc validationComment) IsNumericNonZeroBased() bool {
	s := string(vc)
	return s == numericNonZeroRegex
}

func trimWrappers(s string) string {
	trimmed := strings.TrimSuffix(strings.TrimPrefix(s, "(^$|"), "|^$)")    // remove wrapping parenthesis
	trimmed = strings.TrimSuffix(strings.TrimPrefix(trimmed, "^$|"), "|^$") // remove ^$ which allows for empty string and is not needed
	return trimmed
}

const (
	ipv4Regex           = "(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])"
	ipv6Regex           = "(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))"
	numericNonZeroRegex = "^[1-9][0-9]*$"
)

var (
	ipv4RegexGroupsCount = strings.Count(ipv4Regex, "|")
	ipv6RegexGroupsCount = strings.Count(ipv6Regex, "|")
)

func defineFieldValidation(rawValidation string) string {
	if rawValidation == "" {
		return ""
	}
	rawValidation = trimWrappers(rawValidation)
	vc := validationComment(rawValidation)
	if vc.IsOneOf() {
		trimmed := strings.TrimPrefix(strings.TrimSuffix(rawValidation, ")$"), "^(")
		return createValidations(validation{v: oneOf, params: strings.Split(trimmed, "|")})
	} else if vc.HasDefinedLength() {
		sub := rawValidation[2 : len(rawValidation)-1]
		bounds := strings.Split(sub, ",")
		if len(bounds) == 1 {
			return createValidations(validation{v: length, params: []string{bounds[0]}})
		}
		return createValidations(validation{v: gte, params: []string{bounds[0]}}, validation{v: lte, params: []string{bounds[1]}})
	} else if vc.IsWRegex() {
		return createValidations(validation{v: w_regex})
	} else if vc.IsMAC() {
		return createValidations(validation{v: mac})
	} else if vc.IsIPv4() {
		return createValidations(validation{v: ipv4})
	} else if vc.IsIPv6() {
		return createValidations(validation{v: ipv6})
	} else if vc.IsIP() {
		return createValidations(validation{v: ip})
	} else if vc.IsNumericNonZeroBased() {
		return createValidations(validation{v: numeric_nonzero})
	}
	return ""
}
