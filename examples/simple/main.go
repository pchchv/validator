package main

import (
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
