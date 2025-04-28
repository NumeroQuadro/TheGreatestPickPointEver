package postgresql

import (
	"context"
	"database/sql"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/require"
	mock_database "gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/db/mocks"
	"testing"
	"time"
)

func TestOrderRepo_Delete(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()
	)
	correctValues := struct {
		OrderId          int64
		UserID           int64
		ExpirationTime   time.Time
		Status           string
		Weight           int
		Cost             int
		PackageType      string
		IsAdditionalFilm bool
	}{1, 1, time.Now(), "confirmed", 1, 1, "film", false}

	t.Run("fail", func(t *testing.T) {
		t.Parallel()
		t.Run("not found", func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockDb := mock_database.NewMockDB(ctrl)
			mockDb.EXPECT().Exec(ctx, `DELETE FROM orders WHERE order_id = $1;`, correctValues.OrderId).Return(pgconn.CommandTag("DELETE 0"), nil)
			repo := NewOrdersRepo(mockDb)

			err := repo.Delete(ctx, correctValues.OrderId)

			require.EqualError(t, err, "order not found")
		})
	})
}

func TestOrderRepo_Find(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()
	)
	correctValues := struct {
		OrderId          int64
		UserID           int64
		ExpirationTime   time.Time
		Status           string
		Weight           int
		Cost             int
		PackageType      string
		IsAdditionalFilm bool
	}{1, 1, time.Now(), "confirmed", 1, 1, "film", false}
	t.Run("fail", func(t *testing.T) {
		t.Parallel()
		t.Run("not found", func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockDb := mock_database.NewMockDB(ctrl)
			mockDb.EXPECT().Get(gomock.Any(), gomock.Any(), `SELECT 
		order_id,
		user_id,
		expiration_date,
		status,
		weight,
		cost,
		last_changed_at
	FROM orders
	WHERE order_id = $1;
	`, gomock.Any()).Return(sql.ErrNoRows)
			repo := NewOrdersRepo(mockDb)

			_, err := repo.Find(ctx, correctValues.UserID)

			require.EqualError(t, err, "order not found")
		})
	})
}
