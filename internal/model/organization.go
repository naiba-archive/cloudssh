package model

// Organization ..
type Organization struct {
	Common

	UserID uint64
	Name   string
	Pubkey string
}
