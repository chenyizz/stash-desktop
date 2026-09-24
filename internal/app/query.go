package app

import (
	"strings"
	"unicode/utf8"
)

// maxQueryRunes 限制搜索串长度（按 rune，避免 CJK 被字节截断）。
const maxQueryRunes = 256

// ScenesFilter 是场景列表的过滤条件（nil 表示不过滤）。
type ScenesFilter struct {
	Organized *bool  `json:"organized"` // nil=全部
	RatingMin *int   `json:"ratingMin"` // >= 阈值（0-100）
	DateFrom  string `json:"dateFrom"`  // YYYY-MM-DD
	DateTo    string `json:"dateTo"`

	TagIDs         []int `json:"tagIds"`
	TagAll         bool  `json:"tagAll"`         // true=全部命中，默认任一
	IncludeSubTags bool  `json:"includeSubTags"` // 含子孙标签

	PerformerIDs []int `json:"performerIds"`
	PerformerAll bool  `json:"performerAll"` // 默认 true=全部命中

	StudioID          *int `json:"studioId"`
	IncludeSubStudios bool `json:"includeSubStudios"`
}

// ScenesQuery 是场景列表的查询入参（过滤字段将在后续子步骤中追加，方法签名不变）。
type ScenesQuery struct {
	Query    string        `json:"query"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
	Filter   *ScenesFilter `json:"filter,omitempty"`
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
