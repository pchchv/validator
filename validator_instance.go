package validator

import "reflect"

// TagNameFunc allows for adding of a custom tag name parser.
type TagNameFunc func(field reflect.StructField) string
