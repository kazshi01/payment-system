# Repository Guidelines

このファイルは本リポジトリで作業するエージェント/開発者向けの合意事項と実務メモです。実際のディレクトリ構成・実装状況に合わせて更新済みです。

## プロジェクト構成（現況）
- `cmd/api/` — エントリポイント（`main.go`）。DI は未配線（後述）。
- `internal/domain/` — ドメイン層
  - `order/` — `Order`, `Status` 等
  - `payment/` — `Payment` 等
  - ルート直下 — `PaymentGateway` インターフェース, `OrderRepository`, `Tx`
- `internal/usecase/` — `OrderUsecase`（`Repo`/`Tx`/`PG` を注入）。`generateID()` はスタブ。
- `internal/interface/http/` — `OrderHandler`（`Create`, `Pay`）。Go 1.22+ の `http.ServeMux` パターンを使用。
- `internal/infra/` — まだ空。今後 DB/外部API 実装を配置予定。
- `db/` — マイグレーションと簡易 Record 変換（`OrderRecord`）。現状はルート直下に置いているが、将来的に `internal/infra/db/` 配下へ移動を検討。
- `pkg/` — 共有ユーティリティ置き場（現状 `.gitkeep` のみ）。

補足:
- README の例示構成にある `internal/infra/db/postgres.go` 等は未実装です。
- `go.mod` は `go 1.25.1` を宣言。`http.ServeMux` のパターンマッチは Go 1.22+ 必須。

## API エンドポイント（現状）
- `POST /orders` — 注文作成（body: `{ "amount_jpy": number }`）
- `POST /orders/{id}/pay` — 注文支払い実行
- `GET /health` — ヘルスチェック

現状 `cmd/api/main.go` の DI が未配線のため、`go run` しても `nil` フィールドで失敗します。先にインフラ実装と配線を行ってください（次節）。

## 実装 TODO（優先度順）
1) インフラ層の実装と DI 配線
   - `internal/infra/db` に `OrderRepository` 実装（Postgres 予定）
   - `internal/infra/tx` に `Tx` 実装（DB トランザクション境界）
   - `internal/infra/payment_gateway/stripe` に `PaymentGateway` 実装
   - `cmd/api/main.go` で上記を生成し、`OrderUsecase` に注入
2) `generateID()` の実装
   - 例: ULID/UUID。ID 生成は単体テスト可能な関数に分離
3) マイグレーション整備と infra/db への集約
   - 既存 `db/` の `OrderRecord` とマイグレーションを `internal/infra/db/` へ移動検討
4) ユースケース／ハンドラの単体テスト追加
   - `OrderUsecase.CreateOrder/PayOrder` のテーブル駆動テスト
   - `OrderHandler` のハンドラテスト（リクエスト/レスポンス検証）
5) README の構成図を現況に合わせて更新

## ビルド・実行
- ビルド: `go build ./cmd/api`
- 実行: `go run ./cmd/api`（DI 配線完了後）
- 静的解析/整形: `go vet ./...` / `gofmt -s -w .`

## データベース & マイグレーション
- 依存: Docker / Docker Compose
- 環境変数: `.env` に以下の例あり（リポジトリ同梱）
  - `POSTGRES_USER=app`
  - `POSTGRES_PASSWORD=password`
  - `POSTGRES_DB=payment`
- 起動/適用:
  - `make migrate.up` — `docker compose` で `db` 起動後、`migrate/migrate` で `db/migrations` を適用
  - `make migrate.down` — Compose を停止しボリューム削除（破壊的）

## コーディング規約・命名
- フォーマット: `gofmt` 準拠（タブインデント）。PR 前に整形必須。
- 命名: パッケージ小文字（例: `payment`）。公開識別子は `PascalCase`、非公開は `camelCase`。ファイル名は小文字スネーク/ハイフンなし推奨。
- エラー: ラップは `fmt.Errorf("...: %w", err)`、判定は `errors.Is/As`。
- 依存の向き: `usecase`/`interface` は `domain` に依存。`infra` は実装を提供し DI で注入。

## テスト指針
- フレームワーク: 標準 `testing`。テーブル駆動テスト推奨。
- 命名: `*_test.go`、`TestXxx`、必要に応じ `t.Run` でサブテスト。
- 目標: 主要ロジックは >80% 目安。外部依存はモック化。
- 例: `go test ./internal/... -run TestOrder`。

## セキュリティ & 設定
- 機微情報をリポジトリに含めない（API キー等）。
- ログに秘密情報を出力しない。資格情報は `.env` またはシークレットマネージャで管理。
- 依存更新時は `go.mod`/`go.sum` をコミットし、再現性を確認。
- 決済系の外部通信を実装する場合は、テスト用キー/エンドポイントを利用し本番キーの混入を禁止。

## 補助メモ（実装参照）
- `internal/interface/http/order_handler.go` は `r.PathValue` を使用（Go 1.22+）。
- `internal/usecase/order_usecase.go` の `PayOrder` は `Tx.Do` に依存し、トランザクション境界内で `Repo.Update` を実行。
- DB スキーマは `db/migrations/0001_init.*.sql` を参照。`orders`/`payments`/`payment_events` を作成。
