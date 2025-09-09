package domain

import "context"

type PaymentIntent struct {
	OrderID  string
	Amount   int64
	Currency string // "jpy"
}

type PaymentGateway interface {
	Charge(ctx context.Context, intent PaymentIntent) (providerTxID string, err error)
}
