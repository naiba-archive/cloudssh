package model

const (
	// HostLoginWithAuthorizedKey ...
	HostLoginWithAuthorizedKey = "1"
	// HostLoginWithPassword ..
	HostLoginWithPassword = "2"

	// HostOwnerTypeUser ..
	HostOwnerTypeUser = 0
	// HostOwnerTypeOrganization ..
	HostOwnerTypeOrganization = 1
)

// Host ..
type Host struct {
	Common

	Name      string
	IP        string
	Port      string
	User      string
	LoginWith string
	Key       string // password or authorized key

	OwnerType uint
	OwnerID   uint64
}
