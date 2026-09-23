package cover

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
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

// TestResolve_FFmpegRealFrame is an end-to-end check of the ffmpeg source using
// the real ffmpeg binary. It is skipped when ffmpeg is not on PATH.
func TestResolve_FFmpegRealFrame(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available on PATH")
	}

	dir := t.TempDir()
	video := filepath.Join(dir, "clip.mp4")

	gen := exec.Command("ffmpeg", "-y",
		"-f", "lavfi", "-i", "testsrc=size=320x240:rate=10",
		"-t", "2", "-pix_fmt", "yuv420p", video)
	if out, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("generating test video: %v\n%s", err, out)
	}

	screenshot := func(_ context.Context, path string, at float64) ([]byte, error) {
		frame := filepath.Join(dir, "frame.jpg")
		cmd := exec.Command("ffmpeg", "-y", "-ss", fmt.Sprintf("%.3f", at),
			"-i", path, "-frames:v", "1", frame)
		if out, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("%v: %s", err, out)
		}
		return os.ReadFile(frame)
	}

	res, err := Resolve(context.Background(), video, 2, screenshot)
	require.NoError(t, err)
	require.Equal(t, "ffmpeg", res.Source)
	require.NotEmpty(t, res.Data)

	_, _, err = image.DecodeConfig(bytes.NewReader(res.Data))
	require.NoError(t, err)
}
