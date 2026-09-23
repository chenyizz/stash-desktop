// Package cover resolves a scene cover image from local sources, in priority
// order, falling back to an FFmpeg frame grab.
package cover

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// Decoders used to validate candidate images.
	_ "image/jpeg"
	_ "image/png"

	"case/backend/pkg/logger"
)

// DefaultTimeProportion is the position (fraction of duration) used when
// grabbing a frame with FFmpeg. Matches backend/pkg/scene/generate.
const DefaultTimeProportion = 0.2

// ScreenshotFunc grabs a frame from videoPath at the given time (seconds) and
// returns encoded image bytes.
type ScreenshotFunc func(ctx context.Context, videoPath string, at float64) ([]byte, error)

// Result is the resolved cover image and the name of the source that produced it.
type Result struct {
	Data   []byte
	Source string
}

type sourceFunc func(ctx context.Context, videoPath string, duration float64, screenshot ScreenshotFunc) ([]byte, bool, error)

type source struct {
	name    string
	resolve sourceFunc
}

// sources are tried in order; the first that yields data wins. Each source
// fails independently and is unit-testable on its own.
var sources = []source{
	{name: "poster", resolve: fromPoster},
	{name: "same-name", resolve: fromSameName},
	{name: "ffmpeg", resolve: fromFFmpeg},
}

// Resolve returns the first available cover for videoPath. It returns an error
// only for unexpected failures; "no cover found" is reported as an empty Result.
func Resolve(ctx context.Context, videoPath string, duration float64, screenshot ScreenshotFunc) (Result, error) {
	if strings.TrimSpace(videoPath) == "" {
		return Result{}, nil
	}

	for _, s := range sources {
		data, found, err := s.resolve(ctx, videoPath, duration, screenshot)
		if err != nil {
			logger.Warnf("cover source %q failed for %q: %v", s.name, videoPath, err)
			continue
		}
		if !found {
			continue
		}

		return Result{Data: data, Source: s.name}, nil
	}

	return Result{}, nil
}

// fromPoster looks for poster.jpg/poster.png next to the video.
func fromPoster(_ context.Context, videoPath string, _ float64, _ ScreenshotFunc) ([]byte, bool, error) {
	dir := filepath.Dir(videoPath)
	for _, name := range []string{"poster.jpg", "poster.jpeg", "poster.png"} {
		data, found, err := readImage(filepath.Join(dir, name))
		if err != nil || found {
			return data, found, err
		}
	}

	return nil, false, nil
}

// fromSameName looks for an image sharing the video's basename.
func fromSameName(_ context.Context, videoPath string, _ float64, _ ScreenshotFunc) ([]byte, bool, error) {
	base := strings.TrimSuffix(videoPath, filepath.Ext(videoPath))
	for _, ext := range []string{".jpg", ".jpeg", ".png"} {
		data, found, err := readImage(base + ext)
		if err != nil || found {
			return data, found, err
		}
	}

	return nil, false, nil
}

// fromFFmpeg grabs a frame at DefaultTimeProportion of the duration.
func fromFFmpeg(ctx context.Context, videoPath string, duration float64, screenshot ScreenshotFunc) ([]byte, bool, error) {
	if screenshot == nil || duration <= 0 {
		return nil, false, nil
	}

	data, err := screenshot(ctx, videoPath, duration*DefaultTimeProportion)
	if err != nil {
		return nil, false, err
	}
	if len(data) == 0 {
		return nil, false, nil
	}

	return data, true, nil
}

// readImage reads path and returns its bytes if it is a decodable image.
// A missing file is reported as (nil, false, nil).
func readImage(path string) ([]byte, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("reading %q: %w", path, err)
	}

	if len(data) == 0 {
		return nil, false, nil
	}

	if err := validateImage(data); err != nil {
		logger.Warnf("ignoring %q as cover: %v", path, err)
		return nil, false, nil
	}

	return data, true, nil
}

func validateImage(data []byte) error {
	contentType := http.DetectContentType(data)
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("unsupported content type %q", contentType)
	}

	if _, _, err := image.DecodeConfig(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("decoding image: %w", err)
	}

	return nil
}
