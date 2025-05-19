package common

import "github.com/go-playground/validator/v10"

type IValidator interface {
	ValidationTag() string
	Validator(level validator.FieldLevel) bool
}
