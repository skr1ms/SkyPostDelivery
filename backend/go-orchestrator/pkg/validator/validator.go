package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

type Validator interface {
	ValidateVar(field any, tag string) error
}

var (
	russianPhoneRegex = regexp.MustCompile(`^(\+7|7|8)?[\s\-]?\(?\d{3}\)?[\s\-]?\d{3}[\s\-]?\d{2}[\s\-]?\d{2}$`)
	emailRegex        = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	passwordRegex     = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]+$`)
)

type CustomValidator struct {
	validator *validator.Validate
}

func New() *CustomValidator {
	v := validator.New()

	_ = v.RegisterValidation("russian_phone", validateRussianPhone)
	_ = v.RegisterValidation("strong_password", validateStrongPassword)
	_ = v.RegisterValidation("custom_email", validateEmail)

	return &CustomValidator{validator: v}
}

func (cv *CustomValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

func (cv *CustomValidator) ValidateVar(field any, tag string) error {
	return cv.validator.Var(field, tag)
}

func ValidateRussianPhoneField(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return false
	}
	return russianPhoneRegex.MatchString(phone)
}

func ValidateStrongPasswordField(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 6 {
		return false
	}
	if len(password) > 128 {
		return false
	}
	return passwordRegex.MatchString(password)
}

func ValidateEmailField(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	if email == "" {
		return false
	}
	return emailRegex.MatchString(email)
}

func validateRussianPhone(fl validator.FieldLevel) bool {
	return ValidateRussianPhoneField(fl)
}

func validateStrongPassword(fl validator.FieldLevel) bool {
	return ValidateStrongPasswordField(fl)
}

func validateEmail(fl validator.FieldLevel) bool {
	return ValidateEmailField(fl)
}

func NormalizeRussianPhone(phone string) string {
	re := regexp.MustCompile(`[\s\-\(\)]`)
	phone = re.ReplaceAllString(phone, "")

	if len(phone) == 11 && (phone[0] == '8' || phone[0] == '7') {
		return "+7" + phone[1:]
	}
	if len(phone) == 10 {
		return "+7" + phone
	}
	if len(phone) == 12 && phone[:2] == "+7" {
		return phone
	}

	return phone
}

func ValidatePassword(password string) bool {
	if len(password) < 6 || len(password) > 128 {
		return false
	}
	return passwordRegex.MatchString(password)
}

func ValidateRussianPhone(phone string) bool {
	return russianPhoneRegex.MatchString(phone)
}

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}
