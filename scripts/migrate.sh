#!/bin/bash

# じょぎメンバー認証システム - データベースマイグレーションスクリプト
# Usage: ./scripts/migrate.sh [up|down]

set -e

# 環境変数の読み込み
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# デフォルトのデータベースパス
DB_PATH="${DATABASE_PATH:-./jyogi_auth.db}"

# コマンド引数（デフォルトはup）
COMMAND="${1:-up}"

# マイグレーションディレクトリ
MIGRATIONS_DIR="./migrations"

if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "❌ Error: Migrations directory not found: $MIGRATIONS_DIR"
    exit 1
fi

# マイグレーションバージョン管理テーブルを作成
init_schema_migrations() {
    sqlite3 "$DB_PATH" << EOF
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
EOF
}

# マイグレーション実行（up）
migrate_up() {
    echo "📦 Running migrations (up)..."
    echo "Database path: $DB_PATH"

    init_schema_migrations

    # .up.sqlファイルを順番に実行
    for migration in "$MIGRATIONS_DIR"/*.up.sql; do
        if [ -f "$migration" ]; then
            # バージョン番号を取得（ファイル名から）
            filename=$(basename "$migration")
            version=$(echo "$filename" | cut -d'_' -f1)

            # すでに適用済みか確認
            already_applied=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM schema_migrations WHERE version = $version;")

            if [ "$already_applied" -eq 0 ]; then
                echo "📝 Applying migration: $filename"
                sqlite3 "$DB_PATH" < "$migration"
                if [ $? -eq 0 ]; then
                    # バージョンを記録
                    sqlite3 "$DB_PATH" "INSERT INTO schema_migrations (version) VALUES ($version);"
                    echo "✅ Successfully applied: $filename"
                else
                    echo "❌ Failed to apply: $filename"
                    exit 1
                fi
            else
                echo "⏭️  Skipping (already applied): $filename"
            fi
        fi
    done

    echo "✅ All migrations completed successfully!"
    
    # Fix permissions for app container (running as uid 1000)
    # Only if running as root (likely in docker)
    if [ "$(id -u)" -eq 0 ]; then
        chown 1000:1000 "$DB_PATH"
        chown 1000:1000 "$(dirname "$DB_PATH")"
        echo "✅ Fixed permissions for appuser (1000:1000)"
    fi

    echo "Database is ready at: $DB_PATH"
}

# マイグレーションロールバック（down）
migrate_down() {
    echo "🔄 Rolling back last migration..."
    echo "Database path: $DB_PATH"

    init_schema_migrations

    # 最後に適用されたバージョンを取得
    last_version=$(sqlite3 "$DB_PATH" "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;")

    if [ -z "$last_version" ]; then
        echo "ℹ️  No migrations to rollback"
        exit 0
    fi

    # 対応する.down.sqlファイルを探す（バージョン番号をゼロパディング）
    padded_version=$(printf "%06d" "$last_version")
    down_file=$(find "$MIGRATIONS_DIR" -name "${padded_version}_*.down.sql" | head -n 1)

    if [ ! -f "$down_file" ]; then
        echo "❌ Error: Down migration file not found for version $last_version"
        exit 1
    fi

    echo "📝 Rolling back: $(basename "$down_file")"
    sqlite3 "$DB_PATH" < "$down_file"

    if [ $? -eq 0 ]; then
        # バージョン記録を削除
        sqlite3 "$DB_PATH" "DELETE FROM schema_migrations WHERE version = $last_version;"
        echo "✅ Successfully rolled back version $last_version"
    else
        echo "❌ Failed to rollback version $last_version"
        exit 1
    fi
}

# マイグレーションステータス表示
migrate_status() {
    echo "📊 Migration status:"
    echo "Database path: $DB_PATH"
    echo ""

    if [ ! -f "$DB_PATH" ]; then
        echo "Database does not exist yet"
        exit 0
    fi

    init_schema_migrations

    echo "Applied migrations:"
    sqlite3 "$DB_PATH" "SELECT version, applied_at FROM schema_migrations ORDER BY version;"
}

# コマンド実行
case "$COMMAND" in
    up)
        migrate_up
        ;;
    down)
        migrate_down
        ;;
    status)
        migrate_status
        ;;
    *)
        echo "Usage: $0 [up|down|status]"
        echo ""
        echo "Commands:"
        echo "  up      - Apply pending migrations"
        echo "  down    - Rollback last migration"
        echo "  status  - Show migration status"
        exit 1
        ;;
esac
