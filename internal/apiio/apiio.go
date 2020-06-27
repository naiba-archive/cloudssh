package apiio

import "github.com/naiba/cloudssh/internal/model"

// Response ..
type Response struct {
	Success bool
	Message string
}

// UserInfoResponse ..
type UserInfoResponse struct {
	Response
	Data struct {
		Pubkey string
	}
}

// RegisterRequest ..
type RegisterRequest struct {
	Email        string `validate:"required,email,lowercase"`
	PasswordHash string `validate:"required,min=10"`
	EncryptKey   string `validate:"required,min=10"`

	Privatekey string `validate:"required,min=10"`
	Pubkey     string `valiadte:"required,min=10"`
}

// UserResponse ..
type UserResponse struct {
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

// ListOrganizationResponse ..
type ListOrganizationResponse struct {
	Response
	Data struct {
		Orgnazation []model.Organization
		Permission  map[uint64]uint64
	}
}

// ListOrganizationUserResponse ..
type ListOrganizationUserResponse struct {
	Response
	Data struct {
		User  []model.OrganizationUser
		Key   map[uint64]string
		Email map[uint64]string
	}
}

// DeleteOrganizationRequest ..
type DeleteOrganizationRequest struct {
	ID []uint
}

// DeleteServerRequest ..
type DeleteServerRequest struct {
	ID             []uint
	OrganizationID uint64
}

// GetServerResponse ..
type GetServerResponse struct {
	Response
	Data model.Server
}

// OrgRequrest ..
type OrgRequrest struct {
	Name    string
	Pubkey  string
	Servers []model.Server
	Users   []model.OrganizationUser
}

// NewOrgRequrest ..
type NewOrgRequrest struct {
	Name   string
	Pubkey string
	Prikey string
}

// MyOrganizationInfo ..
type MyOrganizationInfo struct {
	Organization     model.Organization
	OrganizationUser model.OrganizationUser
}

// GetOrganizationResponse ..
type GetOrganizationResponse struct {
	Response
	Data MyOrganizationInfo
}

// GetUserOrganizationInfoResponse ..
type GetUserOrganizationInfoResponse struct {
	Response
	Data model.OrganizationUser
}

// AddOrganizationUserRequest ..
type AddOrganizationUserRequest struct {
	Permission uint64
	Email      string
	Prikey     string
}

// PasswdRequest ..
type PasswdRequest struct {
	OldPasswordHash  string
	PasswordHash     string
	EncryptKey       string
	Pubkey           string
	Privatekey       string
	OrganizationUser []model.OrganizationUser
	Server           []model.Server
}
