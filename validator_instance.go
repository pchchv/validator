package validator

import "reflect"

// TagNameFunc allows for adding of a custom tag name parser.
type TagNameFunc func(field reflect.StructField) string

// CustomTypeFunc overrides or adds custom field type handler functions
// field = field type value to return a value to validate Valuer example from sql drive
// see https://golang.org/src/database/sql/driver/types.go?s=1210:1293#L29
type CustomTypeFunc func(field reflect.Value) interface{}
