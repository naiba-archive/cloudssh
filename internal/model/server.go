package model

const (
	// ServerLoginWithAuthorizedKey ...
	ServerLoginWithAuthorizedKey = "1"
	// ServerLoginWithPassword ..
	ServerLoginWithPassword = "2"

	// ServerOwnerTypeUser ..
	ServerOwnerTypeUser = 0
	// ServerOwnerTypeOrganization ..
	ServerOwnerTypeOrganization = 1
)

// Server ..
type Server struct {
	Common

	Name      string
	IP        string
	Port      string
	User      string
	LoginWith string
	Key       string `gorm:"type:text"` // password or authorized key

	OwnerType uint
	OwnerID   uint64
}
