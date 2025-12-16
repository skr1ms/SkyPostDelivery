package request

type CreateUser struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,custom_email"`
	Phone    string `json:"phone" binding:"required,russian_phone"`
	Password string `json:"password" binding:"required,strong_password"`
}

type Login struct {
	Email    string `json:"email" binding:"required,custom_email"`
	Password string `json:"password" binding:"required,strong_password"`
}

type VerifyPhone struct {
	Phone string `json:"phone" binding:"required,russian_phone"`
	Code  string `json:"code" binding:"required,len=4"`
}

type LoginByPhone struct {
	Phone string `json:"phone" binding:"required,russian_phone"`
}

type RequestPasswordReset struct {
	Phone string `json:"phone" binding:"required,russian_phone"`
}

type ResetPassword struct {
	Phone       string `json:"phone" binding:"required,russian_phone"`
	Code        string `json:"code" binding:"required,len=4"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type RegisterDevice struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform" binding:"required"`
}
