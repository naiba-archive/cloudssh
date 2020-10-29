package model

const (
	// ServerLoginWithAuthorizedKey ...
	ServerLoginWithAuthorizedKey = "1"
	// ServerLoginWithPassword ..
	ServerLoginWithPassword = "2"

	// ServerOwnerTypeUser ..
	ServerOwnerTypeUser = 0
	// ServerOwnerTypeTeam ..
	ServerOwnerTypeTeam = 1
)

// Server ..
type Server struct {
	Common

	Name      string `gorm:"type:text"` // encrypted
	IP        string `gorm:"type:text"` // encrypted
	Port      string `gorm:"type:text"` // encrypted
	LoginUser string `gorm:"type:text"` // encrypted
	LoginWith string `gorm:"type:text"` // encrypted
	Key       string `gorm:"type:text"` // encrypted

	OwnerType uint
	OwnerID   uint64
}
