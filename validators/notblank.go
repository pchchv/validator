package validators

import (
	"reflect"
	"strings"

	"github.com/pchchv/validator"
)

// NotBlank is a validation function
// to check whether the current field has a
// value or length greater than zero,
// or is not a space only string.
func NotBlank(fl validator.FieldLevel) bool {
	field := fl.Field()
	switch field.Kind() {
	case reflect.String:
		return len(strings.Trim(strings.TrimSpace(field.String()), "\x1c\x1d\x1e\x1f")) > 0
	case reflect.Chan, reflect.Map, reflect.Slice, reflect.Array:
		return field.Len() > 0
	case reflect.Ptr, reflect.Interface, reflect.Func:
		return !field.IsNil()
	default:
		return field.IsValid() && field.Interface() != reflect.Zero(field.Type()).Interface()
	}
}
