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

## 認証あり

```
docker run -p 8081:8080 \
  -v "$(pwd)/kc-data/data:/opt/keycloak/data" \
  -e KEYCLOAK_ADMIN=admin \
  -e KEYCLOAK_ADMIN_PASSWORD=admin \
  quay.io/keycloak/keycloak:24.0.5 start-dev
```

```
# 1) 発行（client_credentials）
TOKEN=$(curl -s -X POST \
  -d 'grant_type=client_credentials' \
  -d 'client_id=payment-api' \
  -d 'client_secret=<SECRET>' \
  http://localhost:8081/realms/payment/protocol/openid-connect/token \
  | jq -r .access_token)

# 2) 注文作成
RESP=$(curl -s -i -X POST http://localhost:8080/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount_jpy":1200}')

echo "$RESP"
ORDER_ID=$(echo "$RESP" | sed -n 's/Location: \/orders\/\(.*\)/\1/p' | tr -d '\r')
echo "ORDER_ID=$ORDER_ID"

# 3) 支払い
curl -i -X POST "http://localhost:8080/orders/$ORDER_ID/pay" \
  -H "Authorization: Bearer $TOKEN"
```
