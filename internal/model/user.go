package model

import (
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User ..
type User struct {
	Common `json:"common,omitempty"`

	Email        string `json:"email,omitempty" gorm:"type:varchar(100);unique_index"`
	PasswordHash string `json:"-"`
	EncryptKey   string `json:"encrypt_key,omitempty"`
	Pubkey       string `gorm:"type:text" json:"pubkey,omitempty"`
	Privatekey   string `gorm:"type:text" json:"privatekey,omitempty"` // encrypted

	Token        string    `json:"token,omitempty" gorm:"type:varchar(100);unique_index"`
	TokenExpires time.Time `json:"token_expires,omitempty"`
}

// RefreshToken ..
func (u *User) RefreshToken() error {
	token, err := bcrypt.GenerateFromPassword([]byte(fmt.Sprintf("%s-%s-%d", u.Email, u.PasswordHash, rand.Int())), 14)
	if err != nil {
		return err
	}
	u.Token = string(token)
	u.TokenExpires = time.Now().Add(time.Hour * 24 * 7)
	return nil
}
