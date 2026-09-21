// Package scene provides the application logic for scene functionality.
// Most functionality is provided by [Service].
package scene

import (
	"case/backend/pkg/models"
	"case/backend/pkg/models/paths"
	"case/backend/pkg/plugin"
)

type Config interface {
	GetVideoFileNamingAlgorithm() models.HashAlgorithm
}

type Service struct {
	File             models.FileReaderWriter
	Repository       models.SceneReaderWriter
	MarkerRepository models.SceneMarkerReaderWriter
	PluginCache      *plugin.Cache

	Paths  *paths.Paths
	Config Config
}
