package domain

import "errors"

// ビジネスロジックエラー
var (
	// ErrNotGuildMember はユーザーがギルドメンバーでない場合のエラー
	ErrNotGuildMember = errors.New("user is not a member of the guild")

	// ErrSessionNotFound はセッションが見つからない場合のエラー
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionExpired はセッションが期限切れの場合のエラー
	ErrSessionExpired = errors.New("session expired")

	// ErrUserNotFound はユーザーが見つからない場合のエラー
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidAuthCode は認可コードが無効な場合のエラー
	ErrInvalidAuthCode = errors.New("invalid authorization code")

	// ErrAuthCodeExpired は認可コードが期限切れの場合のエラー
	ErrAuthCodeExpired = errors.New("authorization code expired")

	// ErrAuthCodeAlreadyUsed は認可コードが既に使用済みの場合のエラー
	ErrAuthCodeAlreadyUsed = errors.New("authorization code already used")

	// ErrProfileNotFound はプロフィールが見つからない場合のエラー
	ErrProfileNotFound = errors.New("profile not found")
)
