package usecase

import (
	"context"
	"time"

	"github.com/kazshi01/payment-system/internal/domain"
	"github.com/kazshi01/payment-system/internal/domain/order"
)

type OrderUsecase struct {
	Repo domain.OrderRepository
	Tx   domain.Tx
	PG   domain.PaymentGateway
}

func generateID() string { return "1" } // これは後で実装！！

func (uc *OrderUsecase) CreateOrder(ctx context.Context, amountJPY int64) (*order.Order, error) {
	o := &order.Order{
		ID:        order.ID(generateID()), // 実装は省略
		AmountJPY: amountJPY,
		Status:    order.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.Repo.Create(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (uc *OrderUsecase) PayOrder(ctx context.Context, id order.ID) error {
	return uc.Tx.Do(ctx, func(ctx context.Context) error {
		o, err := uc.Repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		txID, err := uc.PG.Charge(ctx, domain.PaymentIntent{
			OrderID: string(o.ID), Amount: o.AmountJPY, Currency: "jpy",
		})
		if err != nil {
			return err
		}
		_ = txID // 後で payments / events に保存
		o.MarkPaid()
		o.UpdatedAt = time.Now()
		return uc.Repo.Update(ctx, o)
	})
}
