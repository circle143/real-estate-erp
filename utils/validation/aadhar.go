package validation

import (
	"circledigital.in/real-state-erp/utils/common"
	"regexp"
)

func CreateAadharValidator() common.IValidator {
	return &regexValidator{
		regex:        regexp.MustCompile(`^[2-9]{1}[0-9]{3}[0-9]{4}[0-9]{4}$`),
		validatorTag: "aadhar",
	}
}
