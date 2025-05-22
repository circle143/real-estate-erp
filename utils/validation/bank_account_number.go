package validation

import (
	"circledigital.in/real-state-erp/utils/common"
	"regexp"
)

func CreateBankAccountNumberValidator() common.IValidator {
	return &regexValidator{
		regex:        regexp.MustCompile("^\\d{9,18}$"),
		validatorTag: "bank-account-number",
	}
}
