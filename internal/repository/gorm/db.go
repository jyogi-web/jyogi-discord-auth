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

	// AutoMigrate専用のサイレントロガー（起動時のスキーマチェッククエリを非表示）
	silentLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent, // すべてのクエリを非表示
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			Colorful:                  false,
		},
	)

	log.Printf("GORM logger configured with level: %v, SlowThreshold: %v", logLevel, time.Second)

	// まずデータベース名なしで接続してデータベースを作成
	dsnWithoutDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local&tls=%s",
		cfg.TiDBUser,
		cfg.TiDBPassword,
		cfg.TiDBHost,
		cfg.TiDBPort,
		tlsConfig,
	)

	tempDB, err := gorm.Open(mysql.Open(dsnWithoutDB), &gorm.Config{
		Logger: silentLogger,
	})
	if err != nil {
		log.Printf("Failed to connect to TiDB %s@%s:%s: %v",
			cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, err)
		return nil, fmt.Errorf("failed to connect to TiDB %s@%s:%s: %w",
			cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, err)
	}

	// データベースを作成（存在しない場合のみ）
	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.TiDBDatabase)
	if err := tempDB.Exec(createDBSQL).Error; err != nil {
		log.Printf("Failed to create database %s on TiDB %s@%s:%s: %v",
			cfg.TiDBDatabase, cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, err)
		return nil, fmt.Errorf("failed to create database %s on TiDB %s@%s:%s: %w",
			cfg.TiDBDatabase, cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, err)
	}
	log.Printf("Database %s ensured on TiDB %s@%s:%s", cfg.TiDBDatabase, cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort)

	// 一時接続をクローズ
	tempSQLDB, dbErr := tempDB.DB()
	if dbErr != nil {
		log.Printf("Failed to get sql.DB from temp connection for TiDB %s@%s:%s: %v",
			cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, dbErr)
		return nil, fmt.Errorf("failed to get sql.DB from temp connection (TiDB: %s@%s:%s): %w",
			cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, dbErr)
	}
	if closeErr := tempSQLDB.Close(); closeErr != nil {
		log.Printf("Failed to close temp connection for TiDB %s@%s:%s: %v",
			cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, closeErr)
		// Close失敗は警告のみで処理を続行
	}

	// データベースを指定して再接続
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s",
		cfg.TiDBUser,
		cfg.TiDBPassword,
		cfg.TiDBHost,
		cfg.TiDBPort,
		cfg.TiDBDatabase,
		tlsConfig,
	)

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

	// AutoMigrate実行（環境変数で制御可能）
	if cfg.DisableAutoMigrate {
		log.Printf("AutoMigrate is disabled for TiDB %s@%s:%s/%s (DISABLE_AUTO_MIGRATE=true)",
			cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase)
	} else {
		log.Printf("Starting AutoMigrate for TiDB %s@%s:%s/%s",
			cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase)

		// AutoMigrate中はサイレントロガーを使用（スキーマチェッククエリを非表示）
		dbWithSilentLogger := db.Session(&gorm.Session{Logger: silentLogger})

		if err := dbWithSilentLogger.AutoMigrate(
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
	}

	log.Printf("TiDB initialized: %s@%s:%s/%s", cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase)

	return db, nil
}
