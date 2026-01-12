package config

import (
	"strings"
	"testing"
)

// TestParseDiscordConfig はParseDiscordConfig関数をテストします
func TestParseDiscordConfig(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *DiscordConfig
		wantErr string
	}{
		{
			name: "Valid JSON",
			envVars: map[string]string{
				"DISCORD_CONFIG": `{
					"client_id": "test-client-id",
					"client_secret": "test-client-secret",
					"redirect_uri": "http://localhost:8080/callback",
					"guild_id": "test-guild-id",
					"jwt_secret": "test-jwt-secret",
					"bot_token": "test-bot-token"
				}`,
			},
			want: &DiscordConfig{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				RedirectURI:  "http://localhost:8080/callback",
				GuildID:      "test-guild-id",
				JWTSecret:    "test-jwt-secret",
				BotToken:     "test-bot-token",
			},
			wantErr: "",
		},
		{
			name: "Valid Fallback",
			envVars: map[string]string{
				"DISCORD_CLIENT_ID":     "fallback-client-id",
				"DISCORD_CLIENT_SECRET": "fallback-client-secret",
				"DISCORD_REDIRECT_URI":  "http://localhost:8080/callback",
				"DISCORD_GUILD_ID":      "fallback-guild-id",
				"JWT_SECRET":            "fallback-jwt-secret",
				"DISCORD_BOT_TOKEN":     "fallback-bot-token",
			},
			want: &DiscordConfig{
				ClientID:     "fallback-client-id",
				ClientSecret: "fallback-client-secret",
				RedirectURI:  "http://localhost:8080/callback",
				GuildID:      "fallback-guild-id",
				JWTSecret:    "fallback-jwt-secret",
				BotToken:     "fallback-bot-token",
			},
			wantErr: "",
		},
		{
			name: "Invalid JSON",
			envVars: map[string]string{
				"DISCORD_CONFIG": `{invalid json}`,
			},
			want:    nil,
			wantErr: "failed to parse DISCORD_CONFIG",
		},
		{
			name: "Missing ClientID",
			envVars: map[string]string{
				"DISCORD_CLIENT_SECRET": "test-client-secret",
				"DISCORD_REDIRECT_URI":  "http://localhost:8080/callback",
				"DISCORD_GUILD_ID":      "test-guild-id",
				"JWT_SECRET":            "test-jwt-secret",
			},
			want:    nil,
			wantErr: "DISCORD_CLIENT_ID is required",
		},
		{
			name: "Missing ClientSecret",
			envVars: map[string]string{
				"DISCORD_CLIENT_ID":    "test-client-id",
				"DISCORD_REDIRECT_URI": "http://localhost:8080/callback",
				"DISCORD_GUILD_ID":     "test-guild-id",
				"JWT_SECRET":           "test-jwt-secret",
			},
			want:    nil,
			wantErr: "DISCORD_CLIENT_SECRET is required",
		},
		{
			name: "Missing RedirectURI",
			envVars: map[string]string{
				"DISCORD_CLIENT_ID":     "test-client-id",
				"DISCORD_CLIENT_SECRET": "test-client-secret",
				"DISCORD_GUILD_ID":      "test-guild-id",
				"JWT_SECRET":            "test-jwt-secret",
			},
			want:    nil,
			wantErr: "DISCORD_REDIRECT_URI is required",
		},
		{
			name: "Missing GuildID",
			envVars: map[string]string{
				"DISCORD_CLIENT_ID":     "test-client-id",
				"DISCORD_CLIENT_SECRET": "test-client-secret",
				"DISCORD_REDIRECT_URI":  "http://localhost:8080/callback",
				"JWT_SECRET":            "test-jwt-secret",
			},
			want:    nil,
			wantErr: "DISCORD_GUILD_ID is required",
		},
		{
			name: "Missing JWTSecret",
			envVars: map[string]string{
				"DISCORD_CLIENT_ID":     "test-client-id",
				"DISCORD_CLIENT_SECRET": "test-client-secret",
				"DISCORD_REDIRECT_URI":  "http://localhost:8080/callback",
				"DISCORD_GUILD_ID":      "test-guild-id",
			},
			want:    nil,
			wantErr: "JWT_SECRET is required",
		},
		{
			name: "Optional BotToken",
			envVars: map[string]string{
				"DISCORD_CLIENT_ID":     "test-client-id",
				"DISCORD_CLIENT_SECRET": "test-client-secret",
				"DISCORD_REDIRECT_URI":  "http://localhost:8080/callback",
				"DISCORD_GUILD_ID":      "test-guild-id",
				"JWT_SECRET":            "test-jwt-secret",
				// BotToken not set
			},
			want: &DiscordConfig{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				RedirectURI:  "http://localhost:8080/callback",
				GuildID:      "test-guild-id",
				JWTSecret:    "test-jwt-secret",
				BotToken:     "", // オプション
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数をクリア（他のテストの影響を排除）
			t.Setenv("DISCORD_CONFIG", "")
			t.Setenv("DISCORD_CLIENT_ID", "")
			t.Setenv("DISCORD_CLIENT_SECRET", "")
			t.Setenv("DISCORD_REDIRECT_URI", "")
			t.Setenv("DISCORD_GUILD_ID", "")
			t.Setenv("JWT_SECRET", "")
			t.Setenv("DISCORD_BOT_TOKEN", "")

			// テストケースの環境変数を設定
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			got, err := ParseDiscordConfig()

			// エラーチェック
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// 値チェック
			if got.ClientID != tt.want.ClientID {
				t.Errorf("ClientID = %q, want %q", got.ClientID, tt.want.ClientID)
			}
			if got.ClientSecret != tt.want.ClientSecret {
				t.Errorf("ClientSecret = %q, want %q", got.ClientSecret, tt.want.ClientSecret)
			}
			if got.RedirectURI != tt.want.RedirectURI {
				t.Errorf("RedirectURI = %q, want %q", got.RedirectURI, tt.want.RedirectURI)
			}
			if got.GuildID != tt.want.GuildID {
				t.Errorf("GuildID = %q, want %q", got.GuildID, tt.want.GuildID)
			}
			if got.JWTSecret != tt.want.JWTSecret {
				t.Errorf("JWTSecret = %q, want %q", got.JWTSecret, tt.want.JWTSecret)
			}
			if got.BotToken != tt.want.BotToken {
				t.Errorf("BotToken = %q, want %q", got.BotToken, tt.want.BotToken)
			}
		})
	}
}

// TestParseTiDBConfig はParseTiDBConfig関数をテストします
func TestParseTiDBConfig(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *TiDBConfig
		wantErr string
	}{
		{
			name: "Valid JSON",
			envVars: map[string]string{
				"TIDB_CONFIG": `{
					"host": "test-host",
					"port": 3306,
					"username": "test-user",
					"password": "test-password",
					"database": "test-db"
				}`,
			},
			want: &TiDBConfig{
				Host:     "test-host",
				Port:     3306,
				Username: "test-user",
				Password: "test-password",
				Database: "test-db",
			},
			wantErr: "",
		},
		{
			name: "Valid Fallback",
			envVars: map[string]string{
				"TIDB_DB_HOST":     "fallback-host",
				"TIDB_DB_PORT":     "3306",
				"TIDB_DB_USERNAME": "fallback-user",
				"TIDB_DB_PASSWORD": "fallback-password",
				"TIDB_DB_DATABASE": "fallback-db",
			},
			want: &TiDBConfig{
				Host:     "fallback-host",
				Port:     3306,
				Username: "fallback-user",
				Password: "fallback-password",
				Database: "fallback-db",
			},
			wantErr: "",
		},
		{
			name: "Invalid JSON",
			envVars: map[string]string{
				"TIDB_CONFIG": `{invalid json}`,
			},
			want:    nil,
			wantErr: "failed to parse TIDB_CONFIG",
		},
		{
			name: "Missing Host",
			envVars: map[string]string{
				"TIDB_DB_USERNAME": "test-user",
				"TIDB_DB_PASSWORD": "test-password",
				"TIDB_DB_DATABASE": "test-db",
			},
			want:    nil,
			wantErr: "invalid TiDB config: host is required",
		},
		{
			name: "Missing Username",
			envVars: map[string]string{
				"TIDB_DB_HOST":     "test-host",
				"TIDB_DB_PASSWORD": "test-password",
				"TIDB_DB_DATABASE": "test-db",
			},
			want:    nil,
			wantErr: "invalid TiDB config: username is required",
		},
		{
			name: "Missing Password",
			envVars: map[string]string{
				"TIDB_DB_HOST":     "test-host",
				"TIDB_DB_USERNAME": "test-user",
				"TIDB_DB_DATABASE": "test-db",
			},
			want:    nil,
			wantErr: "invalid TiDB config: password is required",
		},
		{
			name: "Missing Database",
			envVars: map[string]string{
				"TIDB_DB_HOST":     "test-host",
				"TIDB_DB_USERNAME": "test-user",
				"TIDB_DB_PASSWORD": "test-password",
			},
			want:    nil,
			wantErr: "invalid TiDB config: database is required",
		},
		{
			name: "Default Port",
			envVars: map[string]string{
				"TIDB_DB_HOST":     "test-host",
				"TIDB_DB_USERNAME": "test-user",
				"TIDB_DB_PASSWORD": "test-password",
				"TIDB_DB_DATABASE": "test-db",
				// Port not set
			},
			want: &TiDBConfig{
				Host:     "test-host",
				Port:     4000, // デフォルト値
				Username: "test-user",
				Password: "test-password",
				Database: "test-db",
			},
			wantErr: "",
		},
		{
			name: "Custom Port",
			envVars: map[string]string{
				"TIDB_DB_HOST":     "test-host",
				"TIDB_DB_PORT":     "3307",
				"TIDB_DB_USERNAME": "test-user",
				"TIDB_DB_PASSWORD": "test-password",
				"TIDB_DB_DATABASE": "test-db",
			},
			want: &TiDBConfig{
				Host:     "test-host",
				Port:     3307,
				Username: "test-user",
				Password: "test-password",
				Database: "test-db",
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数をクリア（他のテストの影響を排除）
			t.Setenv("TIDB_CONFIG", "")
			t.Setenv("TIDB_DB_HOST", "")
			t.Setenv("TIDB_DB_PORT", "")
			t.Setenv("TIDB_DB_USERNAME", "")
			t.Setenv("TIDB_DB_PASSWORD", "")
			t.Setenv("TIDB_DB_DATABASE", "")

			// テストケースの環境変数を設定
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			got, err := ParseTiDBConfig()

			// エラーチェック
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// 値チェック
			if got.Host != tt.want.Host {
				t.Errorf("Host = %q, want %q", got.Host, tt.want.Host)
			}
			if got.Port != tt.want.Port {
				t.Errorf("Port = %d, want %d", got.Port, tt.want.Port)
			}
			if got.Username != tt.want.Username {
				t.Errorf("Username = %q, want %q", got.Username, tt.want.Username)
			}
			if got.Password != tt.want.Password {
				t.Errorf("Password = %q, want %q", got.Password, tt.want.Password)
			}
			if got.Database != tt.want.Database {
				t.Errorf("Database = %q, want %q", got.Database, tt.want.Database)
			}
		})
	}
}
