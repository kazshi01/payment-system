package domain

import "context"

type PaymentIntent struct {
	OrderID        string
	Amount         int64
	Currency       string // "jpy"
	IdempotencyKey string // 外部PGに渡して二重請求を防ぐ
}

type PaymentGateway interface {
	Charge(ctx context.Context, intent PaymentIntent) (providerTxID string, err error)
}

/**
Order（注文）
  ↓ 決済を開始したい
PaymentIntent（支払い要求）
  ↓ 外部決済サービスに送信
Payment（実際の支払い結果）
**/
