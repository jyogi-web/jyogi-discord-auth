package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
)

type authCodeRepository struct {
	db *sql.DB
}

// NewAuthCodeRepository は新しいSQLite認可コードリポジトリを作成します
func NewAuthCodeRepository(db *sql.DB) repository.AuthCodeRepository {
	return &authCodeRepository{db: db}
}

// Create は新しい認可コードをデータベースに挿入します
func (r *authCodeRepository) Create(ctx context.Context, authCode *domain.AuthCode) error {
	if err := authCode.Validate(); err != nil {
		return fmt.Errorf("invalid auth code: %w", err)
	}

	query := `
		INSERT INTO auth_codes (id, code, client_id, user_id, redirect_uri, expires_at, created_at, used)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		authCode.ID,
		authCode.Code,
		authCode.ClientID,
		authCode.UserID,
		authCode.RedirectURI,
		authCode.ExpiresAt.Format(time.RFC3339),
		authCode.CreatedAt.Format(time.RFC3339),
		authCode.Used,
	)
	if err != nil {
		return fmt.Errorf("failed to create auth code: %w", err)
	}

	return nil
}

// GetByCode は認可コードで認可コードを取得します
func (r *authCodeRepository) GetByCode(ctx context.Context, code string) (*domain.AuthCode, error) {
	query := `
		SELECT id, code, client_id, user_id, redirect_uri, expires_at, created_at, used
		FROM auth_codes
		WHERE code = ?
	`

	var expiresAtStr, createdAtStr string
	authCode := &domain.AuthCode{}

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&authCode.ID,
		&authCode.Code,
		&authCode.ClientID,
		&authCode.UserID,
		&authCode.RedirectURI,
		&expiresAtStr,
		&createdAtStr,
		&authCode.Used,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("auth code not found: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get auth code: %w", err)
	}

	// 日付文字列をパース
	authCode.ExpiresAt, err = time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expires_at: %w", err)
	}
	authCode.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	return authCode, nil
}

// MarkAsUsed は認可コードを使用済みにマークします
func (r *authCodeRepository) MarkAsUsed(ctx context.Context, code string) error {
	query := `UPDATE auth_codes SET used = 1 WHERE code = ?`

	result, err := r.db.ExecContext(ctx, query, code)
	if err != nil {
		return fmt.Errorf("failed to mark auth code as used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("auth code not found: %s", code)
	}

	return nil
}

// DeleteExpired は期限切れの認可コードをすべて削除します
func (r *authCodeRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM auth_codes WHERE expires_at < ?`

	result, err := r.db.ExecContext(ctx, query, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to delete expired auth codes: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// 削除された認可コード数をログ出力（オプション）
	if rowsAffected > 0 {
		fmt.Printf("Deleted %d expired auth codes\n", rowsAffected)
	}

	return nil
}

// boolToInt はbool値をSQLite用のint（0または1）に変換します
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// intToBool はSQLiteのint（0または1）をbool値に変換します
func intToBool(i int) bool {
	return i != 0
}
