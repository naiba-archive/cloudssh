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

	TeamID uint64
}

// ListServerResponse ..
type ListServerResponse struct {
	Response
	Data []model.Server
}

// ListTeamResponse ..
type ListTeamResponse struct {
	Response
	Data struct {
		Teamnazation []model.Team
		Permission   map[uint64]uint64
	}
}

// ListTeamUserResponse ..
type ListTeamUserResponse struct {
	Response
	Data struct {
		User  []model.TeamUser
		Key   map[uint64]string
		Email map[uint64]string
	}
}

// DeleteTeamRequest ..
type DeleteTeamRequest struct {
	ID []uint
}

// DeleteServerRequest ..
type DeleteServerRequest struct {
	ID     []uint
	TeamID uint64
}

// GetServerResponse ..
type GetServerResponse struct {
	Response
	Data model.Server
}

// TeamRequrest ..
type TeamRequrest struct {
	Name    string
	Pubkey  string
	Servers []model.Server
	Users   []model.TeamUser
}

// NewTeamRequrest ..
type NewTeamRequrest struct {
	Name   string
	Pubkey string
	Prikey string
}

// MyTeamInfo ..
type MyTeamInfo struct {
	Team     model.Team
	TeamUser model.TeamUser
}

// GetTeamResponse ..
type GetTeamResponse struct {
	Response
	Data MyTeamInfo
}

// GetUserTeamInfoResponse ..
type GetUserTeamInfoResponse struct {
	Response
	Data model.TeamUser
}

// AddTeamUserRequest ..
type AddTeamUserRequest struct {
	Permission uint64
	Email      string
	Prikey     string
}

// PasswdRequest ..
type PasswdRequest struct {
	OldPasswordHash string
	PasswordHash    string
	EncryptKey      string
	Pubkey          string
	Privatekey      string
	TeamUser        []model.TeamUser
	Server          []model.Server
}
