package validation

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
)

// passportRegex is a basic passport regex that just checks length and character set
var passportRegex = regexp.MustCompile(`^[A-Z0-9]{6,9}$`)

func PassportValidator(f1 validator.FieldLevel) bool {
	switch v := f1.Field(); v.Kind() {
	case reflect.String:
		return passportRegex.MatchString(v.String())
	default:
		return false
	}
}
