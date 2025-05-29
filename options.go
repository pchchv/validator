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
