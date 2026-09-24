// Package attachments resolves image attachments (fanart/poster/extra) that
// live next to a video file. Attachments are filesystem-only and are not
// persisted; see docs/DESIGN_MEDIA_SCAN.md (D2 方案 A).
package attachments

import (
	"image"
	"os"
	"path/filepath"
	"strings"

	// Image decoders used to read dimensions.
	_ "image/jpeg"
	_ "image/png"
)

// DefaultDirs are the well-known attachment subdirectories, scanned in order.
var DefaultDirs = []string{"fanart", "poster", "extra"}

// Attachment is a single image file found next to a video.
type Attachment struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Resolve returns the image attachments for videoPath by scanning DefaultDirs
// in the video's directory. Missing directories are ignored; undecodable files
// are skipped. Results are ordered by directory then filename.
func Resolve(videoPath string) []Attachment {
	if strings.TrimSpace(videoPath) == "" {
		return nil
	}

	dir := filepath.Dir(videoPath)

	var out []Attachment
	for _, sub := range DefaultDirs {
		entries, err := os.ReadDir(filepath.Join(dir, sub))
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !isImageExt(entry.Name()) {
				continue
			}

			path := filepath.Join(dir, sub, entry.Name())
			width, height, err := imageSize(path)
			if err != nil {
				continue
			}

			out = append(out, Attachment{
				Name:   entry.Name(),
				Path:   path,
				Width:  width,
				Height: height,
			})
		}
	}

	return out
}

func isImageExt(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}

func imageSize(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}

	return cfg.Width, cfg.Height, nil
}
