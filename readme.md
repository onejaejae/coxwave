## 실행 방법

### local 개발용 DB 실행 방법

#### docker-compose up 실행

```
make docker-up
```

#### 마이그레이션 명령어 실행

```
make migration-up
```

처음 마이그레이션 실행 시, 해당 코드에서 에러가 난다면
주석 처리 해주세요

```go
	err = db.Migrator().DropTable(&model.Campaign{})
	err = db.Migrator().DropTable(&model.Coupon{})
	if err != nil {
		panic(err)
	}

```

<br>

### 테스트 코드 실행 방법

#### docker-compose up 실행

```
make docker-test-up
```

#### 테스트 실행 명령어를 입력해주세요

```
go test -v ./cmd/api/services/...
```

<br>

## 동시성 처리 설계

### 구현 방식

동시성 문제를 제어하기 위해 레코드 단위의 Exclusive Lock을 사용하였습니다.
또한, MySQL의 기본 격리 수준인 REPEATABLE READ는 트랜잭션 시작 시점의 데이터 스냅샷을 기준으로 동작하기 때문에, 트랜잭션 내에서 다른 트랜잭션이 커밋한 변경 사항을 인지할 수 없습니다.

예를 들어, 하나의 트랜잭션이 먼저 쿠폰을 발급했더라도, 이후 실행된 다른 트랜잭션에서는 해당 쿠폰 발급 내역을 조회할 수 없습니다. 이는 REPEATABLE READ 격리 수준의 특성으로 인해, 트랜잭션이 시작된 시점의 상태가 고정되어 변경 사항이 보이지 않기 때문입니다.

이러한 이유로 동시 트랜잭션 간 쿠폰 중복 발급 등의 문제가 발생할 수 있다고 판단하였고, 이를 방지하기 위해 격리 수준을 READ COMMITTED로 설정하여 커밋된 변경 사항은 즉시 반영되도록 하였습니다.

결론적으로,

- 레코드 수준의 Exclusive Lock

- READ COMMITTED 격리 수준
  을 조합하여 동시성 문제를 해결하였습니다

```go
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

        // ...

	})
}
```

## 구현하면서 새롭게 알게 된 점

동시성 테스트 과정에서, 격리 수준을 `REPEATABLE READ`로 설정할 경우 기대한 대로 동시성 제어가 되지 않을 것이라고 예상했습니다.

그러나 테스트 결과는 예상과 달리 동시성 제어가 정상적으로 이루어졌고, 그 이유를 분석하면서 중요한 사실을 알게 되었습니다.

```go
campaign, err := s.campaignRepo.GetByIDWithLock(ctx, req.CampaignId, tx)
```

현재 코드에서는 트랜잭션이 시작되자마자 `FOR UPDATE`를 통해 해당 캠페인 레코드에 락을 걸고 있습니다.
일반적으로 `REPEATABLE READ` 격리 수준에서는 트랜잭션 시작 시점의 스냅샷을 기준으로 데이터를 읽기 때문에,
다른 트랜잭션이 커밋한 변경 사항을 인식하지 못해 동시성 문제가 발생할 수 있습니다.

하지만 테스트에서는 동시성 문제가 발생하지 않았습니다.
그 이유는 `FOR UPDATE`로 인한 락 대기가 발생하면서, 트랜잭션이 실제로 시작된 시점이 지연되었기 때문입니다.

즉, 다른 트랜잭션이 락을 해제할 때까지 대기하며 트랜잭션은 논리적으로는 시작됐지만,
실제로는 첫 번째 쿼리(FOR UPDATE)가 실행되고 락을 획득한 시점에서 스냅샷이 생성되었습니다.
그로 인해 최신 커밋된 데이터를 기준으로 트랜잭션이 수행되었고, 예상과 달리 REPEATABLE READ에서도 동시성 제어가 가능했던 것입니다.

반면 아래와 같이 락을 걸기 전에 일반 조회 쿼리를 먼저 실행하는 경우에는 상황이 달라집니다:

```go
// 트랜잭션 시작 후 일반 조회 → 이후 락 시도
campaign, err := s.campaignRepo.GetByID(ctx, req.CampaignId, tx) // 일반 조회
campaign, err = s.campaignRepo.GetByIDWithLock(ctx, req.CampaignId, tx) // 락
```

이 경우 첫 번째 일반 조회 쿼리가 실행되는 시점에 트랜잭션의 스냅샷이 생성됩니다.
이후 FOR UPDATE를 실행하더라도 이미 생성된 스냅샷을 기준으로 하므로,
다른 트랜잭션의 커밋된 변경 사항을 반영하지 못하고 동시성 문제가 발생하였습니다.

실제로 같은 테스트를 `READ COMMITTED` 격리 수준으로 수행했을 때는,
커밋된 최신 데이터를 매번 조회할 수 있기 때문에 동시성 문제가 발생하지 않았습니다.
