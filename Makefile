.PHONY: help build run test clean docker-build docker-up docker-down migrate-up migrate-down migrate-create fmt vet sync-profiles

# デフォルトのヘルプコマンド
help:
	@echo "じょぎメンバー認証システム - 開発コマンド"
	@echo ""
	@echo "利用可能なコマンド:"
	@echo "  make build          - サーバーをビルド"
	@echo "  make run            - サーバーを起動"
	@echo "  make test           - テストを実行"
	@echo "  make fmt            - コードをフォーマット"
	@echo "  make vet            - コードを静的解析"
	@echo "  make clean          - ビルド成果物を削除"
	@echo ""
	@echo "ローカル コマンド:"
	@echo "  make build-local    - サーバーをビルド（ローカル）"
	@echo "  make run-local      - サーバーを起動（ローカル）"
	@echo ""
	@echo "プロフィール同期 コマンド:"
	@echo "  make sync-profiles  - プロフィールを1回同期"
	@echo "  make sync-profiles-daemon - プロフィールを定期的に同期"
	@echo ""
	@echo "Docker コマンド:"
	@echo "  make docker-build   - Dockerイメージをビルド"
	@echo "  make docker-up      - Docker Composeで起動"
	@echo "  make docker-down    - Docker Composeで停止"
	@echo "  make docker-logs    - Dockerログを表示"
	@echo ""
	@echo "マイグレーション コマンド:"
	@echo "  make migrate-up     - マイグレーションを適用"
	@echo "  make migrate-down   - マイグレーションをロールバック"
	@echo "  make migrate-create NAME=<name> - 新しいマイグレーションを作成"

# サーバー起動
run:
	@echo "🚀 Starting server..."
	make docker-up
	@echo "✅ Server started!"

build:
	@echo "🔨 Building server..."
	make docker-build
	@echo "✅ Build complete!"

# ビルド
local-build:
	@echo "🔨 Building server..."
	go build -o bin/server ./cmd/server
	@echo "✅ Build complete!"

# サーバー起動
local-run:
	@echo "🚀 Starting server..."
	go run ./cmd/server

# テスト実行
# テスト実行
test:
	@echo "🧪 Running tests in Docker..."
	docker-compose run --rm dev test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "✅ Tests complete!"

# ローカルでのテスト実行
test-local:
	@echo "🧪 Running tests locally..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "✅ Tests complete!"

# コードフォーマット
fmt:
	@echo "🎨 Formatting code..."
	gofmt -s -w .
	@echo "✅ Format complete!"

# 静的解析
vet:
	@echo "🔍 Running go vet..."
	go vet ./...
	@echo "✅ Vet complete!"

# クリーンアップ
clean:
	@echo "🧹 Cleaning up..."
	rm -rf bin/
	rm -f *.db
	rm -f coverage.txt
	@echo "✅ Clean complete!"

# Docker ビルド
docker-build:
	@echo "🐳 Building Docker image..."
	docker-compose build
	@echo "✅ Docker build complete!"

# Docker Compose 起動
docker-up:
	@echo "🐳 Starting Docker Compose..."
	docker-compose up
	@echo "✅ Docker Compose started!"
	@echo "📝 View logs with: make docker-logs"

# Docker Compose 停止
docker-down:
	@echo "🐳 Stopping Docker Compose..."
	docker-compose down
	@echo "✅ Docker Compose stopped!"

# Docker ログ表示
docker-logs:
	docker-compose logs -f

# マイグレーション実行（up）
migrate-up:
	@echo "🚀 Running migrations..."
	@./scripts/migrate.sh up
	@echo "✅ Migrations complete!"

# マイグレーション実行（down）
migrate-down:
	@echo "🚀 Running migrations..."
	@./scripts/migrate.sh down
	@echo "✅ Migrations complete!"

# マイグレーションステータス確認
migrate-status:
	@echo "🚀 Running migrations..."
	@./scripts/migrate.sh status
	@echo "✅ Migrations complete!"

# 新しいマイグレーション作成
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Error: NAME is required. Usage: make migrate-create NAME=add_users_table"; \
		exit 1; \
	fi
	@NEXT_VERSION=$$(ls migrations/*.up.sql 2>/dev/null | wc -l | tr -d ' '); \
	NEXT_VERSION=$$((NEXT_VERSION + 1)); \
	NEXT_VERSION=$$(printf "%06d" $$NEXT_VERSION); \
	touch "migrations/$${NEXT_VERSION}_$(NAME).up.sql"; \
	touch "migrations/$${NEXT_VERSION}_$(NAME).down.sql"; \
	echo "✅ Created migration files:"; \
	echo "  - migrations/$${NEXT_VERSION}_$(NAME).up.sql"; \
	echo "  - migrations/$${NEXT_VERSION}_$(NAME).down.sql"

# 開発環境セットアップ
setup:
	@echo "🔧 Setting up development environment..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "✅ Created .env file from .env.example"; \
	else \
		echo "ℹ️  .env file already exists"; \
	fi
	@echo "📦 Installing dependencies..."
	go mod download
	@echo "✅ Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Edit .env file with your configuration"
	@echo "  2. Run 'make migrate-up' to create database schema"
	@echo "  3. Run 'make run' to start the server"

# プロフィール同期（1回）
sync-profiles:
	@echo "🔄 Syncing profiles once..."
	go run ./cmd/sync-profiles -once
	@echo "✅ Profile sync complete!"

# プロフィール同期（定期実行）
sync-profiles-daemon:
	@echo "🔄 Starting profile sync daemon..."
	go run ./cmd/sync-profiles
	@echo "✅ Profile sync daemon stopped!"

# プロフィール同期ビルド
build-sync-profiles:
	@echo "🔨 Building sync-profiles..."
	go build -o bin/sync-profiles ./cmd/sync-profiles
	@echo "✅ Build complete!"
