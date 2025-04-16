package payload

import "github.com/go-playground/validator/v10"

// package payload handles encoding / decoding incoming http request body
// it also handles validating incoming request

// validatorObj handles validating incoming request body
var validatorObj = validator.New()