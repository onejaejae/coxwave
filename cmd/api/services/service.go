package services

import "gorm.io/gorm"

type Service struct {
	CampaignService *CampaignService
	CouponService   *CouponService
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		CampaignService: NewCampaignService(db),
		CouponService:   NewCouponService(db),
	}
}
