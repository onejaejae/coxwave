package model

import (
	"time"

	"gorm.io/gorm"
)

type Campaign struct {
	gorm.Model
	Name           string    `gorm:"type:varchar(255);not null"`
	CouponQuantity int       `gorm:"not null"`
	IssueStartAt   time.Time `gorm:"not null"`
}
