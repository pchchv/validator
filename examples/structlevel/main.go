package structLevel

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/pchchv/validator"
)

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

type validationError struct {
	Namespace       string `json:"namespace"` // can differ when a custom TagNameFunc is registered or
	Field           string `json:"field"`     // by passing alt name to ReportError like below
	StructNamespace string `json:"structNamespace"`
	StructField     string `json:"structField"`
	Tag             string `json:"tag"`
	ActualTag       string `json:"actualTag"`
	Kind            string `json:"kind"`
	Type            string `json:"type"`
	Value           string `json:"value"`
	Param           string `json:"param"`
	Message         string `json:"message"`
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

func main() {
	validate = validator.New()
	// register function to get tag name from json tags
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		if name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]; name != "-" {
			return name
		}
		return ""
	})

	// register validation for 'User'
	// NOTE: only have to register a non-pointer type for 'User',
	// validator internally dereferences during it's type checks
	validate.RegisterStructValidation(UserStructLevelValidation, User{})
	// register a custom validation for user genre on a line validates that an enum is within the interval
	if err := validate.RegisterValidation("gender_custom_validation", func(fl validator.FieldLevel) bool {
		value := fl.Field().Interface().(Gender)
		return value.String() != "unknown"
	}); err != nil {
		log.Println(err)
		return
	}

	// build 'User' info,
	// normally posted data etc.
	address := &Address{
		Street: "Eavesdown Docks",
		Planet: "Persphone",
		Phone:  "none",
		City:   "Unknown",
	}

	user := &User{
		FirstName:      "",
		LastName:       "",
		Age:            45,
		Email:          "Badger.Smith@gmail",
		FavouriteColor: "#000",
		Addresses:      []*Address{address},
	}

	// returns InvalidValidationError for bad validation input,
	// nil or ValidationErrors ( []FieldError )
	if err := validate.Struct(user); err != nil {
		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this
		var invalidValidationError *validator.InvalidValidationError
		if errors.As(err, &invalidValidationError) {
			log.Println(err)
			return
		}

		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			for _, err := range validateErrs {
				e := validationError{
					Namespace:       err.Namespace(),
					Field:           err.Field(),
					StructNamespace: err.StructNamespace(),
					StructField:     err.StructField(),
					Tag:             err.Tag(),
					ActualTag:       err.ActualTag(),
					Kind:            err.Kind().String(),
					Type:            err.Type().String(),
					Value:           fmt.Sprint(err.Value()),
					Param:           err.Param(),
					Message:         err.Error(),
				}

				if indent, err := json.MarshalIndent(e, "", "  "); err != nil {
					log.Println(err)
					panic(err)
				} else {
					log.Println(string(indent))
				}
			}
		}
		// here it is possible to create custom error messages
		return
	}
	// save user to database
}
