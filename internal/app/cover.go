package app

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"net/http"
	"strconv"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"

	"case/backend/pkg/hash/md5"
	"case/backend/pkg/logger"
	"case/backend/pkg/metadata/cover"
	"case/backend/pkg/scene/generate"
	"case/backend/pkg/utils"
)

const coverURLPrefix = "/covers/"

// ensureCover returns the URL and dimensions of a scene's cover, generating and
// persisting one on demand when none exists yet. Failures are logged and
// reported as an empty URL so the detail page can fall back to a placeholder.
func (a *App) ensureCover(ctx context.Context, sceneID int, videoPath string, width int, duration float64) (string, int, int) {
	data, err := a.readCover(ctx, sceneID)
	if err != nil {
		logger.Warnf("读取封面失败 scene=%d: %v", sceneID, err)
		return "", 0, 0
	}

	if len(data) == 0 && videoPath != "" {
		result, resolveErr := cover.Resolve(ctx, videoPath, duration, a.screenshotFunc(videoPath, width, duration))
		if resolveErr != nil {
			logger.Warnf("解析封面失败 scene=%d: %v", sceneID, resolveErr)
		} else if len(result.Data) > 0 {
			if saveErr := a.mgr.Repository.WithTxn(ctx, func(ctx context.Context) error {
				return a.mgr.Repository.Scene.UpdateCover(ctx, sceneID, result.Data)
			}); saveErr != nil {
				logger.Warnf("保存封面失败 scene=%d: %v", sceneID, saveErr)
			} else {
				data = result.Data
				logger.Infof("生成封面成功 scene=%d source=%s", sceneID, result.Source)
			}
		}
	}

	if len(data) == 0 {
		return "", 0, 0
	}

	cfg, _, decodeErr := image.DecodeConfig(bytes.NewReader(data))
	if decodeErr != nil {
		logger.Warnf("封面无法解码，改用占位 scene=%d: %v", sceneID, decodeErr)
		return "", 0, 0
	}

	// The blob checksum is md5(bytes), so this changes whenever the cover does,
	// regardless of whether scenes.updated_at was touched.
	checksum := md5.FromBytes(data)

	return fmt.Sprintf("%s%d?v=%s", coverURLPrefix, sceneID, checksum[:8]), cfg.Width, cfg.Height
}

func (a *App) readCover(ctx context.Context, sceneID int) ([]byte, error) {
	var data []byte

	err := a.mgr.Repository.WithReadTxn(ctx, func(ctx context.Context) error {
		var readErr error
		data, readErr = a.mgr.Repository.Scene.GetCover(ctx, sceneID)
		return readErr
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

// screenshotFunc adapts the existing scene generator into a cover.ScreenshotFunc.
func (a *App) screenshotFunc(videoPath string, width int, duration float64) cover.ScreenshotFunc {
	return func(ctx context.Context, _ string, at float64) ([]byte, error) {
		generator := generate.Generator{
			Encoder:      a.mgr.FFMpeg,
			FFMpegConfig: a.mgr.Config,
			LockManager:  a.mgr.ReadLockManager,
			ScenePaths:   a.mgr.Paths.Scene,
			Overwrite:    true,
		}

		return generator.Screenshot(ctx, videoPath, width, duration, generate.ScreenshotOptions{At: &at})
	}
}

// CoverMiddleware serves scene covers at /covers/<sceneID>. It is a package
// function rather than an App method so that Wails does not bind it as a
// frontend-callable service method.
func CoverMiddleware(a *App) application.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, coverURLPrefix) {
				next.ServeHTTP(w, r)
				return
			}

			if a.mgr == nil {
				http.NotFound(w, r)
				return
			}

			sceneID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, coverURLPrefix))
			if err != nil || sceneID <= 0 {
				http.NotFound(w, r)
				return
			}

			data, err := a.readCover(r.Context(), sceneID)
			if err != nil || len(data) == 0 {
				http.NotFound(w, r)
				return
			}

			utils.ServeImage(w, r, data)
		})
	}
}
