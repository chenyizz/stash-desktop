package app

import (
	"context"
	"fmt"
	"strconv"

	"case/backend/pkg/logger"
	"case/backend/pkg/models"
)

const dtoTimeLayout = "2006-01-02 15:04:05"

// TagDTO pairs a tag ID with its name so the frontend can look up either
// without relying on parallel arrays.
type TagDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// PerformerDTO pairs a performer ID with its name.
type PerformerDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SceneFileDTO is a single video file of a scene.
type SceneFileDTO struct {
	ID         int     `json:"id"`
	Path       string  `json:"path"`
	Basename   string  `json:"basename"`
	Duration   float64 `json:"duration"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	VideoCodec string  `json:"videoCodec"`
	AudioCodec string  `json:"audioCodec"`
	FrameRate  float64 `json:"frameRate"`
	BitRate    int64   `json:"bitrate"`
	Format     string  `json:"format"`
}

// SceneDetailDTO is the full representation returned by GetScene.
type SceneDetailDTO struct {
	ID             int    `json:"id"`
	Title          string `json:"title"`
	Code           string `json:"code"`
	Details        string `json:"details"`
	Director       string `json:"director"`
	Date           string `json:"date"`
	ProductionDate string `json:"productionDate"`
	// Rating is expressed on a 1-100 scale; 0 means no rating.
	Rating     int      `json:"rating"`
	Organized  bool     `json:"organized"`
	StudioID   *int     `json:"studioId"`
	StudioName string   `json:"studioName"`
	URLs       []string `json:"urls"`

	Tags       []TagDTO       `json:"tags"`
	Performers []PerformerDTO `json:"performers"`

	// Convenience fields for the primary file. Duration/FrameRate are 0 when
	// unknown; Resolution is empty when dimensions are unknown.
	Path       string  `json:"path"`
	OSHash     string  `json:"oshash"`
	Checksum   string  `json:"checksum"`
	Duration   float64 `json:"duration"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Resolution string  `json:"resolution"`
	VideoCodec string  `json:"videoCodec"`
	AudioCodec string  `json:"audioCodec"`
	FrameRate  float64 `json:"frameRate"`
	BitRate    int64   `json:"bitrate"`

	Files     []SceneFileDTO `json:"files"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
}

// GetScene returns the full detail DTO for a single scene.
func (a *App) GetScene(id int) (*SceneDetailDTO, error) {
	if a.mgr == nil {
		return nil, fmt.Errorf("manager 未初始化")
	}

	var dto *SceneDetailDTO

	err := a.mgr.Repository.WithReadTxn(context.Background(), func(ctx context.Context) error {
		s, err := a.mgr.Repository.Scene.Find(ctx, id)
		if err != nil {
			return err
		}
		if s == nil {
			return fmt.Errorf("场景不存在: %d", id)
		}

		if err := s.LoadRelationships(ctx, a.mgr.Repository.Scene); err != nil {
			return fmt.Errorf("加载场景关联失败: %w", err)
		}

		tags, err := a.findTags(ctx, s.TagIDs.List())
		if err != nil {
			return err
		}

		performers, err := a.findPerformers(ctx, s.PerformerIDs.List())
		if err != nil {
			return err
		}

		studio, err := a.findStudio(ctx, s.StudioID)
		if err != nil {
			return err
		}

		dto = toSceneDetailDTO(s, tags, performers, studio)
		return nil
	})
	if err != nil {
		return nil, err
	}

	logger.Infof("查询场景详情: id=%d", id)
	return dto, nil
}

func (a *App) findTags(ctx context.Context, ids []int) ([]*models.Tag, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	tags, err := a.mgr.Repository.Tag.FindMany(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}

	return tags, nil
}

func (a *App) findPerformers(ctx context.Context, ids []int) ([]*models.Performer, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	performers, err := a.mgr.Repository.Performer.FindMany(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("查询演员失败: %w", err)
	}

	return performers, nil
}

func (a *App) findStudio(ctx context.Context, id *int) (*models.Studio, error) {
	if id == nil {
		return nil, nil
	}

	studio, err := a.mgr.Repository.Studio.Find(ctx, *id)
	if err != nil {
		return nil, fmt.Errorf("查询工作室失败: %w", err)
	}

	return studio, nil
}

func toSceneDetailDTO(s *models.Scene, tags []*models.Tag, performers []*models.Performer, studio *models.Studio) *SceneDetailDTO {
	dto := &SceneDetailDTO{
		ID:             s.ID,
		Title:          s.GetTitle(),
		Code:           s.Code,
		Details:        s.Details,
		Director:       s.Director,
		Date:           formatDate(s.Date),
		ProductionDate: formatDate(s.ProductionDate),
		Organized:      s.Organized,
		StudioID:       s.StudioID,
		URLs:           s.URLs.List(),
		Tags:           toTagDTOs(s.TagIDs.List(), tags),
		Performers:     toPerformerDTOs(s.PerformerIDs.List(), performers),
		Path:           s.Path,
		OSHash:         s.OSHash,
		Checksum:       s.Checksum,
		Files:          toSceneFileDTOs(s.Files.List()),
		CreatedAt:      s.CreatedAt.Format(dtoTimeLayout),
		UpdatedAt:      s.UpdatedAt.Format(dtoTimeLayout),
	}

	if s.Rating != nil {
		dto.Rating = *s.Rating
	}
	if studio != nil {
		dto.StudioName = studio.Name
	}

	if primary := s.Files.Primary(); primary != nil {
		dto.Path = primary.Path
		dto.Duration = primary.DurationFinite()
		dto.Width = primary.Width
		dto.Height = primary.Height
		dto.Resolution = resolutionLabel(primary.Width, primary.Height)
		dto.VideoCodec = primary.VideoCodec
		dto.AudioCodec = primary.AudioCodec
		dto.FrameRate = primary.FrameRateFinite()
		dto.BitRate = primary.BitRate
	}

	return dto
}

func toTagDTOs(ids []int, tags []*models.Tag) []TagDTO {
	if len(ids) == 0 {
		return nil
	}

	names := make(map[int]string, len(tags))
	for _, t := range tags {
		names[t.ID] = t.Name
	}

	out := make([]TagDTO, 0, len(ids))
	for _, id := range ids {
		out = append(out, TagDTO{ID: id, Name: names[id]})
	}

	return out
}

func toPerformerDTOs(ids []int, performers []*models.Performer) []PerformerDTO {
	if len(ids) == 0 {
		return nil
	}

	names := make(map[int]string, len(performers))
	for _, p := range performers {
		names[p.ID] = p.Name
	}

	out := make([]PerformerDTO, 0, len(ids))
	for _, id := range ids {
		out = append(out, PerformerDTO{ID: id, Name: names[id]})
	}

	return out
}

func toSceneFileDTOs(files []*models.VideoFile) []SceneFileDTO {
	if len(files) == 0 {
		return nil
	}

	out := make([]SceneFileDTO, 0, len(files))
	for _, f := range files {
		out = append(out, SceneFileDTO{
			ID:         int(f.ID),
			Path:       f.Path,
			Basename:   f.Basename,
			Duration:   f.DurationFinite(),
			Width:      f.Width,
			Height:     f.Height,
			VideoCodec: f.VideoCodec,
			AudioCodec: f.AudioCodec,
			FrameRate:  f.FrameRateFinite(),
			BitRate:    f.BitRate,
			Format:     f.Format,
		})
	}

	return out
}

func formatDate(d *models.Date) string {
	if d == nil {
		return ""
	}

	return d.String()
}

// resolutionLabel derives a human-readable label (e.g. "1080p") from the
// smaller dimension, which works for both landscape and portrait video.
func resolutionLabel(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	minDimension := width
	if height < minDimension {
		minDimension = height
	}

	return strconv.Itoa(minDimension) + "p"
}
