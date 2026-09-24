package app

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestResizeJPEG_ScalesWidth(t *testing.T) {
	out, err := resizeJPEG(pngBytes(t, 400, 200), 100)
	require.NoError(t, err)

	cfg, _, err := image.DecodeConfig(bytes.NewReader(out))
	require.NoError(t, err)
	assert.Equal(t, 100, cfg.Width)
	assert.NotEmpty(t, out)
}

func TestResizeJPEG_InvalidData(t *testing.T) {
	_, err := resizeJPEG([]byte("not an image"), 100)
	require.Error(t, err)
}
