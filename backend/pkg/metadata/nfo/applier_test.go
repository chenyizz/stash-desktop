package nfo

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"case/backend/pkg/models"
	"case/backend/pkg/models/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const sampleNFO = "testdata/FDD-2002.nfo"

// writeSample creates a video file plus an NFO sidecar in a temp directory and
// returns the video path.
func writeSample(t *testing.T, nfoContents string) string {
	t.Helper()

	dir := t.TempDir()
	videoPath := filepath.Join(dir, "FDD-2002.mp4")
	require.NoError(t, os.WriteFile(videoPath, []byte("x"), 0o644))

	nfoPath := filepath.Join(dir, "FDD-2002.nfo")
	require.NoError(t, os.WriteFile(nfoPath, []byte(nfoContents), 0o644))

	return videoPath
}

func sampleContents(t *testing.T) string {
	t.Helper()

	data, err := os.ReadFile(sampleNFO)
	require.NoError(t, err)

	return string(data)
}

func TestApplier_NoNFOSidecar(t *testing.T) {
	dir := t.TempDir()
	videoPath := filepath.Join(dir, "alone.mp4")
	require.NoError(t, os.WriteFile(videoPath, []byte("x"), 0o644))

	db := mocks.NewDatabase()
	applier := &Applier{Repo: db.Repository()}

	require.NoError(t, applier.Apply(context.Background(), 1, videoPath))

	db.Scene.AssertNotCalled(t, "Find", mock.Anything, mock.Anything)
	db.Scene.AssertNotCalled(t, "UpdatePartial", mock.Anything, mock.Anything, mock.Anything)
}

func TestApplier_BadNFOIsSkipped(t *testing.T) {
	videoPath := writeSample(t, `<movie><title>a & b</title></movie>`)

	db := mocks.NewDatabase()
	applier := &Applier{Repo: db.Repository()}

	require.NoError(t, applier.Apply(context.Background(), 1, videoPath))

	db.Scene.AssertNotCalled(t, "Find", mock.Anything, mock.Anything)
	db.Scene.AssertNotCalled(t, "UpdatePartial", mock.Anything, mock.Anything, mock.Anything)
}

func TestApplier_FillsEmptyScene(t *testing.T) {
	videoPath := writeSample(t, sampleContents(t))

	db := mocks.NewDatabase()
	sceneID := 1

	db.Scene.On("Find", mock.Anything, sceneID).Return(&models.Scene{ID: sceneID}, nil)
	db.Studio.On("FindByName", mock.Anything, "ファンタドリーム", true).Return(&models.Studio{ID: 10, Name: "ファンタドリーム"}, nil)

	db.Tag.On("FindByNames", mock.Anything, mock.Anything, true).Return([]*models.Tag{}, nil)
	db.Tag.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		args.Get(1).(*models.CreateTagInput).Tag.ID = 100
	})

	db.Performer.On("FindByNames", mock.Anything, mock.Anything, true).Return([]*models.Performer{}, nil)
	db.Performer.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		args.Get(1).(*models.CreatePerformerInput).Performer.ID = 200
	})

	var captured models.ScenePartial
	db.Scene.On("UpdatePartial", mock.Anything, sceneID, mock.Anything).
		Run(func(args mock.Arguments) { captured = args.Get(2).(models.ScenePartial) }).
		Return(&models.Scene{ID: sceneID}, nil)

	applier := &Applier{Repo: db.Repository()}
	require.NoError(t, applier.Apply(context.Background(), sceneID, videoPath))

	require.True(t, captured.Title.Set)
	require.Equal(t, "FDD-2002 ビッグ·ティッツ·ビューティー CD1", captured.Title.Value)
	require.True(t, captured.Code.Set)
	require.Equal(t, "FDD-2002", captured.Code.Value)
	require.True(t, captured.Date.Set)
	require.Equal(t, "2002-06-28", captured.Date.Value.String())
	require.True(t, captured.StudioID.Set)
	require.Equal(t, 10, captured.StudioID.Value)

	require.NotNil(t, captured.TagIDs)
	require.Equal(t, models.RelationshipUpdateModeAdd, captured.TagIDs.Mode)
	require.Len(t, captured.TagIDs.IDs, 4)

	require.NotNil(t, captured.PerformerIDs)
	require.Equal(t, models.RelationshipUpdateModeAdd, captured.PerformerIDs.Mode)
	require.Len(t, captured.PerformerIDs.IDs, 3)

	require.NotNil(t, captured.URLs)
	require.Equal(t, models.RelationshipUpdateModeAdd, captured.URLs.Mode)

	db.Tag.AssertNumberOfCalls(t, "Create", 4)
	db.Performer.AssertNumberOfCalls(t, "Create", 3)
}

func TestApplier_DoesNotOverwriteExistingFields(t *testing.T) {
	videoPath := writeSample(t, sampleContents(t))

	db := mocks.NewDatabase()
	sceneID := 1

	date := sampleDate(t)
	rating := 42
	studioID := 7

	existing := &models.Scene{
		ID:             sceneID,
		Title:          "user title",
		Code:           "user code",
		Details:        "user details",
		Director:       "user director",
		Date:           &date,
		ProductionDate: &date,
		Rating:         &rating,
		StudioID:       &studioID,
	}

	db.Scene.On("Find", mock.Anything, sceneID).Return(existing, nil)
	db.Tag.On("FindByNames", mock.Anything, mock.Anything, true).Return([]*models.Tag{}, nil)
	db.Tag.On("Create", mock.Anything, mock.Anything).Return(nil)
	db.Performer.On("FindByNames", mock.Anything, mock.Anything, true).Return([]*models.Performer{}, nil)
	db.Performer.On("Create", mock.Anything, mock.Anything).Return(nil)

	var captured models.ScenePartial
	db.Scene.On("UpdatePartial", mock.Anything, sceneID, mock.Anything).
		Run(func(args mock.Arguments) { captured = args.Get(2).(models.ScenePartial) }).
		Return(existing, nil)

	applier := &Applier{Repo: db.Repository()}
	require.NoError(t, applier.Apply(context.Background(), sceneID, videoPath))

	require.False(t, captured.Title.Set)
	require.False(t, captured.Code.Set)
	require.False(t, captured.Details.Set)
	require.False(t, captured.Director.Set)
	require.False(t, captured.Date.Set)
	require.False(t, captured.ProductionDate.Set)
	require.False(t, captured.Rating.Set)
	require.False(t, captured.StudioID.Set)

	// Studio already set: must not look it up or create an unused one.
	db.Studio.AssertNotCalled(t, "FindByName", mock.Anything, mock.Anything, mock.Anything)

	// Relations are additive and still applied.
	require.NotNil(t, captured.TagIDs)
	require.NotNil(t, captured.PerformerIDs)
}

func TestApplier_ReusesExistingTagsAndPerformers(t *testing.T) {
	videoPath := writeSample(t, sampleContents(t))

	db := mocks.NewDatabase()
	sceneID := 1

	db.Scene.On("Find", mock.Anything, sceneID).Return(&models.Scene{ID: sceneID}, nil)
	db.Studio.On("FindByName", mock.Anything, mock.Anything, true).Return(&models.Studio{ID: 10}, nil)

	db.Tag.On("FindByNames", mock.Anything, mock.Anything, true).Return([]*models.Tag{
		{ID: 1, Name: "FDD"},
		{ID: 2, Name: "无码"},
		{ID: 3, Name: "BigTitsBeauty"},
		{ID: 4, Name: "ファンタドリーム"},
	}, nil)

	db.Performer.On("FindByNames", mock.Anything, mock.Anything, true).Return([]*models.Performer{
		{ID: 11, Name: "夕樹洋子"},
		{ID: 12, Name: "松坂季美子"},
		{ID: 13, Name: "加山なつこ"},
	}, nil)

	var captured models.ScenePartial
	db.Scene.On("UpdatePartial", mock.Anything, sceneID, mock.Anything).
		Run(func(args mock.Arguments) { captured = args.Get(2).(models.ScenePartial) }).
		Return(&models.Scene{ID: sceneID}, nil)

	applier := &Applier{Repo: db.Repository()}
	require.NoError(t, applier.Apply(context.Background(), sceneID, videoPath))

	db.Tag.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	db.Performer.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)

	require.ElementsMatch(t, []int{1, 2, 3, 4}, captured.TagIDs.IDs)
	require.ElementsMatch(t, []int{11, 12, 13}, captured.PerformerIDs.IDs)
}

func sampleDate(t *testing.T) models.Date {
	t.Helper()

	d, err := models.ParseDate("2001-01-01")
	require.NoError(t, err)

	return d
}
