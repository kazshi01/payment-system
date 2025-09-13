package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kazshi01/payment-system/internal/domain"
	"github.com/kazshi01/payment-system/internal/domain/order"
	sqlcdb "github.com/kazshi01/payment-system/internal/infra/db/sqlc"
)

// PostgresOrderRepository implements domain.OrderRepository using sqlc.
type PostgresOrderRepository struct {
	DB *sql.DB
	Q  *sqlcdb.Queries
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{
		DB: db,
		Q:  sqlcdb.New(db),
	}
}

// Tx が ctx に乗っていればその Tx にバインドした Queries を返す
func (r *PostgresOrderRepository) getQ(ctx context.Context) *sqlcdb.Queries {
	if tx := getTx(ctx); tx != nil {
		return r.Q.WithTx(tx)
	}
	return r.Q
}

// Create inserts a new order.
func (r *PostgresOrderRepository) Create(ctx context.Context, o *order.Order) error {
	params := sqlcdb.CreateOrderParams{
		ID:        string(o.ID),
		AmountJpy: o.AmountJPY,
		Status:    string(o.Status),
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
	if err := r.getQ(ctx).CreateOrder(ctx, params); err != nil {
		return fmt.Errorf("create order: %w", err)
	}
	return nil
}

// FindByID fetches an order by ID.
func (r *PostgresOrderRepository) FindByID(ctx context.Context, id order.ID) (*order.Order, error) {
	rec, err := r.getQ(ctx).GetOrder(ctx, string(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get order: %w", err)
	}

	o := &order.Order{
		ID:        order.ID(rec.ID),
		AmountJPY: rec.AmountJpy,
		Status:    order.Status(rec.Status),
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
	}
	return o, nil
}

// Update updates mutable fields of an order.
func (r *PostgresOrderRepository) Update(ctx context.Context, o *order.Order) error {
	params := sqlcdb.UpdateOrderParams{
		ID:        string(o.ID),
		AmountJpy: o.AmountJPY,
		Status:    string(o.Status),
		UpdatedAt: o.UpdatedAt,
	}
	if err := r.getQ(ctx).UpdateOrder(ctx, params); err != nil {
		return fmt.Errorf("update order: %w", err)
	}
	return nil
}

func (r *PostgresOrderRepository) UpdateStatusIfPending(
	ctx context.Context,
	id order.ID,
	newStatus order.Status,
	updatedAt time.Time,
) (int64, error) {

	params := sqlcdb.UpdateOrderStatusIfPendingParams{
		ID:        string(id),
		Status:    string(newStatus),
		UpdatedAt: updatedAt,
	}
	affected, err := r.getQ(ctx).UpdateOrderStatusIfPending(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("update order status if pending: %w", err)
	}
	return affected, nil
}
