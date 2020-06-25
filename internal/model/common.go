package model

import "time"

// Common ..
type Common struct {
	ID        uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
