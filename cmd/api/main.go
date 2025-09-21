package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/kazshi01/payment-system/internal/auth"
	"github.com/kazshi01/payment-system/internal/infra/clock"
	"github.com/kazshi01/payment-system/internal/infra/db"
	"github.com/kazshi01/payment-system/internal/infra/db/pg"
	"github.com/kazshi01/payment-system/internal/infra/idgen"
	"github.com/kazshi01/payment-system/internal/interface/httpi"
	"github.com/kazshi01/payment-system/internal/usecase"
)

func main() {
	// --- .env を読み込む ---

	// 開発時は.envがないとエラーにする
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	name := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", user, pass, host, name)

	// --- DB 接続 ---
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("DB connected")

	// --- Repository & Tx ---
	repo := db.NewPostgresOrderRepository(sqlDB)
	txMgr := &db.TxManager{DB: sqlDB}

	// --- Payment Gateway ---
	gateway := pg.Nop{} // まだモック

	// --- Usecase ---
	orderUC := &usecase.OrderUsecase{
		Repo:  repo,
		Tx:    txMgr,
		PG:    gateway,
		Clock: clock.System{},
		IDGen: idgen.UUIDGen{},
	}

	// --- OrderHandler ---
	handler := &httpi.OrderHandler{UC: orderUC}

	// --- AuthHandler ---
	authH, err := httpi.NewAuthHandler(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// --- Middleware M2M ---

	mw, err := auth.Middleware(auth.Config{
		Issuer:   os.Getenv("OIDC_ISSUER"),
		Audience: os.Getenv("OIDC_AUDIENCE"),
	})
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", httpi.Home)

	mux.Handle("POST /orders", mw(http.HandlerFunc(handler.Create)))
	mux.Handle("POST /orders/{id}/pay", mw(http.HandlerFunc(handler.Pay)))

	mux.HandleFunc("GET /auth/login", authH.Login)
	mux.HandleFunc("GET /auth/callback", authH.Callback)
	mux.HandleFunc("POST /auth/refresh", authH.Refresh)
	mux.HandleFunc("GET /auth/logout", authH.Logout)
	mux.HandleFunc("POST /auth/logout", authH.Logout)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
