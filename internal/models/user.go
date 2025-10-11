package models

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	CreatedAt string `json:"created_at"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	User      UserResponse `json:"user"`
	ExpiresAt string       `json:"expires_at"`
}