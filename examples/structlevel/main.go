package structLevel

// Address houses a users address information.
type Address struct {
	Street string `validate:"required"`
	City   string `validate:"required"`
	Planet string `validate:"required"`
	Phone  string `validate:"required"`
}

// User contains user information.
type User struct {
	FirstName      string     `json:"fname"`
	LastName       string     `json:"lname"`
	Age            uint8      `validate:"gte=0,lte=130"`
	Email          string     `json:"e-mail" validate:"required,email"`
	FavouriteColor string     `validate:"hexcolor|rgb|rgba"`
	Addresses      []*Address `validate:"required,dive,required"`
	Gender         Gender     `json:"gender" validate:"required,gender_custom_validation"`
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
