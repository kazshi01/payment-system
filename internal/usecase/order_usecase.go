package usecase

import (
	"context"
	"time"

	"github.com/kazshi01/payment-system/internal/auth"
	"github.com/kazshi01/payment-system/internal/domain"
	"github.com/kazshi01/payment-system/internal/domain/order"
)

const (
	CurrencyJPY = "jpy"
	lockTTL     = 15
)

type Clock interface{ Now() time.Time }
type IDGen interface{ New() string }

type OrderUsecase struct {
	Repo domain.OrderRepository
	Tx   domain.Tx
	PG   domain.PaymentGateway

	Clock  Clock
	IDGen  IDGen
	Locker domain.Locker
}

// --- Create ---

func (uc *OrderUsecase) CreateOrder(ctx context.Context, amountJPY int64) (*order.Order, error) {
	if amountJPY <= 0 {
		return nil, domain.ErrInvalidArgument
	}

	if uc.IDGen == nil {
		return nil, domain.ErrInternal
	}

	// ログインユーザーIDを取得
	userID, ok := auth.UserIDFrom(ctx)
	if !ok || userID == "" {
		return nil, domain.ErrUnauthorized
	}

	o := &order.Order{
		ID:        order.ID(uc.IDGen.New()),
		UserID:    userID,
		AmountJPY: amountJPY,
		Status:    order.StatusPending,
		CreatedAt: uc.Clock.Now(),
		UpdatedAt: uc.Clock.Now(),
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
	isAdmin := auth.IsAdmin(ctx)

	// 一般ユーザは userID 必須。管理者は不要
	userID, _ := auth.UserIDFrom(ctx)
	if !isAdmin {
		if userID == "" {
			return domain.ErrUnauthorized
		}
	}

	// 入口ガード（同時実行を1本化）
	lockKey := "lock:pay:" + string(id)

	ok, token, err := uc.Locker.TryLock(ctx, lockKey, lockTTL)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrConflict
	}

	defer func() {
		uctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_ = uc.Locker.Unlock(uctx, lockKey, token)
	}()

	// ---- 注文取得は 3s ----
	dbReadCtx, cancelRead := context.WithTimeout(ctx, 3*time.Second)
	defer cancelRead()

	var (
		o *order.Order
	)

	if isAdmin {
		o, err = uc.Repo.FindByID(dbReadCtx, id)
	} else {
		o, err = uc.Repo.FindByIDForUser(dbReadCtx, id, userID)
	}
	if err != nil {
		return err
	}
	if o.Status != order.StatusPending {
		return domain.ErrConflict
	}

	// ---- PG 呼び出しは 5s ----
	pgCtx, cancelPG := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPG()

	txID, err := uc.PG.Charge(pgCtx, domain.PaymentIntent{
		OrderID:        string(o.ID),
		Amount:         o.AmountJPY,
		Currency:       CurrencyJPY,
		IdempotencyKey: "pay:" + string(o.ID), // 他操作(cancel/refund)は将来別prefixで対応
	})
	if err != nil {
		return err
	}

	// ---- DB 反映は 3s ----
	dbCtx, cancelDB := context.WithTimeout(ctx, 3*time.Second)
	defer cancelDB()

	return uc.Tx.Do(dbCtx, func(dbCtx context.Context) error {
		updatedAt := uc.Clock.Now()

		var rows int64

		if isAdmin {
			rows, err = uc.Repo.UpdateStatusIfPending(dbCtx, o.ID, order.StatusPaid, updatedAt)
		} else {
			rows, err = uc.Repo.UpdateStatusIfPendingForUser(dbCtx, o.ID, userID, order.StatusPaid, updatedAt)
		}
		if err != nil {
			return err
		}
		if rows == 0 {
			return domain.ErrConflict
		}

		_ = txID // 将来 payments / events で利用

		return nil
	})
}
