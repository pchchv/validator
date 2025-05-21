package validator

import "reflect"

// FieldError contains all functions to get error details.
type FieldError interface {
	// Tag returns the validation tag that failed.
	// If the validation was an alias,
	// this will return the alias name and not the underlying tag that failed.
	// For example, alias "iscolor": "hexcolor|rgb|rgba|hsl|hsla" will return "iscolor"
	Tag() string
	// ActualTag returns the validation tag that failed,
	// even if an alias the actual tag within the alias will be returned.
	// If an 'or' validation fails the entire or will be returned.
	// For example, alias "iscolor": "hexcolor|rgb|rgba|hsl|hsla" will return "hexcolor|rgb|rgba|hsl|hsla"
	ActualTag() string
	// Namespace returns the namespace for the field error,
	// with the tag name taking precedence over the field's actual name.
	// For example, JSON name "User.fname"
	//
	// See StructNamespace() for a version that returns actual names.
	//
	// NOTE: this field may be empty when validating a single primitive field using validate.Field(...),
	// as there is no way to extract its name.
	Namespace() string
	// StructNamespace returns the namespace for the field error, with the field's actual name.
	// For example, "User.FirstName" see Namespace for comparison
	//
	// NOTE: this field may be empty when validating a single primitive field using validate.Field(...),
	// as there is no way to extract its name.
	StructNamespace() string
	// Field returns the fields name with the tag name taking precedence over the field's actual name.
	//
	// `RegisterTagNameFunc` must be registered to get tag value.
	// For example, JSON name "fname" see StructField for comparison,
	Field() string
	// StructField returns the field's actual name from the struct, when able to determine.
	// For example, "FirstName" see Field for comparison
	StructField() string
	// Value returns the actual field's value in case needed for creating the error message.
	Value() interface{}
	// Param returns the parameter value as a string for comparison.
	// This will also help when generating an error message.
	Param() string
	// Kind returns the Field's reflect Kind.
	// For example, time.Time's kind is a struct
	Kind() reflect.Kind
	// Type returns the Field's reflect Type.
	// For example, time.Time's type is time.Time
	Type() reflect.Type
	// Error returns the FieldError's message.
	Error() string
}

// fieldError contains a single field's validation error along with other properties that
// may be needed for error message creation it complies with the FieldError interface.
type fieldError struct {
	tag            string
	actualTag      string
	ns             string
	structNs       string
	fieldLen       uint8
	structfieldLen uint8
	value          interface{}
	param          string
	kind           reflect.Kind
	typ            reflect.Type
}
