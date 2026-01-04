# アーキテクチャ

じょぎメンバー認証システムのアーキテクチャとプロジェクト構造について説明します。

## 概要

1. **Identity Provider (IdP)**: Discord（ユーザー情報、所属サーバーの管理）
2. **Auth Server（じょぎ認証）**: Discord OAuth2を実行し、ユーザーが「じょぎメンバーであるか」を判定。独自のアクセストークン（JWT）を発行する認証基盤。
3. **Client Apps（各内製ツール）**: 「じょぎ認証」にログインを委ねるアプリケーション群。

## 技術スタック

- **言語**: Go 1.23+
- **データベース**: SQLite（将来PostgreSQLへの移行可能な設計）
- **認証**: Discord OAuth2, JWT
- **デプロイ**: Google Cloud Run（認証サーバー）、Google Cloud Functions（プロフィール同期）

## プロジェクト構造

```
jyogi-discord-auth/
├── cmd/
│   ├── server/          # メインサーバーエントリーポイント
│   ├── sync-profiles/   # プロフィール同期ツール（CLI）
│   └── sync-profiles-fn/ # プロフィール同期Function（HTTP）
├── deployments/
│   ├── cloud-functions/ # Google Cloud Functions設定
│   └── aws-lambda/      # AWS Lambda設定
├── internal/
│   ├── domain/          # ドメインモデル（User, Profile, Session, etc.）
│   ├── repository/      # データアクセス層
│   ├── service/         # ビジネスロジック
│   ├── handler/         # HTTPハンドラー
│   ├── middleware/      # HTTPミドルウェア
│   └── config/          # 設定管理
├── pkg/
│   ├── discord/         # Discord APIクライアント、プロフィールパーサー
│   ├── auth/            # クライアント認証
│   └── jwt/             # JWTユーティリティ
├── web/
│   ├── templates/       # HTMLテンプレート
│   └── static/          # 静的ファイル
├── migrations/          # データベースマイグレーション
├── tests/               # テスト
└── scripts/             # 開発・運用スクリプト
```
