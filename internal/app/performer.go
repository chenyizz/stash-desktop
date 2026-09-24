package app

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"case/backend/pkg/logger"
	"case/backend/pkg/models"
	"case/backend/pkg/utils"
)

// PerformerDetailDTO 是演员详情。按展示策略剔除：Ethnicity/HairColor/EyeColor/
// PenisLength/Circumcised/FakeTits/Piercings/StashIDs（见 docs/TECH_DEBT.md）。
type PerformerDetailDTO struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Disambiguation string `json:"disambiguation"`
	Gender         string `json:"gender"`
	Birthdate      string `json:"birthdate"`
	DeathDate      string `json:"deathDate"`
	Country        string `json:"country"`
	Height         *int   `json:"height"`
	Weight         *int   `json:"weight"`
	Measurements   string `json:"measurements"`
	CareerStart    string `json:"careerStart"`
	CareerEnd      string `json:"careerEnd"`
	Tattoos        string `json:"tattoos"`
	Favorite       bool   `json:"favorite"`
	// Rating expressed in 1-100 scale; 0 means no rating.
	Rating    int      `json:"rating"`
	Details   string   `json:"details"`
	Aliases   []string `json:"aliases"`
	URLs      []string `json:"urls"`
	Tags      []TagDTO `json:"tags"`
	ImageURL  string   `json:"imageUrl"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
}

// GetPerformer 返回演员详情。
func (a *App) GetPerformer(id int) (*PerformerDetailDTO, error) {
	if a.mgr == nil {
		return nil, fmt.Errorf("manager 未初始化")
	}

	var dto *PerformerDetailDTO

	err := a.mgr.Repository.WithReadTxn(context.Background(), func(ctx context.Context) error {
		p, err := a.mgr.Repository.Performer.Find(ctx, id)
		if err != nil {
			return err
		}
		if p == nil {
			return fmt.Errorf("演员不存在: %d", id)
		}

		if err := p.LoadTagIDs(ctx, a.mgr.Repository.Performer); err != nil {
			return fmt.Errorf("加载标签关联失败: %w", err)
		}
		if err := p.LoadURLs(ctx, a.mgr.Repository.Performer); err != nil {
			return fmt.Errorf("加载链接失败: %w", err)
		}
		if err := p.LoadAliases(ctx, a.mgr.Repository.Performer); err != nil {
			return fmt.Errorf("加载别名失败: %w", err)
		}

		tags, err := a.findTags(ctx, p.TagIDs.List())
		if err != nil {
			return err
		}

		hasImage, err := a.mgr.Repository.Performer.HasImage(ctx, id)
		if err != nil {
			hasImage = false
		}

		dto = toPerformerDetailDTO(p, tags, hasImage)
		return nil
	})
	if err != nil {
		return nil, err
	}

	logger.Infof("查询演员详情: id=%d", id)
	return dto, nil
}

func toPerformerDetailDTO(p *models.Performer, tags []*models.Tag, hasImage bool) *PerformerDetailDTO {
	dto := &PerformerDetailDTO{
		ID:             p.ID,
		Name:           p.Name,
		Disambiguation: p.Disambiguation,
		Gender:         genderString(p.Gender),
		Birthdate:      formatDate(p.Birthdate),
		DeathDate:      formatDate(p.DeathDate),
		Country:        p.Country,
		Height:         p.Height,
		Weight:         p.Weight,
		Measurements:   p.Measurements,
		CareerStart:    formatDate(p.CareerStart),
		CareerEnd:      formatDate(p.CareerEnd),
		Tattoos:        p.Tattoos,
		Favorite:       p.Favorite,
		Details:        p.Details,
		Aliases:        p.Aliases.List(),
		URLs:           p.URLs.List(),
		Tags:           toTagDTOs(p.TagIDs.List(), tags),
		CreatedAt:      p.CreatedAt.Format(dtoTimeLayout),
		UpdatedAt:      p.UpdatedAt.Format(dtoTimeLayout),
	}

	if p.Rating != nil {
		dto.Rating = *p.Rating
	}
	if hasImage {
		dto.ImageURL = fmt.Sprintf("%s%d/image", performerImageURLPrefix, p.ID)
	}

	return dto
}

func genderString(g *models.GenderEnum) string {
	if g == nil {
		return ""
	}
	return string(*g)
}

// handlePerformerImage 处理 /performers/<id>/image。
func (a *App) handlePerformerImage(w http.ResponseWriter, r *http.Request) {
	if a.mgr == nil {
		http.NotFound(w, r)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, performerImageURLPrefix), "/")
	if len(parts) != 2 || parts[1] != "image" {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}

	var data []byte
	err = a.mgr.Repository.WithReadTxn(r.Context(), func(ctx context.Context) error {
		var readErr error
		data, readErr = a.mgr.Repository.Performer.GetImage(ctx, id)
		return readErr
	})
	if err != nil || len(data) == 0 {
		http.NotFound(w, r)
		return
	}

	utils.ServeImage(w, r, data)
}
