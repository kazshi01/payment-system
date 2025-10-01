package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kazshi01/payment-system/internal/auth"
	"github.com/kazshi01/payment-system/internal/domain"
	"github.com/kazshi01/payment-system/internal/domain/order"
	"github.com/kazshi01/payment-system/internal/usecase"
)

// ---------- テストダブル ----------

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type fixedIDGen struct{ v string }

func (g fixedIDGen) New() string { return g.v }

type nopTx struct{}

func (nopTx) Do(ctx context.Context, fn func(ctx context.Context) error) error { return fn(ctx) }

type okPG struct {
	txid string
	err  error
}

func (p okPG) Charge(ctx context.Context, intent domain.PaymentIntent) (string, error) {
	return p.txid, p.err
}

type memRepo struct{ m map[order.ID]*order.Order }

func newMemRepo() *memRepo { return &memRepo{m: map[order.ID]*order.Order{}} }

func (r *memRepo) Create(ctx context.Context, o *order.Order) error {
	if _, dup := r.m[o.ID]; dup {
		return errors.New("dup")
	}
	// コピーして保存（テストでの副作用防止）
	cp := *o
	r.m[o.ID] = &cp
	return nil
}

func (r *memRepo) FindByID(ctx context.Context, id order.ID) (*order.Order, error) {
	o, ok := r.m[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *o
	return &cp, nil
}

func (r *memRepo) Update(ctx context.Context, o *order.Order) error {
	if _, ok := r.m[o.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *o
	r.m[o.ID] = &cp
	return nil
}

func (r *memRepo) UpdateStatusIfPending(ctx context.Context, id order.ID, st order.Status, at time.Time) (int64, error) {
	o, ok := r.m[id]
	if !ok {
		return 0, domain.ErrNotFound
	}
	if o.Status != order.StatusPending {
		return 0, nil
	}
	o.Status = st
	o.UpdatedAt = at
	return 1, nil
}

func (r *memRepo) FindByIDForUser(ctx context.Context, id order.ID, userID string) (*order.Order, error) {
	o, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if o.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return o, nil
}
func (r *memRepo) UpdateStatusIfPendingForUser(ctx context.Context, id order.ID, userID string, st order.Status, at time.Time) (int64, error) {
	o, err := r.FindByIDForUser(ctx, id, userID)
	if err != nil {
		return 0, err
	}
	if o.Status != order.StatusPending {
		return 0, nil
	}
	o.Status = st
	o.UpdatedAt = at
	r.m[o.ID] = o
	return 1, nil
}

// Locker ダミー（常にロック成功）
type okLocker struct{}

func (okLocker) TryLock(ctx context.Context, key string, ttlSeconds int) (bool, string, error) {
	return true, "tok", nil
}
func (okLocker) Unlock(ctx context.Context, key, token string) error { return nil }
func (okLocker) Ping(ctx context.Context) error                      { return nil }
func (okLocker) Close() error                                        { return nil }

// ---------- ユーティリティ ----------

func ctxWithUser(userID string) context.Context {
	// ミドルウェアが入れるのと同じ形で claims を突っ込む
	claims := map[string]any{
		"sub": userID,
	}
	return context.WithValue(context.Background(), auth.ClaimsKey, claims)
}

// ---------- テスト ----------

func TestOrderUsecase_CreateOrder_ok(t *testing.T) {
	repo := newMemRepo()
	uc := &usecase.OrderUsecase{
		Repo:   repo,
		Tx:     nopTx{},
		PG:     okPG{txid: "tx1"},
		Clock:  fixedClock{t: time.Date(2025, 9, 27, 10, 0, 0, 0, time.Local)},
		IDGen:  fixedIDGen{v: "order-1"},
		Locker: okLocker{},
	}

	ctx := ctxWithUser("user-1")

	o, err := uc.CreateOrder(ctx, 1200)
	if err != nil {
		t.Fatalf("CreateOrder err = %v", err)
	}
	if got, want := string(o.ID), "order-1"; got != want {
		t.Fatalf("id = %s; want %s", got, want)
	}
	if got, want := o.Status, order.StatusPending; got != want {
		t.Fatalf("status = %s; want %s", got, want)
	}

	stored, err := repo.FindByID(context.Background(), o.ID)
	if err != nil {
		t.Fatalf("repo.FindByID err = %v", err)
	}
	if stored.AmountJPY != 1200 || stored.UserID != "user-1" {
		t.Fatalf("stored mismatch: %+v", stored)
	}
}

func TestOrderUsecase_PayOrder_ok(t *testing.T) {
	repo := newMemRepo()
	now := time.Date(2025, 9, 27, 10, 0, 0, 0, time.Local)
	uc := &usecase.OrderUsecase{
		Repo:   repo,
		Tx:     nopTx{},
		PG:     okPG{txid: "tx-abc"},
		Clock:  fixedClock{t: now},
		IDGen:  fixedIDGen{v: "order-1"},
		Locker: okLocker{},
	}

	// 事前に注文を作成
	ctx := ctxWithUser("user-1")
	o, _ := uc.CreateOrder(ctx, 1000)

	if err := uc.PayOrder(ctx, o.ID); err != nil {
		t.Fatalf("PayOrder err = %v", err)
	}
	got, _ := repo.FindByID(context.Background(), o.ID)
	if got.Status != order.StatusPaid {
		t.Fatalf("status = %s; want PAID", got.Status)
	}
}

// ---------- エラーテスト ----------

func TestOrderUsecase_CreateOrder_invalidAmount(t *testing.T) {
	uc := &usecase.OrderUsecase{
		Repo:   newMemRepo(),
		Tx:     nopTx{},
		PG:     okPG{},
		Clock:  fixedClock{t: time.Now()},
		IDGen:  fixedIDGen{v: "x"},
		Locker: okLocker{},
	}
	ctx := ctxWithUser("user-1")

	if _, err := uc.CreateOrder(ctx, 0); !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("err = %v; want ErrInvalidArgument", err)
	}
}

func TestOrderUsecase_PayOrder_notFound(t *testing.T) {
	uc := &usecase.OrderUsecase{
		Repo:   newMemRepo(),
		Tx:     nopTx{},
		PG:     okPG{},
		Clock:  fixedClock{t: time.Now()},
		IDGen:  fixedIDGen{v: "x"},
		Locker: okLocker{},
	}
	ctx := ctxWithUser("user-1")

	err := uc.PayOrder(ctx, "unknown-id")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v; want ErrNotFound", err)
	}
}

func TestOrderUsecase_PayOrder_wrongUser(t *testing.T) {
	uc := &usecase.OrderUsecase{
		Repo:   newMemRepo(),
		Tx:     nopTx{},
		PG:     okPG{},
		Clock:  fixedClock{t: time.Now()},
		IDGen:  fixedIDGen{v: "x"},
		Locker: okLocker{},
	}

	o, _ := uc.CreateOrder(ctxWithUser("user-1"), 1200)

	err := uc.PayOrder(ctxWithUser("user-2"), o.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v; want ErrNotFound", err)
	}
}

func TestOrderUsecase_PayOrder_alreadyPaid(t *testing.T) {
	uc := &usecase.OrderUsecase{
		Repo:   newMemRepo(),
		Tx:     nopTx{},
		PG:     okPG{},
		Clock:  fixedClock{t: time.Now()},
		IDGen:  fixedIDGen{v: "x"},
		Locker: okLocker{},
	}
	ctx := ctxWithUser("user-1")

	o, _ := uc.CreateOrder(ctx, 1200)

	if err := uc.PayOrder(ctx, o.ID); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if err := uc.PayOrder(ctx, o.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("err = %v; want ErrConflict", err)
	}
}
