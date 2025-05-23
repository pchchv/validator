package validator

import (
	"context"
	"reflect"
)

var restrictedTags = map[string]struct{}{
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
