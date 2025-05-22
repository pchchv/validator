package validator

import "context"

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
