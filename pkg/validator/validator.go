package validator

import (
	v10 "github.com/go-playground/validator/v10"
)

// Validator ..
var Validator *v10.Validate

func init() {
	Validator = v10.New()
}
