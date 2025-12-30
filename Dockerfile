# 開発ステージ
FROM golang:1.23-alpine AS dev
WORKDIR /app
RUN apk add --no-cache git build-base gcc musl-dev sqlite-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENTRYPOINT ["go"]

# マルチステージビルド: ビルドステージ
FROM golang:1.23-alpine AS builder

# 必要なパッケージをインストール（SQLiteビルドに必要なgccとmusl-devを追加）
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev

# 作業ディレクトリを設定
WORKDIR /app

# Go modulesの依存関係をコピーしてダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド（CGO有効でSQLiteサポート）
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags='-w -s -extldflags "-static"' -tags 'osusergo netgo sqlite_omit_load_extension' -o server ./cmd/server

# 本番ステージ: 軽量なイメージ
FROM alpine:latest

# 必要なパッケージをインストール
RUN apk --no-cache add ca-certificates tzdata wget

# 作業ディレクトリを設定
WORKDIR /app

# ビルドステージからバイナリをコピー
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

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
