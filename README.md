payment-system/
├── cmd/                  # エントリポイント（main.goを置く）
│   └── api/
│       └── main.go
├── internal/             # アプリケーションの内部実装
│   ├── domain/           # DDD: エンティティ、値オブジェクト、リポジトリ
│   │   ├── order/
│   │   │   └── order.go
│   │   └── payment/
│   │       └── payment.go
│   ├── usecase/          # アプリケーションサービス層
│   │   └── order_usecase.go
│   ├── infra/            # DB、外部API、メッセージングなどの実装
│   │   ├── db/
│   │   │   └── postgres.go
│   │   └── payment_gateway/
│   │       └── stripe_client.go
│   └── interface/        # プレゼンテーション層（APIハンドラー）
│       └── http/
│           └── order_handler.go
├── pkg/                  # 再利用可能なライブラリ（認証、ロギング等）
├── go.mod
└── go.sum
