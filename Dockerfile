# 開発ステージ
FROM golang:1.25-alpine AS dev
WORKDIR /app
RUN apk add --no-cache git build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENTRYPOINT ["go"]

# マルチステージビルド: ビルドステージ
FROM golang:1.25-alpine AS builder

# 必要なパッケージをインストール
RUN apk add --no-cache git ca-certificates tzdata

# 作業ディレクトリを設定
WORKDIR /app

# Go modulesの依存関係をコピーしてダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド（CGO無効化で完全静的リンク）
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-w -s -extldflags "-static"' -tags 'osusergo netgo' -o server ./cmd/server

# 本番ステージ: 軽量なイメージ
FROM alpine:latest

# 必要なパッケージをインストール（MySQLクライアントはデバッグ用に便利だが必須ではない。sqliteは削除）
RUN apk --no-cache add ca-certificates tzdata wget bash

# 作業ディレクトリを設定
WORKDIR /app

# ビルドステージからバイナリをコピー
COPY --from=builder /app/server .
# migrationsディレクトリはAutoMigrateを使うので不要だが、参照用にあると便利かもしれない。
# しかし、main.goで埋め込まれていない限りバイナリからは読めない。GORM AutoMigrateはコード定義から生成するのでSQLファイルは不要。
# 念のためscriptsだけコピー
COPY scripts ./scripts

# スクリプトに実行権限を付与
RUN chmod +x ./scripts/*.sh

# 非rootユーザーを作成
RUN addgroup -g 1000 appuser && \
  adduser -D -u 1000 -G appuser appuser && \
  chown -R appuser:appuser /app

# 非rootユーザーに切り替え
USER appuser

# ポートを公開
EXPOSE 8080

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# サーバーを起動
CMD ["./server"]
