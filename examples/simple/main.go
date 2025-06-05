package main

import (
	"errors"
	"log"

	"github.com/pchchv/validator"
)

// Address houses a users address information.
type Address struct {
	Street string `validate:"required"`
	City   string `validate:"required"`
	Planet string `validate:"required"`
	Phone  string `validate:"required"`
}

// User contains user information
type User struct {
	FirstName      string     `validate:"required"`
	LastName       string     `validate:"required"`
	Age            uint8      `validate:"gte=0,lte=130"`
	Email          string     `validate:"required,email"`
	Gender         string     `validate:"oneof=male female prefer_not_to"`
	FavouriteColor string     `validate:"iscolor"`                // alias for 'hexcolor|rgb|rgba|hsl|hsla'
	Addresses      []*Address `validate:"required,dive,required"` // a person can have a home and cottage...
}

// If a single instance of Validate is used, it caches struct info.
var validate *validator.Validate

func validateVariable() {
	myEmail := "ipchchv.gmail.com"
	if errs := validate.Var(myEmail, "required,email"); errs != nil {
		log.Println(errs) // output: Key: "" Error:Field validation for "" failed on the "email" tag
		return
	}
	// email ok, move on
}

func validateStruct() {
	address := &Address{
		Street: "Eavesdown Docks",
		Planet: "Persphone",
		Phone:  "none",
	}

	user := &User{
		FirstName:      "Jack",
		LastName:       "Pochechuev",
		Age:            31,
		Gender:         "male",
		Email:          "ipchchv@gmail.com",
		FavouriteColor: "#ff0000",
		Addresses:      []*Address{address},
	}

	// returns nil or ValidationErrors ( []FieldError )
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
			for _, e := range validateErrs {
				log.Println(e.Namespace())
				log.Println(e.Field())
				log.Println(e.StructNamespace())
				log.Println(e.StructField())
				log.Println(e.Tag())
				log.Println(e.ActualTag())
				log.Println(e.Kind())
				log.Println(e.Type())
				log.Println(e.Value())
				log.Println(e.Param())
				log.Println()
			}
		}
		// here it is possible to create custom error messages
		return
	}
	// save user to database
}

func main() {
	validate = validator.New(validator.WithRequiredStructEnabled())
	validateStruct()
	validateVariable()
}
