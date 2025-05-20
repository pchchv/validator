package validator

// Func accepts a FieldLevel interface for all validation needs.
// Return value should be true when validation succeeds.
type Func func(fl FieldLevel) bool
