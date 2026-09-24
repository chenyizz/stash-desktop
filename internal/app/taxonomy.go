package app

import (
	"context"
	"fmt"

	"case/backend/pkg/logger"
	"case/backend/pkg/models"
)

// StudioDTO 是工作室的列表项。
type StudioDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type PerformersPageDTO struct {
	Performers []PerformerDTO `json:"performers"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"pageSize"`
}

type TagsPageDTO struct {
	Tags     []TagDTO `json:"tags"`
	Total    int      `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"pageSize"`
}

type StudiosPageDTO struct {
	Studios  []StudioDTO `json:"studios"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// FindPerformers 分页查询演员列表。
func (a *App) FindPerformers(page int, pageSize int) (*PerformersPageDTO, error) {
	if a.mgr == nil {
		return nil, fmt.Errorf("manager 未初始化")
	}
	page, pageSize = normalizePage(page, pageSize)

	var items []PerformerDTO
	total := 0
	err := a.mgr.Repository.WithReadTxn(context.Background(), func(ctx context.Context) error {
		findFilter := &models.FindFilterType{Page: &page, PerPage: &pageSize}
		performers, count, err := a.mgr.Repository.Performer.Query(ctx, &models.PerformerFilterType{}, findFilter)
		if err != nil {
			return err
		}
		total = count
		for _, p := range performers {
			items = append(items, PerformerDTO{ID: p.ID, Name: p.Name})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("查询演员失败: %w", err)
	}

	logger.Infof("查询演员: page=%d, pageSize=%d, total=%d", page, pageSize, total)

	return &PerformersPageDTO{Performers: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// FindTags 分页查询标签列表。
func (a *App) FindTags(page int, pageSize int) (*TagsPageDTO, error) {
	if a.mgr == nil {
		return nil, fmt.Errorf("manager 未初始化")
	}
	page, pageSize = normalizePage(page, pageSize)

	var items []TagDTO
	total := 0
	err := a.mgr.Repository.WithReadTxn(context.Background(), func(ctx context.Context) error {
		findFilter := &models.FindFilterType{Page: &page, PerPage: &pageSize}
		tags, count, err := a.mgr.Repository.Tag.Query(ctx, &models.TagFilterType{}, findFilter)
		if err != nil {
			return err
		}
		total = count
		for _, t := range tags {
			items = append(items, TagDTO{ID: t.ID, Name: t.Name})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}

	logger.Infof("查询标签: page=%d, pageSize=%d, total=%d", page, pageSize, total)

	return &TagsPageDTO{Tags: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// FindStudios 分页查询工作室列表。
func (a *App) FindStudios(page int, pageSize int) (*StudiosPageDTO, error) {
	if a.mgr == nil {
		return nil, fmt.Errorf("manager 未初始化")
	}
	page, pageSize = normalizePage(page, pageSize)

	var items []StudioDTO
	total := 0
	err := a.mgr.Repository.WithReadTxn(context.Background(), func(ctx context.Context) error {
		findFilter := &models.FindFilterType{Page: &page, PerPage: &pageSize}
		studios, count, err := a.mgr.Repository.Studio.Query(ctx, &models.StudioFilterType{}, findFilter)
		if err != nil {
			return err
		}
		total = count
		for _, s := range studios {
			items = append(items, StudioDTO{ID: s.ID, Name: s.Name})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("查询工作室失败: %w", err)
	}

	logger.Infof("查询工作室: page=%d, pageSize=%d, total=%d", page, pageSize, total)

	return &StudiosPageDTO{Studios: items, Total: total, Page: page, PageSize: pageSize}, nil
}
