package validators

type test struct {
	String    string      `validate:"notblank"`
	Array     []int       `validate:"notblank"`
	Pointer   *int        `validate:"notblank"`
	Number    int         `validate:"notblank"`
	Interface interface{} `validate:"notblank"`
	Func      func()      `validate:"notblank"`
}
