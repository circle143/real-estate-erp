package validation

import (
	"circledigital.in/real-state-erp/utils/common"
	"regexp"
)

func CreatePassportValidator() common.IValidator {
	return &regexValidator{
		regex:        regexp.MustCompile(`^[A-Z0-9]{6,9}$`),
		validatorTag: "passport",
	}
}
