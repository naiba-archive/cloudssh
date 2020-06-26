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

// ServerRequest ..
type ServerRequest struct {
	Name      string `validate:"required,min=10"`
	IP        string `validate:"required,min=10"`
	Port      string `validate:"required,min=10"`
	LoginUser string `validate:"required,min=10"`
	LoginWith string `validate:"required,min=10"`
	Key       string `validate:"required,min=10"`

	OrganizationID uint64
}

// ListServerResponse ..
type ListServerResponse struct {
	Response
	Data []model.Server
}

// DeleteServerRequest ..
type DeleteServerRequest struct {
	ID []uint
}

// GetServerResponse ..
type GetServerResponse struct {
	Response
	Data model.Server
}
