package dbmodel

import (
	"github.com/kazshi01/payment-system/internal/domain/order"
	sqlcdb "github.com/kazshi01/payment-system/internal/infra/db/sqlc"
)

// sqlc（DB層の型）→ domain（ドメイン型）
func OrderToDomain(r sqlcdb.Order) *order.Order {
	return &order.Order{
		ID:        order.ID(r.ID),
		UserID:    r.UserID,
		AmountJPY: r.AmountJpy,            // BIGINT → int64
		Status:    order.Status(r.Status), // string → domain.Status
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// domain → sqlc Create用のParams
func CreateOrderParamsFromDomain(o *order.Order) sqlcdb.CreateOrderParams {
	return sqlcdb.CreateOrderParams{
		ID:        string(o.ID),
		UserID:    o.UserID,
		AmountJpy: o.AmountJPY,
		Status:    string(o.Status),
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

// domain → sqlc Update用のParams
func UpdateOrderParamsFromDomain(o *order.Order) sqlcdb.UpdateOrderParams {
	return sqlcdb.UpdateOrderParams{
		ID:        string(o.ID),
		AmountJpy: o.AmountJPY,
		Status:    string(o.Status),
		UpdatedAt: o.UpdatedAt,
	}
}
