# Payment System

## ディレクトリ構成

```
├── cmd
│   └── api
│       └── main.go
├── db
│   ├── migrate.sh
│   ├── migrations
│   │   ├── 0001_init.down.sql
│   │   └── 0001_init.up.sql
│   └── order_record.go
├── docker-compose.yaml
├── internal
│   ├── auth
│   │   ├── middleware.go
│   │   └── oidc-pkce.go
│   ├── domain
│   │   ├── errors.go
│   │   ├── order
│   │   │   └── order.go
│   │   ├── payment
│   │   │   └── payment.go
│   │   ├── payment_gateway.go
│   │   └── repository.go
│   ├── infra
│   │   ├── clock
│   │   │   └── system.go
│   │   ├── db
│   │   │   ├── dbmodel
│   │   │   │   └── order.go
│   │   │   ├── order_repository.go
│   │   │   ├── pg
│   │   │   │   └── nop.go
│   │   │   ├── sqlc
│   │   │   │   ├── db.go
│   │   │   │   ├── models.go
│   │   │   │   ├── order.sql.go
│   │   │   │   └── queries
│   │   │   │       └── order.sql
│   │   │   └── tx.go
│   │   └── idgen
│   │       └── uuidgen.go
│   ├── interface
│   │   └── httpi
│   │       ├── auth_handler.go
│   │       ├── order_handler.go
│   │       └── respond.go
│   └── usecase
│       └── order_usecase.go
├── kc-data
│   └── data
│       └── h2
│           ├── keycloakdb.mv.db
│           └── keycloakdb.trace.db
├── Makefile
├── pkg
├── README.md
└── sqlc.yaml
```

## DB, Keycloak, Server 起動

- Docker Desktop を起動する

- マイグレーションを適用する

```
make migrate.up
```

- Keycloak を起動する

```
make keycloak.up
```

- サーバーを起動する

```
make dev
```

## 決済

- OIDC認証をするため、ブラウザで下記URLに登録ユーザーでログインする
- Cookie に保存された access_token を取得する

```
http://localhost:8080/auth/login
```

- 注文を作成する

```
TOKEN=<access_token>

curl -s -i -X POST http://localhost:8080/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount_jpy":1200}'
```

- 注文IDを取得して、注文を支払う

```
curl -i -X POST "http://localhost:8080/orders/<order_id>/pay" \
  -H "Authorization: Bearer $TOKEN"
```

## M2M

- Keycloak に管理者ログインして、 payment-api の SECRET を取得する

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


## アクセストークンを更新する

```
curl -i -X POST http://localhost:8080/auth/refresh \
  -H 'Cookie: rt=<ブラウザから rt 取得>'
```

## ログアウト

```
curl -i -X POST http://localhost:8080/auth/logout \
  -H 'Cookie: rt=<ブラウザから rt 取得>'
```

※ ブラウザで`http://localhost:8080/auth/logout`を開いてもOK

## 削除

- Keycloak を削除する

```
make keycloak.remove
```

- DB を削除する

```
make migrate.remove
```

- volume も削除する

```
make db.down
```
