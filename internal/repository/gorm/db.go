package gorm

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/jyogi-web/jyogi-discord-auth/internal/config"
)

// InitDB はTiDBデータベース接続を初期化します
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	tlsConfig := "true"
	if cfg.TiDBDisableTLS {
		tlsConfig = "false"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s",
		cfg.TiDBUser,
		cfg.TiDBPassword,
		cfg.TiDBHost,
		cfg.TiDBPort,
		cfg.TiDBDatabase,
		tlsConfig,
	)

	// 環境に応じたログレベルを設定
	logLevel := logger.Error
	if cfg.Env == "development" {
		logLevel = logger.Info // 開発環境ではSQL文も出力
	}

	// カスタムロガー設定: RecordNotFound エラーを無視
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logLevel,    // Log level (環境別)
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Don't include params in the SQL log (set to true to hide params)
			Colorful:                  false,       // Disable color
		},
	)

	log.Printf("GORM logger configured with level: %v, SlowThreshold: %v", logLevel, time.Second)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Printf("Failed to connect to TiDB %s@%s:%s/%s: %v",
			cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase, err)
		return nil, fmt.Errorf("failed to connect to TiDB %s@%s:%s/%s: %w",
			cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase, err)
	}

	// コネクションプール設定（sql.DBを取得して設定）
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Failed to get sql.DB from GORM for TiDB %s@%s:%s/%s: %v",
			cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase, err)
		return nil, fmt.Errorf("failed to get sql.DB from GORM (TiDB: %s@%s:%s/%s): %w",
			cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase, err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// AutoMigrate実行
	log.Printf("Starting AutoMigrate for TiDB %s@%s:%s/%s",
		cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase)

	if err := db.AutoMigrate(
		&User{},
		&Session{},
		&ClientApp{},
		&AuthCode{},
		&Token{},
		&Profile{},
	); err != nil {
		// マイグレーション失敗時、DB接続をクローズしてリソースリークを防ぐ
		if sqlDB, dbErr := db.DB(); dbErr == nil {
			sqlDB.Close()
		}
		log.Printf("AutoMigrate failed for TiDB %s@%s:%s/%s: %v",
			cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase, err)
		return nil, fmt.Errorf("failed to migrate schema for TiDB %s@%s:%s/%s: %w",
			cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase, err)
	}

	log.Printf("AutoMigrate completed successfully for TiDB %s@%s:%s/%s",
		cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase)

	log.Printf("TiDB initialized: %s@%s:%s/%s", cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase)

	return db, nil
}
