-- じょぎメンバー認証システム - ロールバック
-- Created: 2025-12-22

-- テーブルを逆順で削除（外部キー制約を考慮）
DROP INDEX IF EXISTS idx_tokens_expires_at;
DROP INDEX IF EXISTS idx_tokens_client_id;
DROP INDEX IF EXISTS idx_tokens_user_id;
DROP INDEX IF EXISTS idx_tokens_token;
DROP TABLE IF EXISTS tokens;

DROP INDEX IF EXISTS idx_auth_codes_expires_at;
DROP INDEX IF EXISTS idx_auth_codes_user_id;
DROP INDEX IF EXISTS idx_auth_codes_code;
DROP TABLE IF EXISTS auth_codes;

DROP INDEX IF EXISTS idx_client_apps_client_id;
DROP TABLE IF EXISTS client_apps;

DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_token;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP TABLE IF EXISTS sessions;

DROP INDEX IF EXISTS idx_users_discord_id;
DROP TABLE IF EXISTS users;
