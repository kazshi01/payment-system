package domain

import (
	"context"
	"time"

	"github.com/kazshi01/payment-system/internal/domain/order"
)

type OrderRepository interface {
	Create(ctx context.Context, o *order.Order) error
	FindByID(ctx context.Context, id order.ID) (*order.Order, error)
	FindByIDForUser(ctx context.Context, id order.ID, userID string) (*order.Order, error)
	Update(ctx context.Context, o *order.Order) error
	UpdateStatusIfPending(ctx context.Context, id order.ID, newStatus order.Status, updatedAt time.Time) (int64, error)
	UpdateStatusIfPendingForUser(ctx context.Context, id order.ID, userID string, newStatus order.Status, updatedAt time.Time) (int64, error)
}

type Tx interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
