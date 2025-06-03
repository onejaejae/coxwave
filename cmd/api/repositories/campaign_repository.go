package repositories

import (
	"context"

	"github.com/coxwave/coupon-system/internal/infrastructure/mysql/model"
	"gorm.io/gorm"
)

type CampaignRepository struct {
	db *gorm.DB
}

func NewCampaignRepository(db *gorm.DB) *CampaignRepository {
	return &CampaignRepository{db: db}
}

func (r *CampaignRepository) Create(ctx context.Context, campaign *model.Campaign) error {
	return r.db.WithContext(ctx).Create(campaign).Error
}

func (r *CampaignRepository) GetByIDWithCouponCodes(ctx context.Context, id uint32) (*model.Campaign, error) {
	var campaign model.Campaign
	if err := r.db.WithContext(ctx).
		Preload("Coupons").
		First(&campaign, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (r *CampaignRepository) GetByID(ctx context.Context, id uint32) (*model.Campaign, error) {
	var campaign model.Campaign
	if err := r.db.WithContext(ctx).
		First(&campaign, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}
