---
# Project: jyogi-discord-auth
# Spec Kit Integration - 仕様駆動開発
---

# プロジェクト指示

このプロジェクトでは、**Spec Kit** を使用した仕様駆動開発を採用しています。

## Spec Kit とは

機能開発を以下の段階で進める体系的なアプローチです：

```
仕様定義 → 技術計画 → タスク分解 → 実装 → レビュー
```

各段階で明確な成果物（spec.md, plan.md, tasks.md）を生成し、品質を保ちながら開発を進めます。

## 利用可能なツール

### 1. スラッシュコマンド（手動実行）

個別のステップを直接実行：

- `/speckit.specify "機能説明"` - 仕様定義
- `/speckit.plan` - 技術計画策定
- `/speckit.tasks` - タスク分解
- `/speckit.implement` - 実装
- `/speckit.analyze` - 品質チェック
- `/speckit.clarify` - 仕様の曖昧さ解消
- `/speckit.constitution` - プロジェクト憲章作成
- `/speckit.taskstoissues` - タスクをGitHub Issueに変換
- `/speckit.checklist` - カスタムチェックリスト生成

### 2. スキル（ガイド付き実行）

ワークフロー全体をガイド：

```
# Skillツールを使用
skill: "speckit"
```

スキルを起動すると、各ステップの説明とともに段階的にワークフローを進められます。

### 3. エージェント（自動化実行）

専門エージェントが自動でワークフローを実行：

#### Workflow Coordinator
仕様から計画までを自動実行：

```
# ユーザーが以下のようなリクエストをすると自動起動：
"機能を開発して"
"spec kitで実装"
"仕様駆動で開発"
```

または明示的に：
```
Task: workflow-coordinator
Prompt: "ユーザー認証機能を追加"
```

#### Dev Agent
tasks.mdに基づいて実装を自動実行：

```
# ユーザーが以下のようなリクエストをすると自動起動：
"タスクを実装して"
"speckit implement"
```

または明示的に：
```
Task: dev-agent
Prompt: "specs/1-user-auth/tasks.md を実装"
```

## ワークフローの選択

### ケース1: 完全自動化が欲しい

```
User: "ユーザープロフィール編集機能を追加したい"

→ Workflow Coordinatorが自動起動
→ 仕様・計画・タスクを自動生成
→ Dev Agentに引き継ぎ（オプション）
```

### ケース2: 段階的に進めたい（推奨）

```
User: "ユーザープロフィール編集機能を追加したい"

Step 1: /speckit.specify "ユーザープロフィール編集機能"
→ 仕様を確認・調整

Step 2: /speckit.plan
→ 技術計画を確認・調整

Step 3: /speckit.tasks
→ タスクリストを確認

Step 4: /speckit.implement
→ 実装開始
```

### ケース3: スキルでガイドが欲しい

```
# Skillツールを使用
skill: "speckit"

→ 各ステップの説明付きでガイド
→ 好きなタイミングでスラッシュコマンド実行
```

## ディレクトリ構造

```
specs/
  N-feature-name/          # 各機能ごとのディレクトリ
    spec.md                # 機能仕様（WHAT & WHY）
    plan.md                # 技術計画（HOW）
    tasks.md               # 実装タスク
    data-model.md          # データモデル
    research.md            # 技術調査
    quickstart.md          # テストシナリオ
    contracts/             # API契約
    checklists/            # 品質チェックリスト
      requirements.md
```

## 開発原則

### 1. 既存コードパターンの厳守

新しい機能を追加する際は、**既存のコードベースのパターンを完全に踏襲**してください：

- ファイル構造
- 命名規則
- インポート順序
- エラーハンドリング
- ロギング
- コメントスタイル

### 2. 仕様駆動

コードを書く前に、必ず仕様（spec.md）と計画（plan.md）を確認してください。

### 3. 段階的実装

大きな機能は、ユーザーストーリーごとに分割して実装します。各ストーリーは独立してテスト可能であるべきです。

### 4. 品質優先

- すべての要件が仕様に記載されている
- すべてのタスクが明確で実行可能
- すべてのコードが既存パターンに従っている

## コミット規約

```
[フェーズ] 簡潔な説明

- タスクID: T001, T002
- ユーザーストーリー: US1
```

例：
```
[US1] ユーザー認証の基盤実装

UserモデルとAuthServiceを追加

- T012, T014, T015
- US1: User can sign up with email
```

## FAQ

**Q: どの方法を使うべき？**
A: 初めての機能なら「段階的」（ケース2）がおすすめ。各ステップで確認できます。

**Q: 既存の機能を修正したい場合は？**
A: Spec Kitは新機能向けです。バグ修正や小さな変更は通常の開発フローで。

**Q: spec.mdやplan.mdを手動で編集してもいい？**
A: はい！生成されたファイルは、必要に応じて手動で調整してください。

**Q: エージェントが間違った判断をしたら？**
A: いつでも中断して、手動でステップを進められます。

---

## クイックスタート

新機能を追加する場合：

```bash
# 1. 仕様を作成
/speckit.specify "機能の説明"

# 2. 生成されたspec.mdを確認

# 3. 技術計画を作成
/speckit.plan

# 4. 生成されたplan.mdを確認

# 5. タスクリストを作成
/speckit.tasks

# 6. 実装開始
/speckit.implement
```

準備完了！
