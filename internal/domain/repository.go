package domain

import (
	"context"

	"github.com/kazshi01/payment-system/internal/domain/order"
)

type OrderRepository interface {
	Create(ctx context.Context, o *order.Order) error
	FindByID(ctx context.Context, id order.ID) (*order.Order, error)
	Update(ctx context.Context, o *order.Order) error
}

type Tx interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
