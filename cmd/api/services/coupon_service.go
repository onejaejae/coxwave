// cmd/api/services/coupon_service.go
package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/coxwave/coupon-system/cmd/api/repositories"
	"github.com/coxwave/coupon-system/cmd/api/services/base"
	couponv1 "github.com/coxwave/coupon-system/gen/proto/coupon/v1"
	"github.com/coxwave/coupon-system/internal/infrastructure/mysql/model"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type CouponService struct {
	*base.TransactionalService
	repo         *repositories.CouponRepository
	campaignRepo *repositories.CampaignRepository
}

func NewCouponService(db *gorm.DB) *CouponService {
	return &CouponService{
		TransactionalService: base.NewTransactionalService(db),
		repo:                 repositories.NewCouponRepository(db),
		campaignRepo:         repositories.NewCampaignRepository(db),
	}
}

func (s *CouponService) IssueCoupon(ctx context.Context, req *couponv1.IssueCouponRequest) (*couponv1.IssueCouponResponse, error) {
	var response *couponv1.IssueCouponResponse

	err := s.WithTransaction(func(tx *gorm.DB) error {
		// 1. 캠페인 존재 여부 확인 및 락 획득
		campaign, err := s.campaignRepo.GetByIDWithLock(ctx, req.CampaignId, tx)
		if err != nil {
			return fmt.Errorf("campaign not found: %w", err)
		}

		// 2. 발급 시작 시간 확인
		if time.Now().Before(campaign.IssueStartAt) {
			return fmt.Errorf("coupon issue not started yet")
		}

		// 3. 발급된 쿠폰 수 확인
		coupons, err := s.repo.GetByCampaignID(ctx, req.CampaignId, tx)
		if err != nil {
			return fmt.Errorf("failed to get coupons: %w", err)
		}

		if len(coupons) >= campaign.CouponQuantity {
			return fmt.Errorf("coupon quantity exceeded")
		}

		// 4. 쿠폰 코드 생성 및 중복 확인
		code := generateCouponCode()
		for {
			exists, err := s.repo.ExistsByCode(ctx, code, tx)
			if err != nil {
				return fmt.Errorf("failed to check coupon code: %w", err)
			}
			if !exists {
				break
			}
			code = generateCouponCode()
		}

		// 5. 쿠폰 생성
		coupon := &model.Coupon{
			Code:       code,
			CampaignID: uint(req.CampaignId),
			IssuedAt:   time.Now(),
		}

		if err := s.repo.Create(ctx, coupon, tx); err != nil {
			return fmt.Errorf("failed to create coupon: %w", err)
		}

		// 응답 생성
		response = &couponv1.IssueCouponResponse{
			Coupon: &couponv1.Coupon{
				Id:         uint32(coupon.ID),
				Code:       coupon.Code,
				CampaignId: uint32(coupon.CampaignID),
				IssuedAt:   timestamppb.New(coupon.IssuedAt),
				CreatedAt:  timestamppb.New(coupon.CreatedAt),
				UpdatedAt:  timestamppb.New(coupon.UpdatedAt),
			},
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

func generateCouponCode() string {
	const (
		koreanChars = "가나다라마바사아자차카타파하거너더러머버서어저처커터퍼허기니디리미비시이지치키티피히구누두루무부수우주추쿠투푸후그느드르므브스으즈츠크트프흐개내대래매배새애재채캐태패해게네데레메베세에제체케테페헤고노도로모보소오조초코토포호구누두루무부수우주추쿠투푸후그느드르므브스으즈츠크트프흐기니디리미비시이지치키티피히"
		nums        = "0123456789"
	)

	koreanRunes := []rune(koreanChars)
	koreanCount := rand.Intn(5) + 1

	code := make([]rune, 10)
	positions := make([]int, 10)

	for i := range positions {
		positions[i] = i
	}

	for i := len(positions) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		positions[i], positions[j] = positions[j], positions[i]
	}

	for i := 0; i < 10; i++ {
		if i < koreanCount {
			code[positions[i]] = koreanRunes[rand.Intn(len(koreanRunes))]
		} else {
			code[positions[i]] = rune(nums[rand.Intn(len(nums))])
		}
	}

	return string(code)
}
