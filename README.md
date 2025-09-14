# Payment System

## ディレクトリ構成

```
payment-system/
├── cmd/
│   └── api/
│       └── main.go
├── db/
│   ├── migrate.sh
│   ├── migrations/
│   │   ├── 0001_init.down.sql
│   │   └── 0001_init.up.sql
│   └── order_record.go
│   └── Makefile
├── internal/
│   │
│   ├── domain/
│   │   ├── errors.go
│   │   ├── payment_gateway.go
│   │   ├── repository.go
│   │   ├── order/
│   │   │   └── order.go
│   │   └── payment/
│   │       └── payment.go
│   │
│   ├── usecase/
│   │   └── order_usecase.go
│   ├── interface/
│   │   └── httpi/
│   │       ├── order_handler.go
│   │       └── respond.go
│   │
│   └── infra/
│       ├── clock/
│       │   └── system.go
│       ├── idgen/
│       │   └── uuidgen.go
│       └── db/
│           ├── dbmodel/
│           │   └── order.go
│           ├── order_repository.go
│           ├── tx.go
│           ├── pg/
│           │   └── nop.go
│           └── sqlc/
│               ├── db.go
│               ├── models.go
│               ├── order.sql.go
│               └── queries/
│                   └── order.sql
└── pkg/
    └── .gitkeep
```

## API

- Docker Desktop を起動する

- マイグレーションを適用する

```
make migrate.up
```

- サーバーを起動する

```
go run cmd/api/main.go
```

- 注文を作成する

```
curl -i -X POST http://localhost:8080/orders -H 'Content-Type: application/json' -d '{"amount_jpy": 1200}'
```

- 注文を支払う

```
curl -i -X POST http://localhost:8080/orders/{id}/pay 
```
