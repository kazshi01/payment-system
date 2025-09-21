package db

import (
	"time"

	"github.com/kazshi01/payment-system/internal/domain/order"
)

// DBのテーブルと1対1で対応
type OrderRecord struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	AmountJPY int64     `db:"amount_jpy"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// OrderRecord → Domain変換
func (r OrderRecord) ToDomain() *order.Order {
	return &order.Order{
		ID:        order.ID(r.ID),
		UserID:    r.UserID,
		AmountJPY: r.AmountJPY,
		Status:    order.Status(r.Status),
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// Domain → OrderRecord変換
func FromDomain(o *order.Order) OrderRecord {
	return OrderRecord{
		ID:        string(o.ID),
		UserID:    o.UserID,
		AmountJPY: o.AmountJPY,
		Status:    string(o.Status),
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}
