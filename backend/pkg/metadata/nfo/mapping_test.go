package nfo

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSceneMetadataSample(t *testing.T) {
	m, err := ParseFile(filepath.Join("testdata", "FDD-2002.nfo"))
	require.NoError(t, err)

	md := m.SceneMetadata()

	require.Equal(t, "FDD-2002 ビッグ·ティッツ·ビューティー CD1", md.Title)
	require.Equal(t, "FDD-2002", md.Code)
	require.Equal(t, "发行日期 2002-06-28", md.Details)
	require.Equal(t, "ファンタドリーム", md.StudioName)

	require.NotNil(t, md.Date)
	require.Equal(t, "2002-06-28", md.Date.String())
	require.NotNil(t, md.ProductionDate)
	require.Equal(t, "2002-06-28", md.ProductionDate.String())

	require.Equal(t, []string{"夕樹洋子", "松坂季美子", "加山なつこ"}, md.PerformerNames)
	require.Equal(t, []string{"FDD", "无码", "BigTitsBeauty", "ファンタドリーム"}, md.TagNames)
	require.Equal(t, []string{"夕樹洋子", "松坂季美子", "加山なつこ", "BigTitsBeauty"}, md.GroupNames)
	require.Equal(t, []string{"https://www.javbus.com/FDD2002"}, md.URLs)

	require.Equal(t, "JP-18+", md.CustomFields["mpaa"])
	require.Equal(t, "JP-18+", md.CustomFields["customrating"])
	require.Equal(t, "JP", md.CustomFields["countrycode"])
	require.Equal(t, "FDD-2002", md.CustomFields["javdbsearchid"])
	require.Equal(t, "108", md.CustomFields["runtime"])
}

func TestSceneMetadataDateFallback(t *testing.T) {
	m := &Movie{Year: "2002"}

	md := m.SceneMetadata()
	require.NotNil(t, md.Date)
	require.Equal(t, "2002", md.Date.String())
	require.Nil(t, md.ProductionDate)
}

func TestSceneMetadataRatingScale(t *testing.T) {
	tests := []struct {
		in       string
		expected int
	}{
		{in: "8.5", expected: 85},
		{in: "10", expected: 100},
		{in: "", expected: 0},
	}

	for _, tt := range tests {
		md := (&Movie{Rating: tt.in}).SceneMetadata()
		if tt.expected == 0 {
			require.Nil(t, md.Rating)
			continue
		}
		require.NotNil(t, md.Rating)
		require.Equal(t, tt.expected, *md.Rating)
	}
}

func TestSceneMetadataActorTypes(t *testing.T) {
	m := &Movie{
		Actors: []Actor{
			{Name: "A", Type: "Actor"},
			{Name: "D", Type: "Director"},
			{Name: "C", Type: "Creator"},
			{Name: "B"},
		},
	}

	md := m.SceneMetadata()
	require.Equal(t, []string{"A", "B"}, md.PerformerNames)
	require.Equal(t, "D", md.Director)
}

func TestSceneMetadataDedupAndNormalize(t *testing.T) {
	m := &Movie{
		Genres: []string{"B"},
		Tags:   []string{"A", " A ", "a"},
	}

	md := m.SceneMetadata()
	require.Equal(t, []string{"B", "A"}, md.TagNames)
}

func TestSceneMetadataFullWidthDedup(t *testing.T) {
	m := &Movie{
		Tags: []string{"ＡＢＣ", "ABC"},
	}

	md := m.SceneMetadata()
	require.Equal(t, []string{"ABC"}, md.TagNames)
}

func TestParseRatingAlreadyNormalized(t *testing.T) {
	r := parseRating("85")
	require.NotNil(t, r)
	require.Equal(t, 85, *r)
}
