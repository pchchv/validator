package validator

// Option represents a configurations option to
// be applied to validator during initialization.
type Option func(*Validate)

// WithPrivateFieldValidation activates validation for
// unexported fields using of the `unsafe` package.
//
// WARNING: By using this feature,
// you acknowledge that you are aware of the risks and accept any
// current or future consequences of using this feature.
func WithPrivateFieldValidation() Option {
	return func(v *Validate) {
		v.privateFieldValidation = true
	}
}

// WithRequiredStructEnabled enables required tag on
// non-pointer structs to be applied instead of ignored.
//
// This was made opt-in behaviour in order
// to maintain backward compatibility with the
// behaviour previous to being able
// to apply struct level validations on struct fields directly.
//
// NOTE: It is recommended to enabled this.
func WithRequiredStructEnabled() Option {
	return func(v *Validate) {
		v.requiredStructEnabled = true
	}
}
