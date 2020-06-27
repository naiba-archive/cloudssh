package model

// Organization ..
type Organization struct {
	Common

	Name   string `gorm:"type:text"` // pubkey encrypted
	Pubkey string `gorm:"type:text"`
}
