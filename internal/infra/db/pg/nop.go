package pg

import (
	"context"

	"github.com/kazshi01/payment-system/internal/domain"
)

type Nop struct{}

func (Nop) Charge(ctx context.Context, p domain.PaymentIntent) (string, error) {
	return "tx_mock", nil
}
