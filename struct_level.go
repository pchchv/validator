package validator

import (
	"context"
	"reflect"
)

// StructLevel contains all the information and helper functions to validate the structure.
type StructLevel interface {
	// Validator returns the main validation object, in case one wants to call validations internally.
	// this is so you don't have to use anonymous functions to get access to the validate instance.
	Validator() *Validate
	// Top returns the top level struct, if any.
	Top() reflect.Value
	// Parent returns the current fields parent struct, if any.
	Parent() reflect.Value
	// Current returns the current struct.
	Current() reflect.Value
	// ExtractType gets the actual underlying type of field value.
	// It will dive into pointers,
	// customTypes and return you the underlying value and its kind.
	ExtractType(field reflect.Value) (value reflect.Value, kind reflect.Kind, nullable bool)
	// ReportError reports an error by simply passing field and tag information.
	//
	// fieldName and structFieldName are appended to the existing namespace that the validator resides on.
	// For example, could pass 'FirstName' or 'Names[0]' depending on the nesting.
	//
	// tag can be an existing validation tag or an arbitrary tag (needs handling).
	ReportError(field interface{}, fieldName, structFieldName string, tag, param string)
	// ReportValidationErrors reports an error just by passing ValidationErrors.
	//
	// relativeNamespace and relativeActualNamespace get appended to the existing namespace that validator is on.
	// For example, could pass 'User.FirstName' or 'Users[0].FirstName' depending on the nesting.
	// Most of the time they will be blank, unless you validate at a level lower the current field depth.
	ReportValidationErrors(relativeNamespace, relativeActualNamespace string, errs ValidationErrors)
}

// StructLevelFunc accepts all values needed for struct level validation.
type StructLevelFunc func(sl StructLevel)

// StructLevelFuncCtx accepts all values needed for
// struct level validation but also allows passing of
// contextual validation information via context.Context.
type StructLevelFuncCtx func(ctx context.Context, sl StructLevel)

// Validator returns the main validation object,
// in case one want to call validations internally.
func (v *validate) Validator() *Validate {
	return v.v
}

// Top returns the top level struct.
//
// This can be the same as the current struct being validated if not is a nested struct.
//
// This is only called when within Struct and Field Level validation and
// should not be relied upon for an accurate value otherwise.
func (v *validate) Top() reflect.Value {
	return v.top
}

// wrapStructLevelFunc wraps normal StructLevelFunc makes it compatible with StructLevelFuncCtx.
func wrapStructLevelFunc(fn StructLevelFunc) StructLevelFuncCtx {
	return func(ctx context.Context, sl StructLevel) {
		fn(sl)
	}
}
