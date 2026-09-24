package app

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/disintegration/imaging"

	"case/backend/pkg/hash/md5"
	"case/backend/pkg/logger"
	"case/backend/pkg/utils"
)

const (
	defaultThumbWidth = 320
	// 缩略图磁盘缓存上限；超过则按 mtime 从旧到新淘汰。
	maxThumbnailBytes = 256 << 20
	// 容量检查的节流间隔。
	thumbnailSweepInterval = 60 * time.Second
)

// 仅允许这些宽度，避免缓存键爆炸。
var allowedThumbWidths = map[int]struct{}{160: {}, 320: {}, 640: {}}

var lastThumbnailLimitCheck atomic.Int64

func (a *App) thumbnailDir() string {
	return filepath.Join(a.cfg.GetGeneratedPath(), "thumbnails")
}

// serveThumbnail 按场景 ID 提供缩放后的封面缩略图。
// ETag = 封面 checksum + 宽度；缓存文件按 checksum 定位，不依赖 DTO 传值。
func (a *App) serveThumbnail(w http.ResponseWriter, r *http.Request, sceneID, width int) {
	if _, ok := allowedThumbWidths[width]; !ok {
		width = defaultThumbWidth
	}

	data, err := a.readCover(r.Context(), sceneID)
	if err != nil || len(data) == 0 {
		http.NotFound(w, r)
		return
	}

	checksum := md5.FromBytes(data)
	etag := `"` + checksum + "-" + strconv.Itoa(width) + `"`
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	path := a.thumbnailPath(checksum, width)
	if cached, readErr := os.ReadFile(path); readErr == nil {
		utils.ServeImage(w, r, cached)
		return
	}

	thumb, genErr := resizeJPEG(data, width)
	if genErr != nil {
		logger.Warnf("生成缩略图失败 scene=%d: %v", sceneID, genErr)
		utils.ServeImage(w, r, data) // 降级为原图
		return
	}

	if writeErr := os.MkdirAll(a.thumbnailDir(), 0o755); writeErr == nil {
		if writeErr := writeFileAtomic(path, thumb); writeErr != nil {
			logger.Warnf("写缩略图缓存失败 %s: %v", path, writeErr)
		}
	}
	maybeEnforceThumbnailLimit(a.thumbnailDir())

	utils.ServeImage(w, r, thumb)
}

func (a *App) thumbnailPath(checksum string, width int) string {
	return filepath.Join(a.thumbnailDir(), checksum+"_"+strconv.Itoa(width)+".jpg")
}

func resizeJPEG(data []byte, width int) ([]byte, error) {
	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	thumb := imaging.Resize(img, width, 0, imaging.Lanczos)

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, thumb, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}

	return os.Rename(tmpName, path)
}

func maybeEnforceThumbnailLimit(dir string) {
	now := time.Now().Unix()
	last := lastThumbnailLimitCheck.Load()
	if now-last < int64(thumbnailSweepInterval.Seconds()) {
		return
	}
	if !lastThumbnailLimitCheck.CompareAndSwap(last, now) {
		return
	}

	enforceThumbnailLimit(dir)
}

// enforceThumbnailLimit 保证磁盘缓存不超过上限，按 mtime 从旧到新淘汰。
func enforceThumbnailLimit(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	type fileEntry struct {
		path string
		size int64
		mod  time.Time
	}

	var files []fileEntry
	var total int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileEntry{filepath.Join(dir, e.Name()), info.Size(), info.ModTime()})
		total += info.Size()
	}

	if total <= maxThumbnailBytes {
		return
	}

	sort.Slice(files, func(i, j int) bool { return files[i].mod.Before(files[j].mod) })
	for _, f := range files {
		if total <= maxThumbnailBytes {
			break
		}
		if os.Remove(f.path) == nil {
			total -= f.size
		}
	}
}

// SweepThumbnails 删除已无对应 blob 的孤立缩略图（覆盖场景删除/封面更新后的回收）。
// 在启动时调用一次；精确的随删随清需要 hook blob 删除（冻结区），见 TECH_DEBT。
func (a *App) SweepThumbnails(ctx context.Context) {
	dir := a.thumbnailDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var orphans []string
	err = a.mgr.Repository.WithReadTxn(ctx, func(ctx context.Context) error {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			idx := strings.IndexByte(name, '_')
			if idx <= 0 {
				continue
			}
			checksum := name[:idx]

			exists, entryErr := a.mgr.Repository.Blob.EntryExists(ctx, checksum)
			if entryErr != nil {
				return entryErr
			}
			if !exists {
				orphans = append(orphans, filepath.Join(dir, name))
			}
		}
		return nil
	})
	if err != nil {
		logger.Warnf("扫描孤立缩略图失败: %v", err)
		return
	}

	removed := 0
	for _, p := range orphans {
		if os.Remove(p) == nil {
			removed++
		}
	}
	if removed > 0 {
		logger.Infof("清理孤立缩略图 %d 个", removed)
	}
}
