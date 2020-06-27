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
	UserID         uint64 `gorm:"PRIMARY_KEY;AUTO_INCREMENT:false"`
	OrganizationID uint64 `gorm:"PRIMARY_KEY;AUTO_INCREMENT:false"`
	Permission     uint64

	PrivateKey string `gorm:"type:text"` // encrypted
}

// GetPermissionComment ..
func GetPermissionComment(permission uint64) string {
	switch permission {
	case OUPermissionOwner:
		return "Owner"
	case OUPermissionReadOnly:
		return "ReadOnly"
	case OUPermissionReadWrite:
		return "ReadWrite"
	default:
		return "Unknown"
	}
}
