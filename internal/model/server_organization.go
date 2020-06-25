package model

// ServerOrganization ..
type ServerOrganization struct {
	ServerID       uint64 `gorm:"index"`
	OrganizationID uint64 `gorm:"index"`
}
