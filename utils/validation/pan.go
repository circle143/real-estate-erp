package validation

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
)

var panRegex = regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`)

func PANValidator(f1 validator.FieldLevel) bool {
	switch v := f1.Field(); v.Kind() {
	case reflect.String:
		return panRegex.MatchString(v.String())
	default:
		return false
	}
}
