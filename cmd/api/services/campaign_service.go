package services

import (
	"context"

	"github.com/coxwave/coupon-system/cmd/api/repositories"
	campaignv1 "github.com/coxwave/coupon-system/gen/proto/campaign/v1"
	"github.com/coxwave/coupon-system/internal/infrastructure/mysql/model"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type CampaignService struct {
	repo *repositories.CampaignRepository
}

func NewCampaignService(db *gorm.DB) *CampaignService {
	repo := repositories.NewCampaignRepository(db)
	return &CampaignService{repo: repo}
}

func (s *CampaignService) CreateCampaign(ctx context.Context, req *campaignv1.CreateCampaignRequest) (*campaignv1.CreateCampaignResponse, error) {
	campaign := &model.Campaign{
		Name:           req.Name,
		CouponQuantity: int(req.CouponQuantity),
		IssueStartAt:   req.IssueStartAt.AsTime(),
	}

	if err := s.repo.Create(ctx, campaign, nil); err != nil {
		return nil, err
	}

	return &campaignv1.CreateCampaignResponse{
		Campaign: &campaignv1.Campaign{
			Id:             uint32(campaign.ID),
			Name:           campaign.Name,
			CouponQuantity: int32(campaign.CouponQuantity),
			IssueStartAt:   timestamppb.New(campaign.IssueStartAt),
			CreatedAt:      timestamppb.New(campaign.CreatedAt),
			UpdatedAt:      timestamppb.New(campaign.UpdatedAt),
		},
	}, nil
}

func (s *CampaignService) GetCampaign(ctx context.Context, req *campaignv1.GetCampaignRequest) (*campaignv1.GetCampaignResponse, error) {
	campaign, err := s.repo.GetByIDWithCouponCodes(ctx, req.CampaignId, nil)
	if err != nil {
		return nil, err
	}

	couponCodes := make([]string, len(campaign.Coupons))
	for i, coupon := range campaign.Coupons {
		couponCodes[i] = coupon.Code
	}

	return &campaignv1.GetCampaignResponse{
		Campaign: &campaignv1.Campaign{
			Id:             uint32(campaign.ID),
			Name:           campaign.Name,
			CouponQuantity: int32(campaign.CouponQuantity),
			IssueStartAt:   timestamppb.New(campaign.IssueStartAt),
			CreatedAt:      timestamppb.New(campaign.CreatedAt),
			UpdatedAt:      timestamppb.New(campaign.UpdatedAt),
		},
		CouponCodes: couponCodes,
	}, nil
}
