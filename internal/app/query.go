package app

import (
	"strings"
	"unicode/utf8"
)

// maxQueryRunes 限制搜索串长度（按 rune，避免 CJK 被字节截断）。
const maxQueryRunes = 256

// ScenesQuery 是场景列表的查询入参（过滤字段将在后续子步骤中追加，方法签名不变）。
type ScenesQuery struct {
	Query    string `json:"query"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

type PerformersQuery struct {
	Query    string `json:"query"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

type TagsQuery struct {
	Query    string `json:"query"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

type StudiosQuery struct {
	Query    string `json:"query"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

// normalizeQuery 规整搜索串：trim；空串返回 nil（等价未搜索）；超长按 rune 截断。
func normalizeQuery(q string) *string {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil
	}
	if utf8.RuneCountInString(q) > maxQueryRunes {
		q = string([]rune(q)[:maxQueryRunes])
	}

	return &q
}

// normalizeListQuery 统一规整列表查询参数。
func normalizeListQuery(query string, page, pageSize int) (*string, int, int) {
	page, pageSize = normalizePage(page, pageSize)

	return normalizeQuery(query), page, pageSize
}
