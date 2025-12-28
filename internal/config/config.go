package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config はアプリケーションの全設定を保持します
type Config struct {
	// Discord OAuth2
	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURI  string
	DiscordGuildID      string

	// Discord Bot
	DiscordBotToken        string
	DiscordProfileChannel  string

	// JWT
	JWTSecret string

	// Database
	DatabasePath string

	// Server
	ServerPort string
	HTTPSOnly  bool

	// Environment
	Env string
}

// Load は環境変数から設定を読み込みます
// 開発環境では.envファイルも読み込みます
func Load() (*Config, error) {
	// 開発環境では.envファイルを読み込む
	env := os.Getenv("ENV")
	if env != "production" {
		if err := godotenv.Load(); err != nil {
			// .envファイルはオプショナルなのでエラーを返さない
			// 見つからなかったことをログに記録するだけ
			fmt.Println("Warning: .env file not found, using environment variables")
		}
	}

	cfg := &Config{
		DiscordClientID:     os.Getenv("DISCORD_CLIENT_ID"),
		DiscordClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		DiscordRedirectURI:  os.Getenv("DISCORD_REDIRECT_URI"),
		DiscordGuildID:      os.Getenv("DISCORD_GUILD_ID"),
		DiscordBotToken:     os.Getenv("DISCORD_BOT_TOKEN"),
		DiscordProfileChannel: os.Getenv("DISCORD_PROFILE_CHANNEL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		DatabasePath:        os.Getenv("DATABASE_PATH"),
		ServerPort:          os.Getenv("SERVER_PORT"),
		Env:                 os.Getenv("ENV"),
	}

	// HTTPS_ONLYをbooleanとしてパース
	httpsOnly, err := strconv.ParseBool(os.Getenv("HTTPS_ONLY"))
	if err != nil {
		// 設定されていないか不正な場合はfalseをデフォルトとする
		cfg.HTTPSOnly = false
	} else {
		cfg.HTTPSOnly = httpsOnly
	}

	// デフォルト値を設定
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = "./jyogi_auth.db"
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = "8080"
	}
	if cfg.Env == "" {
		cfg.Env = "development"
	}

	// 必須フィールドを検証
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate は必須設定がすべて存在することを確認します
func (c *Config) Validate() error {
	if c.DiscordClientID == "" {
		return fmt.Errorf("DISCORD_CLIENT_ID is required")
	}
	if c.DiscordClientSecret == "" {
		return fmt.Errorf("DISCORD_CLIENT_SECRET is required")
	}
	if c.DiscordRedirectURI == "" {
		return fmt.Errorf("DISCORD_REDIRECT_URI is required")
	}
	if c.DiscordGuildID == "" {
		return fmt.Errorf("DISCORD_GUILD_ID is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	return nil
}
