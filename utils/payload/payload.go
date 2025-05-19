package payload

import (
	"circledigital.in/real-state-erp/utils/validation"
	"github.com/go-playground/validator/v10"
)

// package payload handles encoding / decoding incoming http request body
// it also handles validating incoming request

// validatorObj handles validating incoming request body
var validatorObj = validator.New()

// RegisterValidators register custom validators
func RegisterValidators() error {
	err := validatorObj.RegisterValidation("gst", validation.GSTValidator)
	if err != nil {
		return err
	}

	err = validatorObj.RegisterValidation("aadhar", validation.AadharValidator)
	if err != nil {
		return err
	}

	err = validatorObj.RegisterValidation("pan", validation.PANValidator)
	if err != nil {
		return err
	}

	err = validatorObj.RegisterValidation("passport", validation.PassportValidator)
	if err != nil {
		return err
	}

	return nil
}
