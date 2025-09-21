package order

import "time"

type ID string
type Status string

const (
	StatusPending  Status = "PENDING"
	StatusPaid     Status = "PAID"
	StatusCanceled Status = "CANCELED"
)

type Order struct {
	ID        ID
	UserID    string
	AmountJPY int64
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OrderのStatusを"PAID"に切り替える
func (o *Order) MarkPaid() { o.Status = StatusPaid }
