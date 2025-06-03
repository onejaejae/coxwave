// cmd/api/repositories/coupon_repository.go
package repositories

import (
	"context"

	"github.com/coxwave/coupon-system/internal/infrastructure/mysql/model"
	"gorm.io/gorm"
)

type CouponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

func (r *CouponRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *CouponRepository) Create(ctx context.Context, coupon *model.Coupon, tx *gorm.DB) error {
	return r.getDB(tx).WithContext(ctx).Create(coupon).Error
}

func (r *CouponRepository) GetByCampaignID(ctx context.Context, campaignID uint32, tx *gorm.DB) ([]model.Coupon, error) {
	var coupons []model.Coupon
	if err := r.getDB(tx).WithContext(ctx).Where("campaign_id = ?", campaignID).Find(&coupons).Error; err != nil {
		return nil, err
	}
	return coupons, nil
}

func (r *CouponRepository) ExistsByCode(ctx context.Context, code string, tx *gorm.DB) (bool, error) {
	var count int64
	if err := r.getDB(tx).WithContext(ctx).Model(&model.Coupon{}).Where("code = ?", code).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
