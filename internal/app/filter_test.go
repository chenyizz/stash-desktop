package app

import (
	"testing"

	"case/backend/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func boolPtr(v bool) *bool { return &v }

func TestBuildSceneFilter_NilAndEmpty(t *testing.T) {
	f, err := buildSceneFilter(nil)
	require.NoError(t, err)
	assert.Nil(t, f)

	f, err = buildSceneFilter(&ScenesFilter{})
	require.NoError(t, err)
	assert.Nil(t, f)
}

func TestBuildSceneFilter_Organized(t *testing.T) {
	f, err := buildSceneFilter(&ScenesFilter{Organized: boolPtr(true)})
	require.NoError(t, err)
	require.NotNil(t, f.Organized)
	assert.True(t, *f.Organized)
}

func TestBuildSceneFilter_Rating(t *testing.T) {
	min := 50
	f, err := buildSceneFilter(&ScenesFilter{RatingMin: &min})
	require.NoError(t, err)
	require.NotNil(t, f.Rating100)
	assert.Equal(t, 50, f.Rating100.Value)
	require.NotNil(t, f.Rating100.Value2)
	assert.Equal(t, 100, *f.Rating100.Value2)
	assert.Equal(t, models.CriterionModifierBetween, f.Rating100.Modifier)

	bad := 150
	_, err = buildSceneFilter(&ScenesFilter{RatingMin: &bad})
	require.Error(t, err)
}

func TestBuildSceneFilter_Date(t *testing.T) {
	f, err := buildSceneFilter(&ScenesFilter{DateFrom: "2020-01-01", DateTo: "2021-12-31"})
	require.NoError(t, err)
	require.NotNil(t, f.Date)
	assert.Equal(t, "2020-01-01", f.Date.Value)
	require.NotNil(t, f.Date.Value2)
	assert.Equal(t, "2021-12-31", *f.Date.Value2)

	// 仅起始：另一端补极值
	f, err = buildSceneFilter(&ScenesFilter{DateFrom: "2020-01-01"})
	require.NoError(t, err)
	require.NotNil(t, f.Date)
	assert.Equal(t, "9999-12-31", *f.Date.Value2)

	_, err = buildSceneFilter(&ScenesFilter{DateFrom: "not-a-date"})
	require.Error(t, err)
}

func TestBuildSceneFilter_Tags(t *testing.T) {
	// 默认任一，仅自身
	f, err := buildSceneFilter(&ScenesFilter{TagIDs: []int{1, 2}})
	require.NoError(t, err)
	require.NotNil(t, f.Tags)
	assert.Equal(t, []string{"1", "2"}, f.Tags.Value)
	assert.Equal(t, models.CriterionModifierIncludes, f.Tags.Modifier)
	assert.Nil(t, f.Tags.Depth)

	// 全部 + 含子孙
	f, err = buildSceneFilter(&ScenesFilter{TagIDs: []int{3}, TagAll: true, IncludeSubTags: true})
	require.NoError(t, err)
	require.NotNil(t, f.Tags)
	assert.Equal(t, models.CriterionModifierIncludesAll, f.Tags.Modifier)
	require.NotNil(t, f.Tags.Depth)
	assert.Equal(t, -1, *f.Tags.Depth)

	_, err = buildSceneFilter(&ScenesFilter{TagIDs: []int{0}})
	require.Error(t, err)
}

func TestBuildSceneFilter_PerformersAndStudio(t *testing.T) {
	f, err := buildSceneFilter(&ScenesFilter{
		PerformerIDs:      []int{7},
		PerformerAll:      true,
		StudioID:          intPtr(9),
		IncludeSubStudios: true,
	})
	require.NoError(t, err)
	require.NotNil(t, f.Performers)
	assert.Equal(t, models.CriterionModifierIncludesAll, f.Performers.Modifier)
	require.NotNil(t, f.Studios)
	assert.Equal(t, []string{"9"}, f.Studios.Value)
	require.NotNil(t, f.Studios.Depth)
	assert.Equal(t, -1, *f.Studios.Depth)

	_, err = buildSceneFilter(&ScenesFilter{StudioID: intPtr(0)})
	require.Error(t, err)
}
