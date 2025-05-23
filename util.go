package validator

import "reflect"

// extractTypeInternal gets the actual underlying type of field value.
// It will dive into pointers, customTypes and return you the underlying value and it's kind.
func (v *validate) extractTypeInternal(current reflect.Value, nullable bool) (reflect.Value, reflect.Kind, bool) {
BEGIN:
	switch current.Kind() {
	case reflect.Ptr:
		nullable = true
		if current.IsNil() {
			return current, reflect.Ptr, nullable
		}

		current = current.Elem()
		goto BEGIN
	case reflect.Interface:
		nullable = true
		if current.IsNil() {
			return current, reflect.Interface, nullable
		}

		current = current.Elem()
		goto BEGIN
	case reflect.Invalid:
		return current, reflect.Invalid, nullable
	default:
		if v.v.hasCustomFuncs {
			if fn, ok := v.v.customFuncs[current.Type()]; ok {
				current = reflect.ValueOf(fn(current))
				goto BEGIN
			}
		}
		return current, current.Kind(), nullable
	}
}
