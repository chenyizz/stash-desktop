package attachments

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})

	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	require.NoError(t, os.WriteFile(path, buf.Bytes(), 0o644))
}

func TestResolve_FindsAttachmentsInKnownDirs(t *testing.T) {
	dir := t.TempDir()
	video := filepath.Join(dir, "FDD-2002.mp4")
	require.NoError(t, os.WriteFile(video, []byte("video"), 0o644))

	writePNG(t, filepath.Join(dir, "fanart", "a.png"), 100, 50)
	writePNG(t, filepath.Join(dir, "poster", "b.png"), 200, 300)
	writePNG(t, filepath.Join(dir, "extra", "c.png"), 10, 10)

	got := Resolve(video)
	require.Len(t, got, 3)

	assert.Equal(t, "a.png", got[0].Name)
	assert.Equal(t, 100, got[0].Width)
	assert.Equal(t, 50, got[0].Height)
	assert.Equal(t, "b.png", got[1].Name)
	assert.Equal(t, 200, got[1].Width)
	assert.Equal(t, "c.png", got[2].Name)
}

func TestResolve_IgnoresUnknownDirsAndNonImages(t *testing.T) {
	dir := t.TempDir()
	video := filepath.Join(dir, "v.mp4")
	require.NoError(t, os.WriteFile(video, []byte("video"), 0o644))

	writePNG(t, filepath.Join(dir, "other", "x.png"), 10, 10)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "fanart"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "fanart", "notes.txt"), []byte("hi"), 0o644))

	assert.Empty(t, Resolve(video))
}

func TestResolve_EmptyForBlankPath(t *testing.T) {
	assert.Empty(t, Resolve(""))
}
