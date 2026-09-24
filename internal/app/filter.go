package app

import (
	"fmt"
	"strconv"

	"case/backend/pkg/models"
)

// buildSceneFilter 把前端过滤条件映射为 Repository 的 SceneFilterType。
// nil 或全空返回 nil（不过滤）；非法值返回 error。
func buildSceneFilter(f *ScenesFilter) (*models.SceneFilterType, error) {
	if f == nil {
		return nil, nil
	}

	filter := &models.SceneFilterType{}
	empty := true

	if f.Organized != nil {
		filter.Organized = f.Organized
		empty = false
	}

	if f.RatingMin != nil {
		min := *f.RatingMin
		if min < 0 || min > 100 {
			return nil, fmt.Errorf("ratingMin 必须在 0-100，得到 %d", min)
		}
		max := 100
		filter.Rating100 = &models.IntCriterionInput{
			Value:    min,
			Value2:   &max,
			Modifier: models.CriterionModifierBetween,
		}
		empty = false
	}

	if f.DateFrom != "" || f.DateTo != "" {
		from, to := f.DateFrom, f.DateTo
		if from == "" {
			from = "0001-01-01"
		}
		if to == "" {
			to = "9999-12-31"
		}
		if _, err := models.ParseDate(from); err != nil {
			return nil, fmt.Errorf("dateFrom 无效: %w", err)
		}
		if _, err := models.ParseDate(to); err != nil {
			return nil, fmt.Errorf("dateTo 无效: %w", err)
		}
		filter.Date = &models.DateCriterionInput{
			Value:    from,
			Value2:   &to,
			Modifier: models.CriterionModifierBetween,
		}
		empty = false
	}

	if len(f.TagIDs) > 0 {
		ids, err := intIDsToStrings(f.TagIDs)
		if err != nil {
			return nil, fmt.Errorf("tagIds: %w", err)
		}
		filter.Tags = &models.HierarchicalMultiCriterionInput{
			Value:    ids,
			Modifier: multiModifier(f.TagAll),
			Depth:    subDepth(f.IncludeSubTags),
		}
		empty = false
	}

	if len(f.PerformerIDs) > 0 {
		ids, err := intIDsToStrings(f.PerformerIDs)
		if err != nil {
			return nil, fmt.Errorf("performerIds: %w", err)
		}
		filter.Performers = &models.MultiCriterionInput{
			Value:    ids,
			Modifier: multiModifier(f.PerformerAll),
		}
		empty = false
	}

	if f.StudioID != nil {
		if *f.StudioID <= 0 {
			return nil, fmt.Errorf("studioId 无效: %d", *f.StudioID)
		}
		filter.Studios = &models.HierarchicalMultiCriterionInput{
			Value:    []string{strconv.Itoa(*f.StudioID)},
			Modifier: models.CriterionModifierIncludes,
			Depth:    subDepth(f.IncludeSubStudios),
		}
		empty = false
	}

	if empty {
		return nil, nil
	}

	return filter, nil
}

func intIDsToStrings(ids []int) ([]string, error) {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, fmt.Errorf("无效 ID: %d", id)
		}
		out = append(out, strconv.Itoa(id))
	}

	return out, nil
}

func multiModifier(all bool) models.CriterionModifier {
	if all {
		return models.CriterionModifierIncludesAll
	}
	return models.CriterionModifierIncludes
}

// subDepth：不含子孙时返回 nil（仅自身）；含子孙时 -1（无限层）。
func subDepth(include bool) *int {
	if !include {
		return nil
	}
	d := -1
	return &d
}
