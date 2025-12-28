package discord

import (
	"testing"
)

func TestParseProfile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected ProfileData
	}{
		{
			name: "完全なプロフィール - ⭕付き",
			content: `⭕本名:じょぎ太郎
⭕学籍番号:20X1234
⭕趣味:カラオケ、ゲーム、アニメ鑑賞
⭕じょぎでやりたいこと:ゲーム作成
⭕ひとこと:よろしくお願いします！`,
			expected: ProfileData{
				RealName:  "じょぎ太郎",
				StudentID: "20X1234",
				Hobbies:   "カラオケ、ゲーム、アニメ鑑賞",
				WhatToDo:  "ゲーム作成",
				Comment:   "よろしくお願いします！",
			},
		},
		{
			name: "完全なプロフィール - ○付き",
			content: `○本名:大江由眞
○学籍番号:ないです
○趣味:ゲーム、カラオケ
○じょぎでやりたいこと:パックカソン
○ひとこと:中洲の専門学校から来ました！`,
			expected: ProfileData{
				RealName:  "大江由眞",
				StudentID: "ないです",
				Hobbies:   "ゲーム、カラオケ",
				WhatToDo:  "パックカソン",
				Comment:   "中洲の専門学校から来ました！",
			},
		},
		{
			name: "全角コロンを使用",
			content: `⭕本名：じょぎ太郎
⭕学籍番号：20X1234
⭕趣味：カラオケ、ゲーム
⭕じょぎでやりたいこと：ゲーム作成
⭕ひとこと：よろしくお願いします！`,
			expected: ProfileData{
				RealName:  "じょぎ太郎",
				StudentID: "20X1234",
				Hobbies:   "カラオケ、ゲーム",
				WhatToDo:  "ゲーム作成",
				Comment:   "よろしくお願いします！",
			},
		},
		{
			name: "記号なし",
			content: `本名:じょぎ太郎
学籍番号:20X1234
趣味:カラオケ、ゲーム
じょぎでやりたいこと:ゲーム作成
ひとこと:よろしくお願いします！`,
			expected: ProfileData{
				RealName:  "じょぎ太郎",
				StudentID: "20X1234",
				Hobbies:   "カラオケ、ゲーム",
				WhatToDo:  "ゲーム作成",
				Comment:   "よろしくお願いします！",
			},
		},
		{
			name: "スペース付き",
			content: `⭕ 本名 : じょぎ太郎
⭕ 学籍番号 : 20X1234
⭕ 趣味 : カラオケ、ゲーム
⭕ じょぎでやりたいこと : ゲーム作成
⭕ ひとこと : よろしくお願いします！`,
			expected: ProfileData{
				RealName:  "じょぎ太郎",
				StudentID: "20X1234",
				Hobbies:   "カラオケ、ゲーム",
				WhatToDo:  "ゲーム作成",
				Comment:   "よろしくお願いします！",
			},
		},
		{
			name: "一部フィールドのみ",
			content: `⭕本名:じょぎ太郎
⭕趣味:カラオケ、ゲーム
⭕ひとこと:よろしくお願いします！`,
			expected: ProfileData{
				RealName:  "じょぎ太郎",
				StudentID: "",
				Hobbies:   "カラオケ、ゲーム",
				WhatToDo:  "",
				Comment:   "よろしくお願いします！",
			},
		},
		{
			name: "箇条書き形式",
			content: `• 本名：奥村　直（なお）
• 学籍番号：ひみつ
• 趣味：ギター
• じょぎでやりたいこと：技術系の勉強（組込とかMLやりたいです！）
• ひとこと：いまは飯塚に住んでいます！よろしくお願いします！`,
			expected: ProfileData{
				RealName:  "奥村　直（なお）",
				StudentID: "ひみつ",
				Hobbies:   "ギター",
				WhatToDo:  "技術系の勉強（組込とかMLやりたいです！）",
				Comment:   "いまは飯塚に住んでいます！よろしくお願いします！",
			},
		},
		{
			name:    "無効なプロフィール",
			content: "これは自己紹介ではありません",
			expected: ProfileData{
				RealName:  "",
				StudentID: "",
				Hobbies:   "",
				WhatToDo:  "",
				Comment:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseProfile(tt.content)

			if result.RealName != tt.expected.RealName {
				t.Errorf("RealName = %v, want %v", result.RealName, tt.expected.RealName)
			}
			if result.StudentID != tt.expected.StudentID {
				t.Errorf("StudentID = %v, want %v", result.StudentID, tt.expected.StudentID)
			}
			if result.Hobbies != tt.expected.Hobbies {
				t.Errorf("Hobbies = %v, want %v", result.Hobbies, tt.expected.Hobbies)
			}
			if result.WhatToDo != tt.expected.WhatToDo {
				t.Errorf("WhatToDo = %v, want %v", result.WhatToDo, tt.expected.WhatToDo)
			}
			if result.Comment != tt.expected.Comment {
				t.Errorf("Comment = %v, want %v", result.Comment, tt.expected.Comment)
			}
		})
	}
}

func TestIsValidProfile(t *testing.T) {
	tests := []struct {
		name     string
		profile  ProfileData
		expected bool
	}{
		{
			name: "全フィールド入力",
			profile: ProfileData{
				RealName:  "じょぎ太郎",
				StudentID: "20X1234",
				Hobbies:   "カラオケ",
				WhatToDo:  "ゲーム作成",
				Comment:   "よろしく",
			},
			expected: true,
		},
		{
			name: "1フィールドのみ",
			profile: ProfileData{
				RealName: "じょぎ太郎",
			},
			expected: true,
		},
		{
			name:     "全フィールド空",
			profile:  ProfileData{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.IsValidProfile()
			if result != tt.expected {
				t.Errorf("IsValidProfile() = %v, want %v", result, tt.expected)
			}
		})
	}
}
