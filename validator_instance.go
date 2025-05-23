package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	utf8HexComma       = "0x2C"
	utf8Pipe           = "0x7C"
	tagSeparator       = ","
	orSeparator        = "|"
	tagKeySeparator    = "="
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

var (
	timeDurationType = reflect.TypeOf(time.Duration(0))
	timeType         = reflect.TypeOf(time.Time{})
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

// RegisterStructValidation registers a StructLevelFunc against a number of types.
//
// NOTE: this method is not thread-safe it is intended that these all be registered prior to any validation.
func (v *Validate) RegisterStructValidation(fn StructLevelFunc, types ...interface{}) {
	v.RegisterStructValidationCtx(wrapStructLevelFunc(fn), types...)
}

// RegisterStructValidationCtx registers a StructLevelFuncCtx against a number of
// types and allows passing of contextual validation information via context.Context.
//
// NOTE: this method is not thread-safe it is intended that these all be registered prior to any validation.
func (v *Validate) RegisterStructValidationCtx(fn StructLevelFuncCtx, types ...interface{}) {
	if v.structLevelFuncs == nil {
		v.structLevelFuncs = make(map[reflect.Type]StructLevelFuncCtx)
	}

	for _, t := range types {
		tv := reflect.ValueOf(t)
		if tv.Kind() == reflect.Ptr {
			t = reflect.Indirect(tv).Interface()
		}

		v.structLevelFuncs[reflect.TypeOf(t)] = fn
	}
}

// RegisterStructValidationMapRules registers validate map rules.
// Be aware that map validation rules supersede those defined on a/the struct if present.
//
// NOTE: this method is not thread-safe it is intended that these all be registered prior to any validation
func (v *Validate) RegisterStructValidationMapRules(rules map[string]string, types ...interface{}) {
	if v.rules == nil {
		v.rules = make(map[reflect.Type]map[string]string)
	}

	deepCopyRules := make(map[string]string)
	for i, rule := range rules {
		deepCopyRules[i] = rule
	}

	for _, t := range types {
		typ := reflect.TypeOf(t)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		if typ.Kind() == reflect.Struct {
			v.rules[typ] = deepCopyRules
		}

	}
}

// RegisterTagNameFunc registers a function to get alternate names for StructFields.
// For example, to use the names which have been specified for JSON representations of structs,
// rather than normal Go field names:
//
//	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
//	    name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
//	    // skip if tag key says it should be ignored
//	    if name == "-" {
//	        return ""
//	    }
//	    return name
//	})
func (v *Validate) RegisterTagNameFunc(fn TagNameFunc) {
	v.tagNameFunc = fn
	v.hasTagNameFunc = true
}

// RegisterCustomTypeFunc registers a CustomTypeFunc against a number of types.
//
// NOTE: this method is not thread-safe it is intended that these all be registered prior to any validation.
func (v *Validate) RegisterCustomTypeFunc(fn CustomTypeFunc, types ...interface{}) {
	if v.customFuncs == nil {
		v.customFuncs = make(map[reflect.Type]CustomTypeFunc)
	}

	for _, t := range types {
		v.customFuncs[reflect.TypeOf(t)] = fn
	}

	v.hasCustomFuncs = true
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
