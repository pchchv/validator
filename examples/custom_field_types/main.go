package main

import (
	"database/sql"
	"database/sql/driver"
	"reflect"

	"github.com/pchchv/validator"
)

// DbBackedUser User struct
type DbBackedUser struct {
	Name sql.NullString `validate:"required"`
	Age  sql.NullInt64  `validate:"required"`
}

// If a single instance of Validate is used, it caches struct info.
var validate *validator.Validate

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
