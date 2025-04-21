package validators

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"regexp"
)

var (
	TimeFormatRegex = regexp.MustCompile(`^[0-9][0-9]:[0-9][0-9]$`)
	TimeFormat      = stringvalidator.RegexMatches(TimeFormatRegex, "must be a valid time in 24-hour format (HH:MM)")

	DateFormatRegex = regexp.MustCompile(`^(20[0-9]{2})-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$`)
	DateFormat      = stringvalidator.RegexMatches(DateFormatRegex, "must be a valid date in the format YYYY-MM-DD")

	PortRangeRegexp = regexp.MustCompile("(([1-9][0-9]{0,3}|[1-5][0-9]{4}|[6][0-4][0-9]{3}|[6][5][0-4][0-9]{2}|[6][5][5][0-2][0-9]|[6][5][5][3][0-5])|([1-9][0-9]{0,3}|[1-5][0-9]{4}|[6][0-4][0-9]{3}|[6][5][0-4][0-9]{2}|[6][5][5][0-2][0-9]|[6][5][5][3][0-5])-([1-9][0-9]{0,3}|[1-5][0-9]{4}|[6][0-4][0-9]{3}|[6][5][0-4][0-9]{2}|[6][5][5][0-2][0-9]|[6][5][5][3][0-5]))+(,([1-9][0-9]{0,3}|[1-5][0-9]{4}|[6][0-4][0-9]{3}|[6][5][0-4][0-9]{2}|[6][5][5][0-2][0-9]|[6][5][5][3][0-5])|,([1-9][0-9]{0,3}|[1-5][0-9]{4}|[6][0-4][0-9]{3}|[6][5][0-4][0-9]{2}|[6][5][5][0-2][0-9]|[6][5][5][3][0-5])-([1-9][0-9]{0,3}|[1-5][0-9]{4}|[6][0-4][0-9]{3}|[6][5][0-4][0-9]{2}|[6][5][5][0-2][0-9]|[6][5][5][3][0-5])){0,14}")
	PortRangeV2     = validation.StringMatch(PortRangeRegexp, "invalid port range")

	MacRegex = regexp.MustCompile(`^([0-9a-fA-F][0-9a-fA-F][-:]){5}([0-9a-fA-F][0-9a-fA-F])$`)
	Mac      = stringvalidator.RegexMatches(MacRegex, "invalid MAC address")

	HexColorRegex = regexp.MustCompile("^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$")
	HexColor      = stringvalidator.RegexMatches(HexColorRegex, "invalid hex color code")

	EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	Email      = stringvalidator.RegexMatches(EmailRegex, "invalid email address")
)
