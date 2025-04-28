package service

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	mock_repository "gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/service/mocks"
	"testing"
	"time"
)

func TestOrderServiceImpl_AddOrder(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()
	)
	correctValues := struct {
		OrderId           int64
		UserID            int64
		ExpirationTime    time.Time
		StatusString      string
		StatusModel       domain.Status
		Weight            int
		Cost              int
		PackageTypeString string
		PackageTypeModel  domain.PackageType
		IsAdditionalFilm  bool
	}{1, 1, time.Now().AddDate(0, 0, 1), "confirmed", domain.Confirmed, 1, 1, "box", domain.Box, false}

	t.Run("expiration date is before current", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		prevTime := time.Now().AddDate(0, 0, -1)
		dto := OrderDto{correctValues.OrderId, correctValues.UserID, prevTime, correctValues.StatusModel, correctValues.Weight, correctValues.Cost}
		repo := mock_repository.NewMockOrderRepository(ctrl)
		repo.EXPECT().Create(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		srv := OrderServiceImpl{repo: repo}

		_, err := srv.AddOrder(ctx, dto, "box", correctValues.IsAdditionalFilm)

		require.EqualError(t, err, "expiration date is in the past")
	})
	t.Run("smoke test", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mock_repository.NewMockOrderRepository(ctrl)
		dto := OrderDto{correctValues.OrderId, correctValues.UserID, correctValues.ExpirationTime, correctValues.StatusModel, correctValues.Weight, correctValues.Cost}
		repo.EXPECT().Create(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
		srv := OrderServiceImpl{repo: repo}

		_, err := srv.AddOrder(ctx, dto, "box", correctValues.IsAdditionalFilm)

		require.NoError(t, err)
	})
}

func Test_applyPackagingStrategy(t *testing.T) {
	t.Parallel()
	type args struct {
		packageStrategy    domain.Package
		withAdditionalFilm bool
		weight             int
		cost               int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"smoke test", args{&domain.BoxPackage{}, false, 1, 1}, 21, false},
		{"smoke test", args{&domain.BagPackage{}, false, 1, 1}, 6, false},
		{"smoke test", args{&domain.FilmPackage{}, false, 1, 1}, 2, false},
		{"box with too much weight", args{&domain.BoxPackage{}, false, 1000, 1}, 0, true},
		{"bag with too much weight", args{&domain.BagPackage{}, false, 1000, 1}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyPackagingStrategy(tt.args.packageStrategy, tt.args.withAdditionalFilm, tt.args.weight, tt.args.cost)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyPackagingStrategy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("applyPackagingStrategy() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrderServiceImpl_ReturnOrder(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()
	)
	correctValues := struct {
		OrderId           int64
		UserID            int64
		ExpirationTime    time.Time
		StatusString      string
		StatusModel       domain.Status
		Weight            int
		Cost              int
		PackageTypeString string
		PackageTypeModel  domain.PackageType
		IsAdditionalFilm  bool
	}{1, 1, time.Now(), "confirmed", domain.Confirmed, 1, 1, "box", domain.Box, false}

	t.Run("order not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mock_repository.NewMockOrderRepository(ctrl)
		repo.EXPECT().Find(ctx, correctValues.OrderId).Return(domain.Order{}, domain.ErrOrderNotFound)
		repo.EXPECT().Delete(ctx, gomock.Any()).Times(0)
		srv := OrderServiceImpl{repo: repo}

		err := srv.ReturnOrder(ctx, correctValues.OrderId)

		require.EqualError(t, err, "order not found")
	})
	t.Run("order already completed", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mock_repository.NewMockOrderRepository(ctrl)
		repo.EXPECT().Find(ctx, correctValues.OrderId).Return(domain.Order{Status: domain.Completed}, nil)
		repo.EXPECT().Delete(ctx, gomock.Any()).Times(0)
		srv := OrderServiceImpl{repo: repo}

		err := srv.ReturnOrder(ctx, correctValues.OrderId)

		require.EqualError(t, err, "order already completed")
	})
	t.Run("order has to be refunded", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mock_repository.NewMockOrderRepository(ctrl)
		repo.EXPECT().Find(ctx, correctValues.OrderId).Return(domain.Order{Status: correctValues.StatusModel}, nil)
		repo.EXPECT().Delete(ctx, gomock.Any()).Times(0)
		srv := OrderServiceImpl{repo: repo}

		err := srv.ReturnOrder(ctx, correctValues.OrderId)

		require.EqualError(t, err, "order has to be refunded")
	})
	t.Run("expiration date is in the future", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mock_repository.NewMockOrderRepository(ctrl)
		futureDate := time.Now().AddDate(0, 0, 1)
		repo.EXPECT().Find(ctx, correctValues.OrderId).Return(domain.Order{Status: correctValues.StatusModel, ExpirationTime: futureDate}, nil)
		repo.EXPECT().Delete(ctx, gomock.Any()).Times(0)
		srv := OrderServiceImpl{repo: repo}

		err := srv.ReturnOrder(ctx, correctValues.OrderId)

		require.EqualError(t, err, "expiration date is in the future")
	})
}

func TestOrderServiceImpl_RefundOrder(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()
	)
	correctValues := struct {
		OrderId           int64
		UserID            int64
		ExpirationTime    time.Time
		StatusString      string
		StatusModel       domain.Status
		Weight            int
		Cost              int
		PackageTypeString string
		PackageTypeModel  domain.PackageType
		IsAdditionalFilm  bool
	}{1, 1, time.Now(), "confirmed", domain.Confirmed, 1, 1, "box", domain.Box, false}
	const expirationDays = 7

	t.Run("order not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mock_repository.NewMockOrderRepository(ctrl)
		repo.EXPECT().Find(ctx, correctValues.OrderId).Return(domain.Order{}, domain.ErrOrderNotFound)
		repo.EXPECT().Update(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		srv := OrderServiceImpl{repo: repo}

		err := srv.RefundOrder(ctx, correctValues.OrderId, expirationDays)

		require.EqualError(t, err, "order not found")
	})

	t.Run("order not completed", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mock_repository.NewMockOrderRepository(ctrl)
		repo.EXPECT().Find(ctx, correctValues.OrderId).Return(domain.Order{Status: domain.Confirmed}, nil)
		repo.EXPECT().Update(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		srv := OrderServiceImpl{repo: repo}

		err := srv.RefundOrder(ctx, correctValues.OrderId, expirationDays)

		require.EqualError(t, err, "order is not completed")
	})

	t.Run("refund period exceeded", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mock_repository.NewMockOrderRepository(ctrl)
		lastChangedAt := time.Now().Add(time.Duration(24*expirationDays+1) * time.Hour)
		repo.EXPECT().Find(ctx, correctValues.OrderId).Return(domain.Order{
			Status:        domain.Completed,
			LastChangedAt: lastChangedAt,
		}, nil)
		repo.EXPECT().Update(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		srv := OrderServiceImpl{repo: repo}

		err := srv.RefundOrder(ctx, correctValues.OrderId, expirationDays)

		require.EqualError(t, err, "order cannot be refunded")
	})

	t.Run("successful refund", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mock_repository.NewMockOrderRepository(ctrl)
		lastChangedAt := time.Now().Add(-time.Hour) // 1 hour ago
		order := domain.Order{
			OrderID:        correctValues.OrderId,
			UserID:         correctValues.UserID,
			ExpirationTime: correctValues.ExpirationTime,
			Status:         domain.Completed,
			Weight:         correctValues.Weight,
			Cost:           correctValues.Cost,
			LastChangedAt:  lastChangedAt,
		}
		repo.EXPECT().Find(ctx, correctValues.OrderId).Return(order, nil)
		repo.EXPECT().Update(ctx, order.OrderID, order.UserID, order.ExpirationTime, domain.Refunded, order.Weight, order.Cost).
			Return(correctValues.OrderId, nil)
		srv := OrderServiceImpl{repo: repo}

		err := srv.RefundOrder(ctx, correctValues.OrderId, expirationDays)

		require.NoError(t, err)
	})
}

func TestOrderServiceImpl_CompleteOrder(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()
	)
	correctValues := struct {
		OrderId        int64
		UserID         int64
		ExpirationTime time.Time
		StatusModel    domain.Status
		Weight         int
		Cost           int
	}{1, 1, time.Now().Add(24 * time.Hour), domain.Confirmed, 1, 1} // Future expiration time

	t.Run("order not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mock_repository.NewMockOrderRepository(ctrl)
		repo.EXPECT().Find(ctx, correctValues.OrderId).Return(domain.Order{}, domain.ErrOrderNotFound)
		repo.EXPECT().Update(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		srv := OrderServiceImpl{repo: repo}

		err := srv.CompleteOrder(ctx, correctValues.OrderId, correctValues.UserID)

		require.EqualError(t, err, "order not found")
	})

	t.Run("order does not belong to user", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mock_repository.NewMockOrderRepository(ctrl)
		order := domain.Order{UserID: correctValues.UserID + 1, ExpirationTime: correctValues.ExpirationTime}
		repo.EXPECT().Find(ctx, correctValues.OrderId).Return(order, nil)
		repo.EXPECT().Update(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		srv := OrderServiceImpl{repo: repo}

		err := srv.CompleteOrder(ctx, correctValues.OrderId, correctValues.UserID)

		require.EqualError(t, err, "order is not belong to user")
	})

	t.Run("expiration date in past", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mock_repository.NewMockOrderRepository(ctrl)
		pastExpiration := time.Now().Add(-24 * time.Hour)
		order := domain.Order{
			OrderID:        correctValues.OrderId,
			UserID:         correctValues.UserID,
			ExpirationTime: pastExpiration,
			Status:         correctValues.StatusModel,
		}
		repo.EXPECT().Find(ctx, correctValues.OrderId).Return(order, nil)
		repo.EXPECT().Update(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		srv := OrderServiceImpl{repo: repo}

		err := srv.CompleteOrder(ctx, correctValues.OrderId, correctValues.UserID)

		require.EqualError(t, err, "expiration date is in the past")
	})

	t.Run("order already completed", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mock_repository.NewMockOrderRepository(ctrl)
		order := domain.Order{
			OrderID:        correctValues.OrderId,
			UserID:         correctValues.UserID,
			ExpirationTime: correctValues.ExpirationTime,
			Status:         domain.Completed,
		}
		repo.EXPECT().Find(ctx, correctValues.OrderId).Return(order, nil)
		repo.EXPECT().Update(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		srv := OrderServiceImpl{repo: repo}

		err := srv.CompleteOrder(ctx, correctValues.OrderId, correctValues.UserID)

		require.EqualError(t, err, "order already completed")
	})

	t.Run("successful completion", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mock_repository.NewMockOrderRepository(ctrl)
		order := domain.Order{
			OrderID:        correctValues.OrderId,
			UserID:         correctValues.UserID,
			ExpirationTime: correctValues.ExpirationTime,
			Status:         correctValues.StatusModel,
			Weight:         correctValues.Weight,
			Cost:           correctValues.Cost,
		}
		repo.EXPECT().Find(ctx, correctValues.OrderId).Return(order, nil)
		repo.EXPECT().Update(ctx, order.OrderID, order.UserID, order.ExpirationTime, domain.Completed, order.Weight, order.Cost).
			Return(correctValues.OrderId, nil)
		srv := OrderServiceImpl{repo: repo}

		err := srv.CompleteOrder(ctx, correctValues.OrderId, correctValues.UserID)

		require.NoError(t, err)
	})
}
