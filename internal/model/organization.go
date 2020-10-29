package model

// Team ..
type Team struct {
	Common

	Name   string `gorm:"type:text"` // pubkey encrypted
	Pubkey string `gorm:"type:text"`
}
