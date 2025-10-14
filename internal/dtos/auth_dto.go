package dtos

type RegisterUserDTO struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
}

type VerifyEmailDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Code     string `json:"code" validate:"required,code"`
}

type LoginDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type PasswordRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
}

type PasswordResetDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Code     string `json:"code" validate:"required,code"`
	NewPassword string `json:"new_password" validate:"required,new_password"`
}