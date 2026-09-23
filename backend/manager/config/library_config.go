package config

import (
	"path/filepath"

	"case/backend/pkg/fsutil"
)

// Library configuration details
type LibraryConfigInput struct {
	Path         string `json:"path"`
	ExcludeVideo bool   `json:"excludeVideo"`
	ExcludeImage bool   `json:"excludeImage"`
}

type LibraryConfig struct {
	Path         string `json:"path"`
	ExcludeVideo bool   `json:"excludeVideo"`
	ExcludeImage bool   `json:"excludeImage"`
}

type LibraryConfigs []*LibraryConfig

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
