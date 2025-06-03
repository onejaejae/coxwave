package handlers

import (
	"context"

	"connectrpc.com/connect"
	"github.com/coxwave/coupon-system/cmd/api/services"
	v1 "github.com/coxwave/coupon-system/gen/proto/campaign/v1"
	campaignv1connect "github.com/coxwave/coupon-system/gen/proto/campaign/v1/campaignv1connect"
)

type CampaignHandler struct {
	campaignv1connect.UnimplementedCampaignServiceHandler
	service *services.CampaignService
}

func NewCampaignHandler(service *services.CampaignService) *CampaignHandler {
	return &CampaignHandler{
		service: service,
	}
}

func (h *CampaignHandler) CreateCampaign(
	ctx context.Context,
	req *connect.Request[v1.CreateCampaignRequest],
) (*connect.Response[v1.CreateCampaignResponse], error) {
	resp, err := h.service.CreateCampaign(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *CampaignHandler) GetCampaign(
	ctx context.Context,
	req *connect.Request[v1.GetCampaignRequest],
) (*connect.Response[v1.GetCampaignResponse], error) {
	resp, err := h.service.GetCampaign(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(resp), nil
}
