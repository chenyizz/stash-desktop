package app

import (
	"math"
	"testing"

	"case/backend/pkg/models"

	"github.com/stretchr/testify/require"
)

func intPtr(v int) *int { return &v }

func mustDate(t *testing.T, s string) *models.Date {
	t.Helper()

	d, err := models.ParseDate(s)
	require.NoError(t, err)

	return &d
}

func fullScene(t *testing.T) *models.Scene {
	t.Helper()

	return &models.Scene{
		ID:             1,
		Title:          "FDD-2002",
		Code:           "FDD-2002",
		Details:        "details",
		Director:       "director",
		Date:           mustDate(t, "2002-06-28"),
		ProductionDate: mustDate(t, "2002-06-01"),
		Rating:         intPtr(85),
		Organized:      true,
		StudioID:       intPtr(7),
		URLs:           models.NewRelatedStrings([]string{"https://www.javbus.com/FDD2002"}),
		TagIDs:         models.NewRelatedIDs([]int{1, 2}),
		PerformerIDs:   models.NewRelatedIDs([]int{11}),
		Files: models.NewRelatedVideoFiles([]*models.VideoFile{
			{
				BaseFile:   &models.BaseFile{ID: models.FileID(10), Path: "dir/a.mp4", Basename: "a.mp4"},
				Duration:   120,
				Width:      1920,
				Height:     1080,
				VideoCodec: "h264",
				AudioCodec: "aac",
				FrameRate:  29.97,
				BitRate:    5000000,
				Format:     "mp4",
			},
			{
				BaseFile: &models.BaseFile{ID: models.FileID(20), Path: "dir/b.mp4", Basename: "b.mp4"},
				Duration: 60,
				Width:    1280,
				Height:   720,
			},
		}),
	}
}

func TestToSceneDetailDTO_Full(t *testing.T) {
	s := fullScene(t)
	// Deliberately pass tags in a different order to prove pairing follows TagIDs.
	tags := []*models.Tag{
		{ID: 2, Name: "无码"},
		{ID: 1, Name: "FDD"},
	}
	performers := []*models.Performer{{ID: 11, Name: "夕樹洋子"}}
	studio := &models.Studio{ID: 7, Name: "ファンタドリーム"}

	dto := toSceneDetailDTO(s, tags, performers, studio)

	require.Equal(t, 1, dto.ID)
	require.Equal(t, "FDD-2002", dto.Title)
	require.Equal(t, "FDD-2002", dto.Code)
	require.Equal(t, "details", dto.Details)
	require.Equal(t, "director", dto.Director)
	require.Equal(t, "2002-06-28", dto.Date)
	require.Equal(t, "2002-06-01", dto.ProductionDate)
	require.Equal(t, 85, dto.Rating)
	require.True(t, dto.Organized)
	require.NotNil(t, dto.StudioID)
	require.Equal(t, 7, *dto.StudioID)
	require.Equal(t, "ファンタドリーム", dto.StudioName)
	require.Equal(t, []string{"https://www.javbus.com/FDD2002"}, dto.URLs)

	require.Equal(t, []TagDTO{{ID: 1, Name: "FDD"}, {ID: 2, Name: "无码"}}, dto.Tags)
	require.Equal(t, []PerformerDTO{{ID: 11, Name: "夕樹洋子"}}, dto.Performers)

	// Primary file convenience fields (first file).
	require.Equal(t, "dir/a.mp4", dto.Path)
	require.Equal(t, float64(120), dto.Duration)
	require.Equal(t, 1920, dto.Width)
	require.Equal(t, 1080, dto.Height)
	require.Equal(t, "1080p", dto.Resolution)
	require.Equal(t, "h264", dto.VideoCodec)
	require.Equal(t, "aac", dto.AudioCodec)
	require.Equal(t, 29.97, dto.FrameRate)
	require.Equal(t, int64(5000000), dto.BitRate)

	require.Len(t, dto.Files, 2)
	require.Equal(t, 10, dto.Files[0].ID)
	require.Equal(t, "a.mp4", dto.Files[0].Basename)
	require.Equal(t, 20, dto.Files[1].ID)
}

func TestToSceneDetailDTO_EmptyRelationsAndNilFields(t *testing.T) {
	s := &models.Scene{
		ID:           2,
		Path:         "fallback/path.mp4",
		URLs:         models.NewRelatedStrings([]string{}),
		TagIDs:       models.NewRelatedIDs([]int{}),
		PerformerIDs: models.NewRelatedIDs([]int{}),
		Files:        models.NewRelatedVideoFiles([]*models.VideoFile{}),
	}

	dto := toSceneDetailDTO(s, nil, nil, nil)

	require.Equal(t, 0, dto.Rating)
	require.Equal(t, "", dto.Date)
	require.Equal(t, "", dto.ProductionDate)
	require.Nil(t, dto.StudioID)
	require.Equal(t, "", dto.StudioName)
	require.Nil(t, dto.Tags)
	require.Nil(t, dto.Performers)
	require.Nil(t, dto.Files)
	require.Equal(t, "fallback/path.mp4", dto.Path)
	require.Equal(t, "", dto.Resolution)
}

func TestToSceneDetailDTO_NonFiniteNumbers(t *testing.T) {
	s := &models.Scene{
		ID:           3,
		URLs:         models.NewRelatedStrings([]string{}),
		TagIDs:       models.NewRelatedIDs([]int{}),
		PerformerIDs: models.NewRelatedIDs([]int{}),
		Files: models.NewRelatedVideoFiles([]*models.VideoFile{
			{
				BaseFile:  &models.BaseFile{ID: models.FileID(30), Path: "c.mp4"},
				Duration:  math.Inf(1),
				FrameRate: math.NaN(),
			},
		}),
	}

	dto := toSceneDetailDTO(s, nil, nil, nil)

	require.Equal(t, float64(0), dto.Duration)
	require.Equal(t, float64(0), dto.FrameRate)
	require.Equal(t, float64(0), dto.Files[0].Duration)
	require.Equal(t, float64(0), dto.Files[0].FrameRate)
}

func TestToSceneDetailDTO_PortraitResolution(t *testing.T) {
	s := &models.Scene{
		ID:           4,
		URLs:         models.NewRelatedStrings([]string{}),
		TagIDs:       models.NewRelatedIDs([]int{}),
		PerformerIDs: models.NewRelatedIDs([]int{}),
		Files: models.NewRelatedVideoFiles([]*models.VideoFile{
			{BaseFile: &models.BaseFile{ID: models.FileID(40), Path: "d.mp4"}, Width: 1080, Height: 1920},
		}),
	}

	dto := toSceneDetailDTO(s, nil, nil, nil)
	require.Equal(t, "1080p", dto.Resolution)
}

func TestResolutionLabel(t *testing.T) {
	require.Equal(t, "", resolutionLabel(0, 0))
	require.Equal(t, "", resolutionLabel(1920, 0))
	require.Equal(t, "1080p", resolutionLabel(1920, 1080))
	require.Equal(t, "1080p", resolutionLabel(1080, 1920))
}
