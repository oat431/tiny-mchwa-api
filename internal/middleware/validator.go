package middleware

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type FiberValidator struct {
	validate *validator.Validate
}

func NewValidator() *FiberValidator {
	return &FiberValidator{validate: validator.New()}
}

func (v *FiberValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

// BadRequest returns a structured 400 error for validation failures.
func BadRequest(c fiber.Ctx, errs error) error {
	details := []string{}
	if validationErrors, ok := errs.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			details = append(details, e.Field()+" failed on '"+e.Tag()+"'")
		}
	} else {
		details = append(details, errs.Error())
	}
	return c.Status(400).JSON(fiber.Map{
		"data":  nil,
		"error": fiber.Map{"code": 400, "message": "Validation failed", "details": details},
		"meta":  nil,
	})
}
