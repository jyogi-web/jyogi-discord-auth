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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=true",
		cfg.TiDBUser,
		cfg.TiDBPassword,
		cfg.TiDBHost,
		cfg.TiDBPort,
		cfg.TiDBDatabase,
	)

	// カスタムロガー設定: RecordNotFound エラーを無視し、エラーのみ出力する
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,  // Slow SQL threshold
			LogLevel:                  logger.Error, // Log level
			IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,        // Don't include params in the SQL log (set to true to hide params)
			Colorful:                  false,        // Disable color
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to TiDB: %w", err)
	}

	// コネクションプール設定（sql.DBを取得して設定）
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// マイグレーションはここで行わず、呼び出し元（特にserver）で明示的に行うか、
	// あるいはここでモデルへの依存を持たせるか検討が必要。
	// 循環参照を避けるため、AutoMigrateは呼び出し元で行うのが無難だが、
	// repositories/gormパッケージ内なので、ここで行っても循環参照にはならない（モデルはこのパッケージ内）。
	// 簡便のため、ここでAutoMigrateを行うようにする。

	if err := db.AutoMigrate(
		&User{},
		&Session{},
		&ClientApp{},
		&AuthCode{},
		&Token{},
		&Profile{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	log.Printf("TiDB initialized: %s@%s:%s/%s", cfg.TiDBUser, cfg.TiDBHost, cfg.TiDBPort, cfg.TiDBDatabase)

	return db, nil
}
