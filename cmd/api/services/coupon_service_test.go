// cmd/api/services/coupon_service_test.go
package services

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	couponv1 "github.com/coxwave/coupon-system/gen/proto/coupon/v1"
	"github.com/coxwave/coupon-system/internal/infrastructure/mysql/model"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	err := godotenv.Load("../../../.env.test")
	if err != nil {
		t.Fatalf("Failed to load .env file: %v", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbDatabase := os.Getenv("DB_DATABASE")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUsername, dbPassword, dbHost, dbPort, dbDatabase)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.Exec("SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED").Error; err != nil {
		t.Fatalf("Failed to set transaction isolation level: %v", err)
	}

	// MySQL 설정 변경
	if err := db.Exec("SET GLOBAL max_connections = 2000").Error; err != nil {
		t.Fatalf("Failed to set max_connections: %v", err)
	}
	if err := db.Exec("SET GLOBAL wait_timeout = 28800").Error; err != nil {
		t.Fatalf("Failed to set wait_timeout: %v", err)
	}
	if err := db.Exec("SET GLOBAL interactive_timeout = 28800").Error; err != nil {
		t.Fatalf("Failed to set interactive_timeout: %v", err)
	}

	if err := db.Exec("SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED").Error; err != nil {
		t.Fatalf("Failed to set transaction isolation level: %v", err)
	}

	// 테스트용 테이블 초기화
	db.Migrator().DropTable(&model.Coupon{}, &model.Campaign{})
	db.AutoMigrate(&model.Campaign{}, &model.Coupon{})

	return db
}

func TestConcurrentCouponIssuance(t *testing.T) {
	db := setupTestDB(t)
	service := NewCouponService(db)

	// 테스트용 캠페인 생성
	campaign := &model.Campaign{
		Name:           "Test Campaign",
		CouponQuantity: 100,                            // 100개의 쿠폰만 발급 가능
		IssueStartAt:   time.Now().Add(-1 * time.Hour), // 이미 시작된 캠페인
	}
	err := db.Create(campaign).Error
	assert.NoError(t, err)

	// 동시 요청 수
	concurrentRequests := 200
	successCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 동시 요청 시뮬레이션
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := service.IssueCoupon(context.Background(), &couponv1.IssueCouponRequest{
				CampaignId: uint32(campaign.ID),
			})

			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	t.Logf("Successfully issued coupons: %d", successCount)
	assert.Equal(t, 100, successCount, "Should only issue exactly 100 coupons")

	// 실제 발급된 쿠폰 수 확인
	var count int64
	db.Model(&model.Coupon{}).Where("campaign_id = ?", campaign.ID).Count(&count)
	assert.Equal(t, int64(100), count, "Database should have exactly 100 coupons")
}

func TestConcurrentCampaignLock(t *testing.T) {
	db := setupTestDB(t)
	service := NewCouponService(db)

	// 테스트용 캠페인 생성
	campaign := &model.Campaign{
		Name:           "Test Campaign",
		CouponQuantity: 1, // 1개의 쿠폰만 발급 가능
		IssueStartAt:   time.Now().Add(-1 * time.Hour),
	}
	err := db.Create(campaign).Error
	assert.NoError(t, err)

	// 동시 요청 수
	concurrentRequests := 10
	successCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 동시 요청 시뮬레이션
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := service.IssueCoupon(context.Background(), &couponv1.IssueCouponRequest{
				CampaignId: uint32(campaign.ID),
			})

			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	// 모든 고루틴이 완료될 때까지 대기
	wg.Wait()

	// 검증
	t.Logf("Successfully issued coupons: %d", successCount)
	assert.Equal(t, 1, successCount, "Should only issue exactly 1 coupon")
}
