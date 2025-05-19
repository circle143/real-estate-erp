package validation

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
)

var gstRegex = regexp.MustCompile(`^[0-9]{2}[A-Z]{3}[ABCFGHLJPTF]{1}[A-Z]{1}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`)

func GSTValidator(f1 validator.FieldLevel) bool {
	switch v := f1.Field(); v.Kind() {
	case reflect.String:
		return gstRegex.MatchString(v.String())
	default:
		return false
	}
}
