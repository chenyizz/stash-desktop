package config

import (
	"path/filepath"

	"case/backend/pkg/fsutil"
)

// LibraryMode controls what a single library path scans.
// 指针字段用于区分「未设置」（nil → 默认全开）与显式配置。
type LibraryMode struct {
	Videos      bool `json:"videos"`
	Images      bool `json:"images"`
	Attachments bool `json:"attachments"`
}

// Effective returns the effective mode: unset (nil) defaults to all enabled.
func (m *LibraryMode) Effective() LibraryMode {
	if m == nil {
		return LibraryMode{Videos: true, Images: true, Attachments: true}
	}

	return *m
}

// Library configuration details
type LibraryConfigInput struct {
	Path         string       `json:"path"`
	ExcludeVideo bool         `json:"excludeVideo"`
	ExcludeImage bool         `json:"excludeImage"`
	Mode         *LibraryMode `json:"mode,omitempty"`
}

type LibraryConfig struct {
	Path         string       `json:"path"`
	ExcludeVideo bool         `json:"excludeVideo"`
	ExcludeImage bool         `json:"excludeImage"`
	Mode         *LibraryMode `json:"mode,omitempty"`
}

type LibraryConfigs []*LibraryConfig

// GetLibraryMode returns the effective scan mode for a file path.
// 未配置 mode 的库路径默认 {Videos:true, Images:true, Attachments:true}。
func (i *Config) GetLibraryMode(path string) LibraryMode {
	i.RLock()
	defer i.RUnlock()

	var libs LibraryConfigs
	if err := i.forKey(Libraries).Unmarshal(Libraries, &libs); err != nil {
		return LibraryMode{Videos: true, Images: true, Attachments: true}
	}

	lib := libs.GetLibraryFromPath(path)
	if lib == nil {
		return LibraryMode{Videos: true, Images: true, Attachments: true}
	}

	return lib.Mode.Effective()
}

// GetLibraryFromPath returns the most specific library configuration containing path.
func (s LibraryConfigs) GetLibraryFromPath(path string) *LibraryConfig {
	return s.GetLibraryFromDirPath(filepath.Dir(path))
}

// GetLibraryFromDirPath returns the most specific library configuration containing dirPath.
func (s LibraryConfigs) GetLibraryFromDirPath(dirPath string) *LibraryConfig {
	var ret *LibraryConfig
	longestPath := -1

	for _, f := range s {
		if f == nil {
			continue
		}

		path := filepath.Clean(f.Path)
		if fsutil.IsPathInDir(path, dirPath) && len(path) > longestPath {
			ret = f
			longestPath = len(path)
		}
	}

	return ret
}

// GetLibraryRootFromDirPath returns the topmost configured library path containing dirPath.
func (s LibraryConfigs) GetLibraryRootFromDirPath(dirPath string) string {
	var ret string
	shortestPath := -1

	for _, f := range s {
		if f == nil {
			continue
		}

		path := filepath.Clean(f.Path)
		if fsutil.IsPathInDir(path, dirPath) && (shortestPath == -1 || len(path) < shortestPath) {
			ret = path
			shortestPath = len(path)
		}
	}

	return ret
}

func (s LibraryConfigs) Paths() []string {
	paths := make([]string, len(s))
	for i, c := range s {
		// #6618 - clean the path to ensure comparison works correctly
		paths[i] = filepath.Clean(c.Path)
	}
	return paths
}
