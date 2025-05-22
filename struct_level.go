package validator

import "reflect"

// StructLevel contains all the information and helper functions to validate the structure.
type StructLevel interface {
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
