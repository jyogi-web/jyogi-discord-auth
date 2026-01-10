import { defineConfig } from 'vitepress'

export default defineConfig({
    base: '/jyogi-discord-auth/',
    title: "Jyogi Auth",
    description: "じょぎメンバー認証システム",
    lang: 'ja-JP',

    themeConfig: {
        nav: [
            { text: 'ホーム', link: '/' },
            { text: 'クイックスタート', link: '/guide/client-integration' },
            { text: '開発者ガイド', link: '/guide/contributing' },
            { text: 'リファレンス', link: '/reference/api' }
        ],
        sidebar: {
            '/guide/': [
                {
                    text: '利用者向け (Client Integrators)',
                    items: [
                        { text: 'クイックスタート (クライアント統合)', link: '/guide/client-integration' },
                        { text: 'Rails: DB直接参照', link: '/guide/rails-direct-db' },
                        { text: 'API リファレンス', link: '/reference/api' }
                    ]
                },
                {
                    text: '開発者向け (Contributors)',
                    items: [
                        { text: '開発環境のセットアップ', link: '/guide/contributing' },
                        { text: 'アーキテクチャ', link: '/guide/architecture' },
                        { text: 'デプロイ', link: '/guide/deployment' },
                        { text: 'プロフィール同期', link: '/guide/profile-sync' },
                        { text: 'テストガイド', link: '/guide/testing' },
                        { text: 'トラブルシューティング', link: '/guide/troubleshooting' },
                        { text: 'データベース設計', link: '/reference/database' }
                    ]
                }
            ],
            '/reference/': [
                {
                    text: 'リファレンス',
                    items: [
                        { text: 'API リファレンス', link: '/reference/api' },
                        { text: 'データベース設計', link: '/reference/database' },
                        { text: '環境変数', link: '/reference/environment' }
                    ]
                }
            ]
        },
        docFooter: {
            prev: '前のページ',
            next: '次のページ'
        },
        outline: {
            label: 'このページの内容'
        },
        lastUpdated: {
            text: '最終更新',
            formatOptions: {
                dateStyle: 'short',
                timeStyle: 'short'
            }
        },
        darkModeSwitchLabel: '外観モード',
        sidebarMenuLabel: 'メニュー',
        returnToTopLabel: 'トップへ戻る',
        socialLinks: [
            { icon: 'github', link: 'https://github.com/jyogi-web/jyogi-discord-auth' }
        ]
    }
})