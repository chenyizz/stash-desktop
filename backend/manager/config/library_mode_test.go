package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLibraryMode_Effective(t *testing.T) {
	var nilMode *LibraryMode
	assert.Equal(t, LibraryMode{Videos: true, Images: true, Attachments: true}, nilMode.Effective())

	m := LibraryMode{Videos: true}
	assert.Equal(t, LibraryMode{Videos: true}, (&m).Effective())
}

func TestGetLibraryMode(t *testing.T) {
	cfg := InitializeEmpty()
	cfg.SetInterface(Libraries, LibraryConfigs{
		{Path: "/video", Mode: &LibraryMode{Videos: true, Attachments: true}},
		{Path: "/image", Mode: &LibraryMode{Images: true}},
		{Path: "/default"},
	})

	assert.Equal(t, LibraryMode{Videos: true, Attachments: true}, cfg.GetLibraryMode("/video/a.mp4"))
	assert.Equal(t, LibraryMode{Images: true}, cfg.GetLibraryMode("/image/b.jpg"))
	assert.Equal(t, LibraryMode{Videos: true, Images: true, Attachments: true}, cfg.GetLibraryMode("/default/c.mp4"))
	assert.Equal(t, LibraryMode{Videos: true, Images: true, Attachments: true}, cfg.GetLibraryMode("/unknown/x.mp4"))
}

func TestValidate_RejectsBothDisabled(t *testing.T) {
	cfg := InitializeEmpty()
	cfg.SetInterface(Database, "/tmp/case.db")
	cfg.SetInterface(Generated, "/tmp/gen")
	cfg.SetInterface(Libraries, LibraryConfigs{{Path: "/x", Mode: &LibraryMode{}}})

	require.Error(t, cfg.Validate())
}

func TestValidate_AllowsDefaultMode(t *testing.T) {
	cfg := InitializeEmpty()
	cfg.SetInterface(Database, "/tmp/case.db")
	cfg.SetInterface(Generated, "/tmp/gen")
	cfg.SetInterface(Libraries, LibraryConfigs{{Path: "/x"}})

	require.NoError(t, cfg.Validate())
}
