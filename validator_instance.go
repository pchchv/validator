package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

const (
	utf8HexComma       = "0x2C"
	utf8Pipe           = "0x7C"
	structOnlyTag      = "structonly"
	noStructLevelTag   = "nostructlevel"
	omitzero           = "omitzero"
	omitempty          = "omitempty"
	omitnil            = "omitnil"
	isdefault          = "isdefault"
	skipValidationTag  = "-"
	diveTag            = "dive"
	keysTag            = "keys"
	endKeysTag         = "endkeys"
	requiredTag        = "required"
	restrictedTagChars = ".[],|=+()`~!@#$%^&*\\\"/?<>{}"
	restrictedAliasErr = "Alias '%s' either contains restricted characters or is the same as a restricted tag needed for normal operation"
	restrictedTagErr   = "Tag '%s' either contains restricted characters or is the same as a restricted tag needed for normal operation"
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

// RegisterAlias registers a mapping of a single validation tag that defines a
// common or complex set of validation(s) to simplify adding validations to structures.
//
// NOTE: this method is not thread-safe it is intended that these all be registered prior to any validation.
func (v *Validate) RegisterAlias(alias, tags string) {
	if _, ok := restrictedTags[alias]; ok || strings.ContainsAny(alias, restrictedTagChars) {
		panic(fmt.Sprintf(restrictedAliasErr, alias))
	}

	v.aliases[alias] = tags
}

// RegisterValidation adds a validation with the given tag.
//
// NOTES:
// If the key already exists, the previous validation function will be replaced.
// This method is not thread-safe it is intended that these all be registered prior to any validation.
func (v *Validate) RegisterValidation(tag string, fn Func, callValidationEvenIfNull ...bool) error {
	return v.RegisterValidationCtx(tag, wrapFunc(fn), callValidationEvenIfNull...)
}

// RegisterValidationCtx does the same as RegisterValidation on accepts a
// FuncCtx validation allowing context.Context validation support.
func (v *Validate) RegisterValidationCtx(tag string, fn FuncCtx, callValidationEvenIfNull ...bool) error {
	var nilCheckable bool
	if len(callValidationEvenIfNull) > 0 {
		nilCheckable = callValidationEvenIfNull[0]
	}

	return v.registerValidation(tag, fn, false, nilCheckable)
}

func (v *Validate) registerValidation(tag string, fn FuncCtx, bakedIn bool, nilCheckable bool) error {
	if len(tag) == 0 {
		return errors.New("function Key cannot be empty")
	}

	if fn == nil {
		return errors.New("function cannot be empty")
	}

	if _, ok := restrictedTags[tag]; !bakedIn && (ok || strings.ContainsAny(tag, restrictedTagChars)) {
		panic(fmt.Sprintf(restrictedTagErr, tag))
	}

	v.validations[tag] = internalValidationFuncWrapper{fn: fn, runValidationOnNil: nilCheckable}
	return nil
}

type internalValidationFuncWrapper struct {
	fn                 FuncCtx
	runValidationOnNil bool
}
