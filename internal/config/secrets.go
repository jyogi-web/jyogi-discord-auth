package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// DiscordConfig はDiscord関連の設定を保持します
type DiscordConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	GuildID      string `json:"guild_id"`
	JWTSecret    string `json:"jwt_secret"`
	BotToken     string `json:"bot_token,omitempty"`
}

// TiDBConfig はTiDB接続設定を保持します
type TiDBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// ParseDiscordConfig は環境変数からDiscord設定をパースします
func ParseDiscordConfig() (*DiscordConfig, error) {
	jsonStr := os.Getenv("DISCORD_CONFIG")
	if jsonStr == "" {
		// フォールバック: 個別の環境変数から読み込み（後方互換性）
		clientID := os.Getenv("DISCORD_CLIENT_ID")
		clientSecret := os.Getenv("DISCORD_CLIENT_SECRET")
		redirectURI := os.Getenv("DISCORD_REDIRECT_URI")
		guildID := os.Getenv("DISCORD_GUILD_ID")
		jwtSecret := os.Getenv("JWT_SECRET")

		// 必須フィールドのバリデーション
		if clientID == "" {
			return nil, fmt.Errorf("DISCORD_CLIENT_ID is required")
		}
		if clientSecret == "" {
			return nil, fmt.Errorf("DISCORD_CLIENT_SECRET is required")
		}
		if redirectURI == "" {
			return nil, fmt.Errorf("DISCORD_REDIRECT_URI is required")
		}
		if guildID == "" {
			return nil, fmt.Errorf("DISCORD_GUILD_ID is required")
		}
		if jwtSecret == "" {
			return nil, fmt.Errorf("JWT_SECRET is required")
		}

		return &DiscordConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURI:  redirectURI,
			GuildID:      guildID,
			JWTSecret:    jwtSecret,
			BotToken:     os.Getenv("DISCORD_BOT_TOKEN"), // オプション
		}, nil
	}

	var config DiscordConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, fmt.Errorf("failed to parse DISCORD_CONFIG: %w", err)
	}
	return &config, nil
}

// ParseTiDBConfig は環境変数からTiDB設定をパースします
func ParseTiDBConfig() (*TiDBConfig, error) {
	jsonStr := os.Getenv("TIDB_CONFIG")
	if jsonStr == "" {
		// フォールバック: 個別の環境変数から読み込み（後方互換性）
		host := os.Getenv("TIDB_DB_HOST")
		username := os.Getenv("TIDB_DB_USERNAME")
		password := os.Getenv("TIDB_DB_PASSWORD")
		database := os.Getenv("TIDB_DB_DATABASE")

		// 必須フィールドのバリデーション
		if host == "" {
			return nil, fmt.Errorf("TIDB_DB_HOST is required")
		}
		if username == "" {
			return nil, fmt.Errorf("TIDB_DB_USERNAME is required")
		}
		if password == "" {
			return nil, fmt.Errorf("TIDB_DB_PASSWORD is required")
		}
		if database == "" {
			return nil, fmt.Errorf("TIDB_DB_DATABASE is required")
		}

		return &TiDBConfig{
			Host:     host,
			Port:     getEnvInt("TIDB_DB_PORT", 4000),
			Username: username,
			Password: password,
			Database: database,
		}, nil
	}

	var config TiDBConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, fmt.Errorf("failed to parse TIDB_CONFIG: %w", err)
	}
	return &config, nil
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	var result int
	fmt.Sscanf(val, "%d", &result)
	return result
}
