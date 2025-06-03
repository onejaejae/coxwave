package handlers

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/coxwave/coupon-system/cmd/api/services"
	v1 "github.com/coxwave/coupon-system/gen/proto/coupon/v1"
	couponv1connect "github.com/coxwave/coupon-system/gen/proto/coupon/v1/couponv1connect"
)

type CouponHandler struct {
	couponv1connect.UnimplementedCouponServiceHandler
	service *services.CouponService
}

func NewCouponHandler(service *services.CouponService) *CouponHandler {
	return &CouponHandler{
		service: service,
	}
}

func (h *CouponHandler) IssueCoupon(
	ctx context.Context,
	req *connect.Request[v1.IssueCouponRequest],
) (*connect.Response[v1.IssueCouponResponse], error) {
	fmt.Println("IssueCoupon")
	resp, err := h.service.IssueCoupon(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}
