package main

import (
	"log"
	"net/http"

	"github.com/coxwave/coupon-system/cmd/api/handlers"
	"github.com/coxwave/coupon-system/cmd/api/services"
	campaignv1connect "github.com/coxwave/coupon-system/gen/proto/campaign/v1/campaignv1connect"
	couponv1connect "github.com/coxwave/coupon-system/gen/proto/coupon/v1/couponv1connect"
	"github.com/coxwave/coupon-system/internal/infrastructure/mysql"
)

func main() {
	// DB 연결
	db, err := mysql.NewMySQL()
	if err != nil {
		log.Fatal(err)
	}

	service := services.NewService(db)
	campaignHandler := handlers.NewCampaignHandler(service.CampaignService)
	couponHandler := handlers.NewCouponHandler(service.CouponService)

	path, handler := campaignv1connect.NewCampaignServiceHandler(campaignHandler)
	path2, handler2 := couponv1connect.NewCouponServiceHandler(couponHandler)
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	mux.Handle(path2, handler2)

	// 서버 시작
	log.Printf("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
