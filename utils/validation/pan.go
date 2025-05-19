package validation

import (
	"circledigital.in/real-state-erp/utils/common"
	"regexp"
)

func CreatePanValidator() common.IValidator {
	return &regexValidator{
		regex:        regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`),
		validatorTag: "pan",
	}
}
