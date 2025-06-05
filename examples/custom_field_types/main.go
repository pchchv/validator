package main

import (
	"database/sql/driver"
	"reflect"
)

// ValidateValuer implements validator.CustomTypeFunc.
func ValidateValuer(field reflect.Value) interface{} {
	if valuer, ok := field.Interface().(driver.Valuer); ok {
		if val, err := valuer.Value(); err == nil {
			return val
		}
		// handle the error how you want
	}
	return nil
}
