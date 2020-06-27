package model

const (
	_ = iota
	// OUPermissionReadOnly ..
	OUPermissionReadOnly
	// OUPermissionReadWrite ..
	OUPermissionReadWrite
	// OUPermissionOwner ..
	OUPermissionOwner
)

// OrganizationUser ..
type OrganizationUser struct {
	UserID         uint64 `gorm:"PRIMARY_KEY"`
	OrganizationID uint64 `gorm:"PRIMARY_KEY"`
	Permission     uint64

	PrivateKey string `gorm:"type:text"` // encrypted
}
