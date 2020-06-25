package apiio

import "github.com/naiba/cloudssh/internal/model"

// Response ..
type Response struct {
	Success bool
	Message string
}

// RegisterRequest ..
type RegisterRequest struct {
	Email        string `validate:"required,email,lowercase"`
	PasswordHash string `validate:"required,min=10"`
	EncryptKey   string `validate:"required,min=10"`

	Privatekey string `validate:"required,min=10"`
	Pubkey     string `valiadte:"required,min=10"`
}

// RegisterResponse ..
type RegisterResponse struct {
	Response
	Data model.User
}

// LoginRequest ..
type LoginRequest struct {
	Email        string `validate:"required,email,lowercase"`
	PasswordHash string `validate:"required,min=10"`
}
