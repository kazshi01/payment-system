package usecase

import (
	"context"
	"time"

	"github.com/kazshi01/payment-system/internal/domain"
	"github.com/kazshi01/payment-system/internal/domain/order"
)

const CurrencyJPY = "jpy"

type Clock interface{ Now() time.Time }
type IDGen interface{ New() string }

type OrderUsecase struct {
	Repo domain.OrderRepository
	Tx   domain.Tx
	PG   domain.PaymentGateway

	Clock Clock
	IDGen IDGen
}

func (uc *OrderUsecase) now() time.Time {
	if uc.Clock != nil {
		return uc.Clock.Now()
	}
	return time.Now()
}

// --- Create ---

func (uc *OrderUsecase) CreateOrder(ctx context.Context, amountJPY int64) (*order.Order, error) {
	if amountJPY <= 0 {
		return nil, domain.ErrInvalidArgument
	}

	if uc.IDGen == nil {
		return nil, domain.ErrInternal
	}

	o := &order.Order{
		ID:        order.ID(uc.IDGen.New()),
		AmountJPY: amountJPY,
		Status:    order.StatusPending,
		CreatedAt: uc.now(),
		UpdatedAt: uc.now(),
	}

	// ---- DB 反映は 3s ----
	dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := uc.Repo.Create(dbCtx, o); err != nil {
		return nil, err
	}
	return o, nil
}

// --- Pay ---

// 外部決済(PG)はTxの外で行い、DB反映はTxでまとめる
func (uc *OrderUsecase) PayOrder(ctx context.Context, id order.ID) error {
	// 1) 注文取得（Tx外・3s）
	dbReadCtx, cancelRead := context.WithTimeout(ctx, 3*time.Second)
	defer cancelRead()

	o, err := uc.Repo.FindByID(dbReadCtx, id)
	if err != nil {
		return err
	}
	if o.Status != order.StatusPending {
		return domain.ErrConflict
	}

	//    - idempotencyKey を付ける（注文IDなど）
	// ---- PG 呼び出しは 5s ----
	pgCtx, cancelPG := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPG()

	txID, err := uc.PG.Charge(pgCtx, domain.PaymentIntent{
		OrderID:        string(o.ID),
		Amount:         o.AmountJPY,
		Currency:       CurrencyJPY,
		IdempotencyKey: string(o.ID), // OrderIDをキーにすれば冪等性を保証できる
	})
	if err != nil {
		return err
	}

	// ---- DB 反映は 3s ----
	dbCtx, cancelDB := context.WithTimeout(ctx, 3*time.Second)
	defer cancelDB()

	return uc.Tx.Do(dbCtx, func(dbCtx context.Context) error {
		// 最新状態の軽い再確認
		oo, err := uc.Repo.FindByID(dbCtx, id)
		if err != nil {
			return err
		}
		if oo.Status != order.StatusPending {
			return domain.ErrConflict
		}

		_ = txID // Todo
		// 支払いレコードやイベントを保存
		// 例:
		// if err := uc.PaymentsRepo.Create(dbCtx, payment.Payment{...txID...}); err != nil { return err }
		// if err := uc.PaymentEventsRepo.Append(dbCtx, ...); err != nil { return err }

		oo.MarkPaid()
		updatedAt := uc.now()
		rows, err := uc.Repo.UpdateStatusIfPending(dbCtx, oo.ID, order.StatusPaid, updatedAt)
		if err != nil {
			return err
		}
		if rows == 0 {
			return domain.ErrConflict
		}
		return nil
	})
}
