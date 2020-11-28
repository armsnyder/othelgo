package messages

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	alphaNumSpacePattern = regexp.MustCompile(`^[A-Za-z0-9 ]*$`)

	// Taken from https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	semVerPattern = regexp.MustCompile(`^v(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)(?:-(?:(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
)

func RegisterCustomValidations(v *validator.Validate) {
	registerRegexpValidation(v, "alphanumspace", alphaNumSpacePattern)
	registerRegexpValidation(v, "semver", semVerPattern)
}

func registerRegexpValidation(v *validator.Validate, tag string, pattern *regexp.Regexp) {
	err := v.RegisterValidation(tag, func(fl validator.FieldLevel) bool {
		return pattern.MatchString(fl.Field().String())
	})
	if err != nil {
		panic(err)
	}
}
