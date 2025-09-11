# Repository Guidelines

## プロジェクト構成 & モジュール
- `cmd/api/` — エントリポイント（`main.go`）。
- `internal/domain/` — エンティティ/値オブジェクト/リポジトリ（例: `order/`, `payment/`）。
- `internal/usecase/` — アプリケーションサービス層（ユースケース）。
- `internal/interface/http/` — プレゼンテーション層（ハンドラ）。
- `internal/infra/` — DB・外部API 実装（例: `db/postgres.go`, `payment_gateway/stripe_client.go`）。
- `pkg/` — 再利用可能ユーティリティ。

## ビルド・実行・テスト
- `go build ./cmd/api` — API バイナリをビルド。
- `go run ./cmd/api` — ローカルで起動（環境変数が必要な場合あり）。
- `go test ./...` — すべてのテストを実行。
- `go test ./... -cover` — カバレッジ計測。
- `go vet ./...` / `gofmt -s -w .` — 静的解析とコード整形。

## コーディング規約・命名
- フォーマット: `gofmt` 準拠（タブインデント）。PR 前に整形必須。
- 命名: パッケージ小文字（例: `payment`）。公開識別子は `PascalCase`、非公開は `camelCase`。ファイル名は小文字スネーク/ハイフンなしを推奨。
- エラー: ラップは `fmt.Errorf("...: %w", err)`、判定は `errors.Is/As`。
- 依存の向き: `usecase`/`interface` は `domain` に依存。`infra` は実装を提供し DI で注入。

## テスト指針
- フレームワーク: 標準 `testing`。テーブル駆動テスト推奨。
- 命名: `*_test.go`、`TestXxx`、必要に応じ `t.Run` でサブテスト。
- 目標: 変更箇所は主要ロジックで >80% 目安。外部依存はモック化。
- 例: `go test ./internal/... -run TestOrder`。

## コミット & PR
- コミット: 簡潔な要約（現在形、和文可）。必要なら本文で背景/方針/影響を記述。関連 Issue は `#123` を記載。小さく原子的に分割。
- PR 必須事項: 目的/背景、変更点、動作確認手順、影響範囲、関連 Issue、必要なら API リクエスト例やスクリーンショット。`go build`/`go test` が通過していること。

## セキュリティ & 設定
- 環境変数例: `DATABASE_URL=postgres://...`、`STRIPE_API_KEY=sk_test_...`。機微情報はリポジトリに含めない。
- ログに秘密情報を出力しない。資格情報は `.env` またはシークレットマネージャで管理。
- 依存更新時は `go.mod`/`go.sum` をコミットし、再現性を確認。

