package validation

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
)

var aadharRegex = regexp.MustCompile(`^[2-9]{1}[0-9]{3}[0-9]{4}[0-9]{4}$`)

func AadharValidator(f1 validator.FieldLevel) bool {
	switch v := f1.Field(); v.Kind() {
	case reflect.String:
		return aadharRegex.MatchString(v.String())
	default:
		return false
	}
}
