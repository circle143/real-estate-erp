package payload

import (
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/validation"
	"github.com/go-playground/validator/v10"
)

// package payload handles encoding / decoding incoming http request body
// it also handles validating incoming request

// validatorObj handles validating incoming request body
var validatorObj = validator.New()

type validatorFactory func() common.IValidator

var customValidators = []validatorFactory{
	validation.CreateAadharValidator,
	validation.CreateGSTValidator,
	validation.CreatePassportValidator,
	validation.CreatePanValidator,
}

// RegisterValidators register custom validators
func RegisterValidators() error {
	// add custom validations
	for _, factory := range customValidators {
		v := factory()
		err := validatorObj.RegisterValidation(v.ValidationTag(), v.Validator)
		if err != nil {
			return err
		}
	}

	return nil
}
