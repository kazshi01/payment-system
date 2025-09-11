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

// 注文を作成する
func (uc *OrderUsecase) CreateOrder(ctx context.Context, amountJPY int64) (*order.Order, error) {

	// ① 注文を作成
	o := &order.Order{
		ID:        order.ID(generateID()), // 実装は省略
		AmountJPY: amountJPY,
		Status:    order.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// ② 注文をDBに保存
	if err := uc.Repo.Create(ctx, o); err != nil {
		return nil, err
	}

	return o, nil
}

// 注文を支払いする
func (uc *OrderUsecase) PayOrder(ctx context.Context, id order.ID) error {
	return uc.Tx.Do(ctx, func(ctx context.Context) error {

		// ① 注文を取得
		o, err := uc.Repo.FindByID(ctx, id)
		if err != nil {
			return err
		}

		// ② 外部決済サービスで支払い実行
		txID, err := uc.PG.Charge(ctx, domain.PaymentIntent{
			OrderID: string(o.ID), Amount: o.AmountJPY, Currency: "jpy",
		})
		if err != nil {
			return err
		}

		// ③ 後で保存する予定なので今は未使用
		_ = txID // 後で payments / events に保存

		// ④ 注文ステータスを「支払い済み」に変更
		o.MarkPaid()
		o.UpdatedAt = time.Now()

		// ⑤ 注文情報をDBに更新
		return uc.Repo.Update(ctx, o)
	})
}
