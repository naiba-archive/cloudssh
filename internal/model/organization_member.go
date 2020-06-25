package model

const (
	_ = iota
	// OMPermissionReadOnly ..
	OMPermissionReadOnly
	// OMPermissionReadWrite ..
	OMPermissionReadWrite
)

// OrganizationMember ..
type OrganizationMember struct {
	UserID         uint64 `gorm:"index"`
	OrganizationID uint64 `gorm:"index"`
	Permission     uint64

	PrivateKey string // organization privatekey in user's masterKey encrypted
}
