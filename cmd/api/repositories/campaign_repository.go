package repositories

import (
	"context"

	"github.com/coxwave/coupon-system/internal/infrastructure/mysql/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CampaignRepository struct {
	db *gorm.DB
}

func NewCampaignRepository(db *gorm.DB) *CampaignRepository {
	return &CampaignRepository{db: db}
}

func (r *CampaignRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *CampaignRepository) Create(ctx context.Context, campaign *model.Campaign, tx *gorm.DB) error {
	return r.getDB(tx).WithContext(ctx).Create(campaign).Error
}

func (r *CampaignRepository) GetByIDWithLock(ctx context.Context, id uint32, tx *gorm.DB) (*model.Campaign, error) {
	var campaign model.Campaign
	if err := r.getDB(tx).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&campaign, id).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (r *CampaignRepository) GetByIDWithCouponCodes(ctx context.Context, id uint32, tx *gorm.DB) (*model.Campaign, error) {
	var campaign model.Campaign
	if err := r.getDB(tx).WithContext(ctx).
		Preload("Coupons").
		First(&campaign, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (r *CampaignRepository) GetByID(ctx context.Context, id uint32, tx *gorm.DB) (*model.Campaign, error) {
	var campaign model.Campaign
	if err := r.getDB(tx).WithContext(ctx).
		First(&campaign, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}
