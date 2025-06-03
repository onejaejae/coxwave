package model

import (
	"time"

	"gorm.io/gorm"
)

type Coupon struct {
	gorm.Model
	Code       string    `gorm:"type:varchar(10);primary_key"`
	CampaignID uint      `gorm:"not null"`
	IssuedAt   time.Time `gorm:"not null"`
	Campaign   Campaign  `gorm:"foreignKey:CampaignID"`
}
