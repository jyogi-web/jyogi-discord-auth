# Spec Kit タスク進捗自動更新フック

実装完了時に自動的にspecs/配下のtasks.mdを更新するhookシステムです。

## 概要

Claudeが実装タスクを完了して停止する際、自動的に：
1. specs/配下のtasks.mdを検索
2. 完了したタスクを検出
3. タスクのステータスを`completed`に更新

## 2つの実装方法

### 方法1: Prompt-Based Hook（推奨）

LLMが文脈を理解してタスク進捗を判断・更新します。

**ファイル:** `hooks.json`

**特徴:**
- 文脈を理解した柔軟な判断
- タスクの関連性を自動判定
- より自然な更新

**使い方:**
```bash
# hooks.jsonを使用（既に配置済み）
# Claude Code再起動で有効化
```

### 方法2: Command Hook（確実性重視）

シェルスクリプトで確実にチェック・更新します。

**ファイル:** `hooks-command-version.json` + `update-tasks.sh`

**特徴:**
- 高速・確実な動作
- 未完了タスクがある場合は必ず通知
- デバッグしやすい

**使い方:**
```bash
# hooks-command-version.jsonをhooks.jsonにコピー
cp .claude/hooks/hooks-command-version.json .claude/hooks/hooks.json

# Claude Code再起動で有効化
```

## 有効化手順

1. **hookファイルの選択**
   ```bash
   # Prompt版を使う場合（デフォルト、既に配置済み）
   # 何もしなくてOK

   # Command版を使う場合
   cp .claude/hooks/hooks-command-version.json .claude/hooks/hooks.json
   ```

2. **Claude Codeを再起動**
   ```bash
   # 現在のセッションを終了
   exit

   # 再起動
   claude
   ```

3. **動作確認**
   ```bash
   # デバッグモードで起動して確認
   claude --debug
   ```

## 動作フロー

### Prompt-Based Hook

```
実装完了 → Stop Hook発火
  ↓
文脈分析：
  - 今回のセッションで何が実装されたか？
  - どのタスクが完了したか？
  - tasks.mdに未完了タスクがあるか？
  ↓
必要に応じてtasks.mdを更新
  ↓
停止を承認
```

### Command Hook

```
実装完了 → Stop Hook発火
  ↓
update-tasks.sh実行：
  - specs/*/tasks.mdを検索
  - 未完了タスク（pending/in_progress）をチェック
  ↓
未完了タスクがある場合：
  - "block"決定を返す
  - LLMにタスク更新を依頼
  ↓
LLMがタスクを更新
  ↓
停止を承認
```

## テスト方法

### 1. テストタスクを作成

```markdown
# specs/001-jyogi-member-auth/tasks.md

## User Story 1: テスト

### Tasks
- [pending] T001: テスト実装
  - [ ] テストファイル作成
```

### 2. 簡単な実装を行う

```bash
# 何かファイルを編集
echo "// test" > test.go
```

### 3. 停止を試みる

Claudeに「終了」と伝えて、hookが動作するか確認します。

**期待される動作:**
- Prompt版: LLMがタスク完了を検出してtasks.mdを更新
- Command版: 未完了タスクがある旨の通知、LLMが更新を促される

## カスタマイズ

### タイムアウト調整

```json
{
  "type": "prompt",
  "prompt": "...",
  "timeout": 60  // 60秒に延長
}
```

### 対象ファイルの拡張

update-tasks.shを編集：

```bash
# 特定のディレクトリのみ対象
tasks_files=$(find specs/active -name "tasks.md" 2>/dev/null || true)

# 複数パターン対応
tasks_files=$(find specs -name "tasks*.md" -o -name "todo.md" 2>/dev/null || true)
```

### 更新ロジックの変更

hooks.jsonのpromptを編集：

```json
"prompt": "実装タスク完了時、以下のルールでtasks.mdを更新：\n1. 小タスク（[ ]）が全て完了したら、親タスクを[completed]に\n2. 更新時にコミットハッシュを記録\n3. 完了日時を追加\n..."
```

## トラブルシューティング

### Hookが動作しない

```bash
# デバッグモードで確認
claude --debug

# hookが読み込まれているか確認
/hooks
```

### Hookが正しく読み込まれない

```bash
# JSON構文チェック
jq . .claude/hooks/hooks.json

# スクリプトの実行権限確認
ls -la .claude/hooks/update-tasks.sh
```

### 手動テスト

```bash
# Command版のスクリプトを手動実行
echo '{"cwd": "/Users/uozumikouhei/workspace/jyogi-discord-auth"}' | \
  bash .claude/hooks/update-tasks.sh

echo "Exit code: $?"
```

## 無効化方法

### 一時的に無効化

```bash
# hooks.jsonをリネーム
mv .claude/hooks/hooks.json .claude/hooks/hooks.json.disabled

# Claude Code再起動
```

### 完全に削除

```bash
rm -rf .claude/hooks/
```

## ベストプラクティス

1. **最初はPrompt版を試す** - 柔軟で文脈理解が優れている
2. **確実性が必要ならCommand版** - デバッグしやすく、動作が予測可能
3. **デバッグモードで開発** - `claude --debug`で詳細ログを確認
4. **段階的に導入** - まずStopフックのみ、後でPostToolUseフックを追加
5. **タイムアウトは余裕を持たせる** - 特にPrompt版は処理時間が必要

## 今後の拡張アイデア

- [ ] 自動コミット: タスク完了時に自動でgit commit
- [ ] GitHub Issue連携: 完了タスクを自動でIssueクローズ
- [ ] 進捗レポート: 日次で進捗サマリーを生成
- [ ] タスク依存関係チェック: 前提タスクが完了しているか確認
- [ ] タスク推定時間記録: 実際の所要時間をログ

## 参考リンク

- [Claude Code Hooks公式ドキュメント](https://docs.claude.com/en/docs/claude-code/hooks)
- [Spec Kit](../../.specify/)
