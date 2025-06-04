# validator [![GoDoc](https://godoc.org/github.com/pchchv/validator?status.svg)](https://pkg.go.dev/github.com/pchchv/validator) [![Go Report Card](https://goreportcard.com/badge/github.com/pchchv/validator)](https://goreportcard.com/report/github.com/pchchv/validator) [![Coverage Status](https://coveralls.io/repos/pchchv/validator/badge.svg?branch=master&service=github)](https://coveralls.io/github/pchchv/validator?branch=master)

Package `validator` implements structure and field validation, including cross fields, cross structures, maps, slices and arrays.

## Features

- Customizable i18n aware error messages.
- Handles custom field types such as sql driver Valuer.
- Ability to dive into both map keys and values for validation.
- Handles type interface by determining it's underlying type prior to validation.
- Cross Field and Cross Struct validations by using validation tags or custom validators.
- Slice, Array and Map diving, which allows any or all levels of a multidimensional field to be validated.
- Alias validation tags, which allows for mapping of several validations to a single tag for easier defining of validations on structs.
- Extraction of custom defined Field Name e.g. can specify to extract the JSON name while validating and have it available in the resulting FieldError.

## Installation

#### Use go get

```sh
	go get github.com/pchchv/validator
```
And

```go
	import "github.com/pchchv/validator"
```