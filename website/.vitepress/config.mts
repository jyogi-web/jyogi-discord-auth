import { defineConfig } from 'vitepress'

export default defineConfig({
    base: '/jyogi-discord-auth/',
    title: "Jyogi Auth",
    description: "Discord Authentication System for Jyogi",

    locales: {
        root: {
            label: '日本語',
            lang: 'ja-JP',
            title: 'Jyogi Auth',
            description: 'じょぎメンバー認証システム',
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
                returnToTopLabel: 'トップへ戻る'
            }
        },
        en: {
            label: 'English',
            lang: 'en-US',
            link: '/en/',
            title: 'Jyogi Auth',
            description: 'Discord Authentication System for Jyogi',
            themeConfig: {
                nav: [
                    { text: 'Home', link: '/en/' },
                    { text: 'Quick Start', link: '/en/guide/client-integration' },
                    { text: 'Contributor Guide', link: '/en/guide/contributing' },
                    { text: 'Reference', link: '/en/reference/api' }
                ],
                sidebar: {
                    '/en/guide/': [
                        {
                            text: 'For Client Integrators',
                            items: [
                                { text: 'Quick Start (Client Integration)', link: '/en/guide/client-integration' },
                                { text: 'Rails: Direct DB Access', link: '/en/guide/rails-direct-db' },
                                { text: 'API Reference', link: '/en/reference/api' }
                            ]
                        },
                        {
                            text: 'For Contributors',
                            items: [
                                { text: 'Setup Guide', link: '/en/guide/contributing' },
                                { text: 'Architecture', link: '/en/guide/architecture' },
                                { text: 'Deployment', link: '/en/guide/deployment' },
                                { text: 'Profile Sync', link: '/en/guide/profile-sync' },
                                { text: 'Testing Guide', link: '/en/guide/testing' },
                                { text: 'Troubleshooting', link: '/en/guide/troubleshooting' }
                            ]
                        }
                    ],
                    '/en/reference/': [
                        {
                            text: 'Reference',
                            items: [
                                { text: 'API Reference', link: '/en/reference/api' },
                                { text: 'Database Design', link: '/reference/database' },
                                { text: 'Environment Variables', link: '/en/reference/environment' }
                            ]
                        }
                    ]
                }
            }
        }
    },

    themeConfig: {
        socialLinks: [
            { icon: 'github', link: 'https://github.com/jyogi-web/jyogi-discord-auth' }
        ]
    }
})
