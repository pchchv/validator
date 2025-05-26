package validator

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var (
	oneofValsCache       = map[string][]string{}
	oneofValsCacheRWLock = sync.RWMutex{}
	restrictedTags       = map[string]struct{}{
		diveTag:           {},
		keysTag:           {},
		endKeysTag:        {},
		structOnlyTag:     {},
		omitzero:          {},
		omitempty:         {},
		omitnil:           {},
		skipValidationTag: {},
		utf8HexComma:      {},
		utf8Pipe:          {},
		noStructLevelTag:  {},
		requiredTag:       {},
		isdefault:         {},
	}
	// bakedInAliases is a default mapping of a single validation tag that
	// defines a common or complex set of validation(s) to simplify adding validation to structs
	bakedInAliases = map[string]string{
		"iscolor":         "hexcolor|rgb|rgba|hsl|hsla",
		"country_code":    "iso3166_1_alpha2|iso3166_1_alpha3|iso3166_1_alpha_numeric",
		"eu_country_code": "iso3166_1_alpha2_eu|iso3166_1_alpha3_eu|iso3166_1_alpha_numeric_eu",
	}
)

// Func accepts a FieldLevel interface for all validation needs.
// Return value should be true when validation succeeds.
type Func func(fl FieldLevel) bool

// FuncCtx accepts a context.Context and FieldLevel interface for all validation needs.
// The return value should be true when validation succeeds.
type FuncCtx func(ctx context.Context, fl FieldLevel) bool

// wrapFunc wraps normal Func makes it compatible with FuncCtx.
func wrapFunc(fn Func) FuncCtx {
	if fn == nil {
		return nil
	}

	return func(ctx context.Context, fl FieldLevel) bool {
		return fn(fl)
	}
}

// hasMultiByteCharacter is the validation function for validating if the
// field's value has a multi byte character.
func hasMultiByteCharacter(fl FieldLevel) bool {
	field := fl.Field()
	if field.Len() == 0 {
		return true
	}

	return multibyteRegex().MatchString(field.String())
}

func parseOneOfParam(s string) []string {
	oneofValsCacheRWLock.RLock()
	vals, ok := oneofValsCache[s]
	oneofValsCacheRWLock.RUnlock()
	if !ok {
		oneofValsCacheRWLock.Lock()
		vals = splitParamsRegex().FindAllString(s, -1)
		for i := 0; i < len(vals); i++ {
			vals[i] = strings.ReplaceAll(vals[i], "'", "")
		}

		oneofValsCache[s] = vals
		oneofValsCacheRWLock.Unlock()
	}

	return vals
}

// requireCheckFieldValue is a func for check field value
func requireCheckFieldValue(fl FieldLevel, param, value string, defaultNotFoundValue bool) bool {
	field, kind, _, found := fl.GetStructFieldOKAdvanced(fl.Parent(), param)
	if !found {
		return defaultNotFoundValue
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() == asInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return field.Uint() == asUint(value)
	case reflect.Float32:
		return field.Float() == asFloat32(value)
	case reflect.Float64:
		return field.Float() == asFloat64(value)
	case reflect.Slice, reflect.Map, reflect.Array:
		return int64(field.Len()) == asInt(value)
	case reflect.Bool:
		return field.Bool() == (value == "true")
	case reflect.Ptr:
		if field.IsNil() {
			return value == "nil"
		}

		// handle non-nil pointers
		return requireCheckFieldValue(fl, param, value, defaultNotFoundValue)
	}

	// default reflect.String:
	return field.String() == value
}

// requiredIf is the validation function.
// The field under validation must be present and not empty only if all the
// other specified fields are equal to the value following with the specified field.
func requiredIf(fl FieldLevel) bool {
	params := parseOneOfParam(fl.Param())
	if len(params)%2 != 0 {
		panic(fmt.Sprintf("Bad param number for required_if %s", fl.FieldName()))
	}

	for i := 0; i < len(params); i += 2 {
		if !requireCheckFieldValue(fl, params[i], params[i+1], false) {
			return true
		}
	}

	return hasValue(fl)
}

// requireCheckFieldKind is a func for check field kind.
func requireCheckFieldKind(fl FieldLevel, param string, defaultNotFoundValue bool) bool {
	var nullable, found bool
	field := fl.Field()
	kind := field.Kind()
	if len(param) > 0 {
		field, kind, nullable, found = fl.GetStructFieldOKAdvanced(fl.Parent(), param)
		if !found {
			return defaultNotFoundValue
		}
	}

	switch kind {
	case reflect.Invalid:
		return defaultNotFoundValue
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return field.IsNil()
	default:
		if nullable && field.Interface() != nil {
			return false
		}

		return field.IsValid() && field.IsZero()
	}
}

// requiredUnless is the validation function.
// The field under validation must be present and not empty only unless all the
// other specified fields are equal to the value following with the specified field.
func requiredUnless(fl FieldLevel) bool {
	params := parseOneOfParam(fl.Param())
	if len(params)%2 != 0 {
		panic(fmt.Sprintf("Bad param number for required_unless %s", fl.FieldName()))
	}

	for i := 0; i < len(params); i += 2 {
		if requireCheckFieldValue(fl, params[i], params[i+1], false) {
			return true
		}
	}

	return hasValue(fl)
}

// requiredWith is the validation function.
// The field under validation must be present and not empty only if any of the
// other specified fields are present.
func requiredWith(fl FieldLevel) bool {
	params := parseOneOfParam(fl.Param())
	for _, param := range params {
		if !requireCheckFieldKind(fl, param, true) {
			return hasValue(fl)
		}
	}
	return true
}

// hasValue is the validation function for validating if the current field's value is not the default static value.
func hasValue(fl FieldLevel) bool {
	field := fl.Field()
	switch field.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return !field.IsNil()
	default:
		if fl.(*validate).fldIsPointer && field.Interface() != nil {
			return true
		}
		return field.IsValid() && !field.IsZero()
	}
}

// hasNotZeroValue is the validation function for validating if the current field's value is not the zero value for its type.
func hasNotZeroValue(fl FieldLevel) bool {
	field := fl.Field()
	switch field.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return !field.IsNil()
	default:
		if fl.(*validate).fldIsPointer && field.Interface() != nil {
			return !field.IsZero()
		}
		return field.IsValid() && !field.IsZero()
	}
}
