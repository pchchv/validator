package structLevel

import "github.com/pchchv/validator"

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

// If a single instance of Validate is used, it caches struct info.
var validate *validator.Validate

// UserStructLevelValidation contains custom structure level validations
// that don't always make sense at the field validation level.
// For example, this function validates whether FirstName or LastName exists.
// It could do this with a custom field validation,
// but then would have to add it to both fields, duplicating logic + overhead,
// and this way it only validated once.
func UserStructLevelValidation(sl validator.StructLevel) {
	user := sl.Current().Interface().(User)
	if len(user.FirstName) == 0 && len(user.LastName) == 0 {
		sl.ReportError(user.FirstName, "fname", "FirstName", "fnameorlname", "")
		sl.ReportError(user.LastName, "lname", "LastName", "fnameorlname", "")
	}
	// plus can do more, even with different tag than "fnameorlname"
}
