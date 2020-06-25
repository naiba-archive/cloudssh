package model

// HostOrganization ..
type HostOrganization struct {
	HostID         uint64 `gorm:"index"`
	OrganizationID uint64 `gorm:"index"`
}
