package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"case/backend/pkg/metadata/attachments"
	"case/backend/pkg/utils"
)

// AttachmentDTO 是详情页的附件项。
type AttachmentDTO struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// handleAttachment 处理 /attachments/<sceneID>/<index>：
// 通过 sceneID 重新解析附件并按索引定位，避免暴露任意文件路径。
func (a *App) handleAttachment(w http.ResponseWriter, r *http.Request) {
	if a.mgr == nil {
		http.NotFound(w, r)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, attachmentURLPrefix), "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	sceneID, sceneErr := strconv.Atoi(parts[0])
	index, indexErr := strconv.Atoi(parts[1])
	if sceneErr != nil || indexErr != nil || sceneID <= 0 || index < 0 {
		http.NotFound(w, r)
		return
	}

	videoPath, err := a.scenePrimaryPath(r.Context(), sceneID)
	if err != nil || videoPath == "" {
		http.NotFound(w, r)
		return
	}

	list := attachments.Resolve(videoPath)
	if index >= len(list) {
		http.NotFound(w, r)
		return
	}

	data, err := os.ReadFile(list[index].Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	utils.ServeImage(w, r, data)
}

func (a *App) scenePrimaryPath(ctx context.Context, sceneID int) (string, error) {
	var path string

	err := a.mgr.Repository.WithReadTxn(ctx, func(ctx context.Context) error {
		s, err := a.mgr.Repository.Scene.Find(ctx, sceneID)
		if err != nil {
			return err
		}
		if s == nil {
			return nil
		}
		if err := s.LoadPrimaryFile(ctx, a.mgr.Repository.File); err != nil {
			return err
		}
		if primary := s.Files.Primary(); primary != nil {
			path = primary.Path
		} else {
			path = s.Path
		}
		return nil
	})

	return path, err
}

// resolveAttachments 把文件系统附件映射为带索引 URL 的 DTO。
func (a *App) resolveAttachments(sceneID int, videoPath string) []AttachmentDTO {
	if !a.mgr.Config.GetLibraryMode(videoPath).Attachments {
		return nil
	}

	list := attachments.Resolve(videoPath)
	if len(list) == 0 {
		return nil
	}

	out := make([]AttachmentDTO, 0, len(list))
	for i, at := range list {
		out = append(out, AttachmentDTO{
			Name:   at.Name,
			URL:    fmt.Sprintf("%s%d/%d", attachmentURLPrefix, sceneID, i),
			Width:  at.Width,
			Height: at.Height,
		})
	}

	return out
}
