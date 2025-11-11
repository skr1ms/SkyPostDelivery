package request

type CreateUser struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type Login struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type VerifyPhone struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required,len=4"`
}

type LoginByPhone struct {
	Phone string `json:"phone" binding:"required"`
}

type RequestPasswordReset struct {
	Phone string `json:"phone" binding:"required"`
}

type ResetPassword struct {
	Phone       string `json:"phone" binding:"required"`
	Code        string `json:"code" binding:"required,len=4"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type RegisterDevice struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform" binding:"required"`
}
