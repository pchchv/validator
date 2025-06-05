package structLevel

// Address houses a users address information.
type Address struct {
	Street string `validate:"required"`
	City   string `validate:"required"`
	Planet string `validate:"required"`
	Phone  string `validate:"required"`
}

type Gender uint

func (gender Gender) String() string {
	terms := []string{"Male", "Female"}
	if gender < Male || gender > Female {
		return "unknown"
	}
	return terms[gender]
}

const (
	Male Gender = iota + 1
	Female
)
