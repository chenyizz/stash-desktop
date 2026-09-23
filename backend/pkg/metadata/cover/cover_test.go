package cover

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})

	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))

	return buf.Bytes()
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, data, 0o644))
}

func videoPath(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	p := filepath.Join(dir, "FDD-2002.mp4")
	writeFile(t, p, []byte("video"))

	return p
}

func TestResolve_PosterPriority(t *testing.T) {
	video := videoPath(t)
	dir := filepath.Dir(video)

	writeFile(t, filepath.Join(dir, "poster.png"), pngBytes(t, 2, 3))
	writeFile(t, filepath.Join(dir, "FDD-2002.png"), pngBytes(t, 3, 4))

	called := false
	screenshot := func(context.Context, string, float64) ([]byte, error) {
		called = true
		return []byte("ffmpeg"), nil
	}

	res, err := Resolve(context.Background(), video, 100, screenshot)
	require.NoError(t, err)
	require.Equal(t, "poster", res.Source)
	require.False(t, called)
	require.NotEmpty(t, res.Data)
}

func TestResolve_SameName(t *testing.T) {
	video := videoPath(t)
	dir := filepath.Dir(video)

	writeFile(t, filepath.Join(dir, "FDD-2002.jpg"), pngBytes(t, 3, 4))

	res, err := Resolve(context.Background(), video, 100, nil)
	require.NoError(t, err)
	require.Equal(t, "same-name", res.Source)
}

func TestResolve_FFmpegFallback(t *testing.T) {
	video := videoPath(t)

	screenshot := func(_ context.Context, path string, at float64) ([]byte, error) {
		require.Equal(t, video, path)
		require.InDelta(t, 20.0, at, 0.001)
		return pngBytes(t, 2, 2), nil
	}

	res, err := Resolve(context.Background(), video, 100, screenshot)
	require.NoError(t, err)
	require.Equal(t, "ffmpeg", res.Source)
	require.NotEmpty(t, res.Data)
}

func TestResolve_None(t *testing.T) {
	video := videoPath(t)

	res, err := Resolve(context.Background(), video, 100, nil)
	require.NoError(t, err)
	require.Empty(t, res.Source)
	require.Empty(t, res.Data)
}

func TestResolve_SkipsInvalidPoster(t *testing.T) {
	video := videoPath(t)
	dir := filepath.Dir(video)

	writeFile(t, filepath.Join(dir, "poster.jpg"), []byte("<html>not an image</html>"))
	writeFile(t, filepath.Join(dir, "FDD-2002.png"), pngBytes(t, 3, 4))

	res, err := Resolve(context.Background(), video, 100, nil)
	require.NoError(t, err)
	require.Equal(t, "same-name", res.Source)
}

func TestResolve_FFmpegSkippedWithoutDuration(t *testing.T) {
	video := videoPath(t)

	called := false
	screenshot := func(context.Context, string, float64) ([]byte, error) {
		called = true
		return []byte("ffmpeg"), nil
	}

	res, err := Resolve(context.Background(), video, 0, screenshot)
	require.NoError(t, err)
	require.Empty(t, res.Source)
	require.False(t, called)
}
