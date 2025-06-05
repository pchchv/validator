package main

import (
	"database/sql"
	"database/sql/driver"
	"log"
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

func main() {
	validate = validator.New()
	// register all sql.Null* types to use the ValidateValuer CustomTypeFunc
	validate.RegisterCustomTypeFunc(ValidateValuer, sql.NullString{}, sql.NullInt64{}, sql.NullBool{}, sql.NullFloat64{})
	// build object for validation
	x := DbBackedUser{Name: sql.NullString{String: "", Valid: true}, Age: sql.NullInt64{Int64: 0, Valid: false}}
	if err := validate.Struct(x); err != nil {
		log.Printf("Err(s):\n%+v\n", err)
	}
}
