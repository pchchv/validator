package validator

import (
	"reflect"
	"sync"
)

// TagNameFunc allows for adding of a custom tag name parser.
type TagNameFunc func(field reflect.StructField) string

// CustomTypeFunc overrides or adds custom field type handler functions
// field = field type value to return a value to validate Valuer example from sql drive
// see https://golang.org/src/database/sql/driver/types.go?s=1210:1293#L29
type CustomTypeFunc func(field reflect.Value) interface{}

// FilterFunc is the type used to filter fields using the StructFiltered(...) function.
// Return true causes the field to be filtered/skipped on validation.
type FilterFunc func(ns []byte) bool

// Validate contains the validator settings and cache.
type Validate struct {
	tagName                string
	pool                   *sync.Pool
	tagNameFunc            TagNameFunc
	structLevelFuncs       map[reflect.Type]StructLevelFuncCtx
	customFuncs            map[reflect.Type]CustomTypeFunc
	aliases                map[string]string
	validations            map[string]internalValidationFuncWrapper
	rules                  map[reflect.Type]map[string]string
	tagCache               *tagCache
	structCache            *structCache
	hasCustomFuncs         bool
	hasTagNameFunc         bool
	requiredStructEnabled  bool
	privateFieldValidation bool
}

type internalValidationFuncWrapper struct {
	fn                 FuncCtx
	runValidationOnNil bool
}
