// Package paths provides functions to return paths to various resources.
package paths

import (
	"path/filepath"

	"case/backend/pkg/fsutil"
)

type Paths struct {
	Generated *generatedPaths

	Scene        *scenePaths
	SceneMarkers *sceneMarkerPaths
	Blobs        string
}

func NewPaths(generatedPath string, blobsPath string) Paths {
	p := Paths{}
	p.Generated = newGeneratedPaths(generatedPath)

	p.Scene = newScenePaths(p)
	p.SceneMarkers = newSceneMarkerPaths(p)
	p.Blobs = blobsPath

	return p
}

// GetCaseHomeDirectory returns the default per-user config directory.
// 仅作为未配置时的回退默认值；已显式配置的路径不受影响。
func GetCaseHomeDirectory() string {
	return filepath.Join(fsutil.GetHomeDirectory(), ".case")
}

// GetDefaultDatabaseFilePath returns the default database path.
// 仅作为未配置时的回退默认值；已显式配置的路径不受影响。
func GetDefaultDatabaseFilePath() string {
	return filepath.Join(GetCaseHomeDirectory(), "case.db")
}
