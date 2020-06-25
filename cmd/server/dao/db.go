package dao

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// DB ..
var DB *gorm.DB

// InitDB ..
func InitDB(dsn string) error {
	var err error
	DB, err = gorm.Open("mysql", dsn)
	return err
}

// FindIDResp ..
type FindIDResp struct {
	ID []uint64
}
