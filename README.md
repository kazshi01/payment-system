# Payment System

## ディレクトリ構成

```
.
├── cmd
│   └── api
│       ├── main.go
│       └── openapi.yaml
├── db
│   ├── migrate.sh
│   ├── migrations
│   │   ├── 0001_init.down.sql
│   │   └── 0001_init.up.sql
│   └── order_record.go
├── docker-compose.yaml
├── go.mod
├── go.sum
├── internal
│   ├── auth
│   │   ├── middleware.go
│   │   └── oidc-pkce.go
│   ├── docs
│   │   └── swagger.go
│   ├── domain
│   │   ├── errors.go
│   │   ├── locker.go
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
│   │   ├── idgen
│   │   │   └── uuidgen.go
│   │   └── redislocker
│   │       └── locker.go
│   ├── interface
│   │   └── httpi
│   │       ├── auth_handler.go
│   │       ├── home.go
│   │       ├── order_handler.go
│   │       └── respond.go
│   └── usecase
│       ├── order_usecase_test.go
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

## テスト実行

```
make test
```
<details>
<summary>実行結果</summary>

```
% make test
EMPTY internal/docs
EMPTY internal/domain
EMPTY internal/domain/payment
        github.com/kazshi01/payment-system/internal/infra/clock         coverage: 0.0% of statements
EMPTY internal/infra/clock (coverage: 0.0% of statements)
        github.com/kazshi01/payment-system/internal/auth                coverage: 0.0% of statements
EMPTY internal/auth (coverage: 0.0% of statements)
        github.com/kazshi01/payment-system/internal/infra/db/sqlc               coverage: 0.0% of statements
EMPTY internal/infra/db/sqlc (coverage: 0.0% of statements)
        github.com/kazshi01/payment-system/internal/infra/db            coverage: 0.0% of statements
EMPTY internal/infra/db (coverage: 0.0% of statements)
        github.com/kazshi01/payment-system/internal/infra/db/pg         coverage: 0.0% of statements
        github.com/kazshi01/payment-system/cmd/api              coverage: 0.0% of statements
EMPTY internal/infra/db/pg (coverage: 0.0% of statements)
EMPTY cmd/api (coverage: 0.0% of statements)
        github.com/kazshi01/payment-system/internal/infra/idgen         coverage: 0.0% of statements
EMPTY internal/infra/idgen (coverage: 0.0% of statements)
        github.com/kazshi01/payment-system/internal/infra/db/dbmodel            coverage: 0.0% of statements
EMPTY internal/infra/db/dbmodel (coverage: 0.0% of statements)
        github.com/kazshi01/payment-system/db           coverage: 0.0% of statements
EMPTY db (coverage: 0.0% of statements)
        github.com/kazshi01/payment-system/internal/domain/order                coverage: 0.0% of statements
EMPTY internal/domain/order (coverage: 0.0% of statements)
PASS internal/usecase.TestOrderUsecase_CreateOrder_ok (0.00s)
PASS internal/usecase.TestOrderUsecase_PayOrder_ok (0.00s)
PASS internal/usecase.TestOrderUsecase_CreateOrder_invalidAmount (0.00s)
PASS internal/usecase.TestOrderUsecase_PayOrder_notFound (0.00s)
PASS internal/usecase.TestOrderUsecase_PayOrder_wrongUser (0.00s)
PASS internal/usecase.TestOrderUsecase_PayOrder_alreadyPaid (0.00s)
coverage: 80.7% of statements
PASS internal/usecase (cached) (coverage: 80.7% of statements)
        github.com/kazshi01/payment-system/internal/infra/redislocker           coverage: 0.0% of statements
EMPTY internal/infra/redislocker (coverage: 0.0% of statements)
        github.com/kazshi01/payment-system/internal/interface/httpi             coverage: 0.0% of statements
EMPTY internal/interface/httpi (coverage: 0.0% of statements)

DONE 6 tests in 0.230s
%
```

</details>

## DB、Redis、Keycloak、Go Server 起動

- Docker Desktop を起動する

- マイグレーションを適用する

```
make migrate.up
```

- Redis を起動する

```
make redis.up
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

### terminal で実行

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

### Swagger UI

- ブラウザで下記にアクセスする

```
http://localhost:8080/docs
```
- 注文を作成する（Create order）

```
Authorize ボタンをクリックして、ブラウザから取得した access_token を登録する
Try it out ボタンをクリックして、任意の amount_jpy を入力して、Execute ボタンをクリックする
```

- 注文を支払う（Pay order）

```
Try it out ボタンをクリックして、注文 id を入力して、Execute ボタンをクリックする
```

- 

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
make keycloak.down
```

- Redis を削除する

```
make redis.down
```

- DB を削除する

```
make migrate.down
```

- volume も削除する

```
make db.remove
```
