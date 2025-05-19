package validation

import (
	"circledigital.in/real-state-erp/utils/common"
	"regexp"
)

func CreateGSTValidator() common.IValidator {
	return &regexValidator{
		regex:        regexp.MustCompile(`^[0-9]{2}[A-Z]{3}[ABCFGHLJPTF]{1}[A-Z]{1}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`),
		validatorTag: "gst",
	}
}
