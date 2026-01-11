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
		return &DiscordConfig{
			ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
			ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
			RedirectURI:  os.Getenv("DISCORD_REDIRECT_URI"),
			GuildID:      os.Getenv("DISCORD_GUILD_ID"),
			JWTSecret:    os.Getenv("JWT_SECRET"),
			BotToken:     os.Getenv("DISCORD_BOT_TOKEN"),
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
		return &TiDBConfig{
			Host:     os.Getenv("TIDB_DB_HOST"),
			Port:     getEnvInt("TIDB_DB_PORT", 4000),
			Username: os.Getenv("TIDB_DB_USERNAME"),
			Password: os.Getenv("TIDB_DB_PASSWORD"),
			Database: os.Getenv("TIDB_DB_DATABASE"),
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
