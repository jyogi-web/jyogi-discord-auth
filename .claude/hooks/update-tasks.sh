#!/bin/bash
set -euo pipefail

# タスク進捗自動更新スクリプト
# Stopフックから呼び出され、完了したタスクをtasks.mdに記録する

# 入力JSON読み込み
input=$(cat)

# プロジェクトディレクトリ
project_dir=$(echo "$input" | jq -r '.cwd')
cd "$project_dir"

# specs/配下のtasks.mdを探す
tasks_files=$(find specs -name "tasks.md" 2>/dev/null || true)

if [ -z "$tasks_files" ]; then
  # tasks.mdが見つからない場合は何もしない
  echo '{"decision": "approve", "systemMessage": "タスクファイルが見つかりませんでした"}' >&2
  exit 0
fi

# 各tasks.mdをチェック（通常は1つだけ）
updated=false
for tasks_file in $tasks_files; do
  # tasks.mdに未完了タスク（[ ]）があるか確認
  # T0XX形式のタスクIDを持つ未完了タスクを検索
  incomplete_count=$(grep -cE '^- \[ \] T[0-9]+' "$tasks_file" 2>/dev/null || echo "0")

  if [ "$incomplete_count" -gt 0 ]; then
    # 未完了タスクが存在する場合、LLMに更新を依頼
    echo "{
      \"decision\": \"block\",
      \"reason\": \"タスク進捗を更新してください\",
      \"systemMessage\": \"$tasks_file に${incomplete_count}件の未完了タスクがあります。今回のセッションで完了したタスクがあれば、Editツールを使用して '- [ ]' を '- [X]' に更新してください。完了したタスクがない場合は、このメッセージを無視して停止してください。\"
    }" >&2
    exit 2
  fi
done

# すべてのタスクが完了している、または更新が必要ない場合
echo '{"decision": "approve", "systemMessage": "すべてのタスクが完了しています（または更新の必要なし）"}' >&2
exit 0
