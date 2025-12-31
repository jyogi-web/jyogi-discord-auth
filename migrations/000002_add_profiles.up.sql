-- プロフィールテーブル
-- Discord自己紹介チャンネルから取得した情報を保存

CREATE TABLE IF NOT EXISTS profiles (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    discord_message_id TEXT UNIQUE NOT NULL,
    real_name TEXT,
    student_id TEXT,
    hobbies TEXT,
    what_to_do TEXT,
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_profiles_user_id ON profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_profiles_discord_message_id ON profiles(discord_message_id);
