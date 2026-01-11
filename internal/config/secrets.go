package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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

// Validate は必須のシークレットフィールドを検証します
func (c *DiscordConfig) Validate() error {
	if c.ClientSecret == "" {
		return fmt.Errorf("client_secret is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("jwt_secret is required")
	}
	if c.BotToken == "" {
		return fmt.Errorf("bot_token is required")
	}
	return nil
}

// TiDBConfig はTiDB接続設定を保持します
type TiDBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// Validate は必須のTiDB接続設定フィールドを検証します
func (c *TiDBConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Username == "" {
		return fmt.Errorf("username is required")
	}
	if c.Database == "" {
		return fmt.Errorf("database is required")
	}
	if c.Password == "" {
		return fmt.Errorf("password is required")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}
	return nil
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
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Discord config: %w", err)
	}
	return &config, nil
}

// ParseTiDBConfig は環境変数からTiDB設定をパースします
func ParseTiDBConfig() (*TiDBConfig, error) {
	jsonStr := os.Getenv("TIDB_CONFIG")
	if jsonStr == "" {
		// フォールバック: 個別の環境変数から読み込み（後方互換性）
		config := &TiDBConfig{
			Host:     os.Getenv("TIDB_DB_HOST"),
			Port:     getEnvInt("TIDB_DB_PORT", 4000),
			Username: os.Getenv("TIDB_DB_USERNAME"),
			Password: os.Getenv("TIDB_DB_PASSWORD"),
			Database: os.Getenv("TIDB_DB_DATABASE"),
		}
		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("invalid TiDB config: %w", err)
		}
		return config, nil
	}

	var config TiDBConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, fmt.Errorf("failed to parse TIDB_CONFIG: %w", err)
	}
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid TiDB config: %w", err)
	}
	return &config, nil
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	result, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return result
}
