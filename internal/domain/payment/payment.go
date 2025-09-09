package payment

type ID string
type Method string

const (
	MethodCard Method = "CARD"
)

type Payment struct {
	ID       ID
	OrderID  string
	Method   Method
	Provider string // e.g. "stripe"
	TxID     string // プロバイダ側のトランザクションID
}
