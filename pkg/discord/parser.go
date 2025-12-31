package discord

import (
	"regexp"
	"strings"
)

// ProfileData は自己紹介から抽出したプロフィールデータを表します
type ProfileData struct {
	RealName  string
	StudentID string
	Hobbies   string
	WhatToDo  string
	Comment   string
}

// ParseProfile は自己紹介メッセージからプロフィール情報をパースします
// フォーマット例:
// ⭕本名:じょぎ太郎
// ⭕学籍番号:20X1234
// ⭕趣味:アニメ鑑賞
// ⭕じょぎでやりたいこと:ゲーム作成
// ⭕ひとこと:よろしくお願いします！
func ParseProfile(content string) *ProfileData {
	profile := &ProfileData{}

	// 各フィールドを抽出する正規表現パターン
	patterns := map[string]*regexp.Regexp{
		"real_name":  regexp.MustCompile(`(?:⭕|○|◯)?本名\s*[:：]\s*(.+?)(?:\n|$)`),
		"student_id": regexp.MustCompile(`(?:⭕|○|◯)?学籍番号\s*[:：]\s*(.+?)(?:\n|$)`),
		"hobbies":    regexp.MustCompile(`(?:⭕|○|◯)?趣味\s*[:：]\s*(.+?)(?:\n|$)`),
		"what_to_do": regexp.MustCompile(`(?:⭕|○|◯)?じょぎでやりたいこと\s*[:：]\s*(.+?)(?:\n|$)`),
		"comment":    regexp.MustCompile(`(?:⭕|○|◯)?ひとこと\s*[:：]\s*(.+?)(?:\n|$)`),
	}

	// 本名を抽出
	if match := patterns["real_name"].FindStringSubmatch(content); len(match) > 1 {
		profile.RealName = strings.TrimSpace(match[1])
	}

	// 学籍番号を抽出
	if match := patterns["student_id"].FindStringSubmatch(content); len(match) > 1 {
		profile.StudentID = strings.TrimSpace(match[1])
	}

	// 趣味を抽出
	if match := patterns["hobbies"].FindStringSubmatch(content); len(match) > 1 {
		profile.Hobbies = strings.TrimSpace(match[1])
	}

	// じょぎでやりたいことを抽出
	if match := patterns["what_to_do"].FindStringSubmatch(content); len(match) > 1 {
		profile.WhatToDo = strings.TrimSpace(match[1])
	}

	// ひとことを抽出
	if match := patterns["comment"].FindStringSubmatch(content); len(match) > 1 {
		profile.Comment = strings.TrimSpace(match[1])
	}

	return profile
}

// IsValidProfile はパースされたプロフィールが有効かどうかを確認します
func (p *ProfileData) IsValidProfile() bool {
	// 少なくとも1つのフィールドに値があれば有効とみなす
	return p.RealName != "" ||
		p.StudentID != "" ||
		p.Hobbies != "" ||
		p.WhatToDo != "" ||
		p.Comment != ""
}
