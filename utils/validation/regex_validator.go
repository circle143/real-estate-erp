package validation

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
)

type regexValidator struct {
	validatorTag string
	regex        *regexp.Regexp
}

func (v *regexValidator) ValidationTag() string {
	return v.validatorTag
}

func (v *regexValidator) Validator(field validator.FieldLevel) bool {
	switch val := field.Field(); val.Kind() {
	case reflect.String:
		return v.regex.MatchString(val.String())
	default:
		return false
	}
}
