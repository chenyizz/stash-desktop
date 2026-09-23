package nfo

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFileSample(t *testing.T) {
	m, err := ParseFile(filepath.Join("testdata", "FDD-2002.nfo"))
	require.NoError(t, err)

	require.Equal(t, "movie", m.XMLName.Local)
	require.Equal(t, "FDD-2002", m.Num)
	require.Equal(t, "FDD-2002 ビッグ·ティッツ·ビューティー CD1", m.Title)
	require.Equal(t, "2002-06-28", m.Premiered)
	require.Equal(t, "2002", m.Year)
	require.Equal(t, "108", m.Runtime)
	require.Equal(t, "ファンタドリーム", m.Studio)
	require.Equal(t, "BigTitsBeauty", m.Series)
	require.Equal(t, "https://www.javbus.com/FDD2002", m.JavbusID)
	require.Equal(t, "FDD-2002", m.JavdbSearchID)

	require.Len(t, m.Actors, 3)
	require.Equal(t, "夕樹洋子", m.Actors[0].Name)
	require.Equal(t, "Actor", m.Actors[0].Type)

	require.Len(t, m.Sets, 4)
	require.Equal(t, "BigTitsBeauty", m.Sets[3].Name)

	require.Equal(t, []string{"FDD", "无码", "BigTitsBeauty", "ファンタドリーム"}, m.Tags)
	require.Equal(t, []string{"FDD", "无码", "BigTitsBeauty", "ファンタドリーム"}, m.Genres)
}

func TestParseStripsBOM(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "FDD-2002.nfo"))
	require.NoError(t, err)

	withBOM := append([]byte{0xEF, 0xBB, 0xBF}, data...)

	m, err := Parse(bytes.NewReader(withBOM))
	require.NoError(t, err)
	require.Equal(t, "FDD-2002", m.Num)
}

func TestParseInvalidXML(t *testing.T) {
	_, err := Parse(strings.NewReader(`<movie><title>a & b</title></movie>`))
	require.Error(t, err)
}

func TestParseWrongRoot(t *testing.T) {
	_, err := Parse(strings.NewReader(`<episodedetails><title>x</title></episodedetails>`))
	require.Error(t, err)
}

func TestParseFileMissing(t *testing.T) {
	_, err := ParseFile(filepath.Join("testdata", "does-not-exist.nfo"))
	require.Error(t, err)
}

func TestParseFileDirectory(t *testing.T) {
	_, err := ParseFile(t.TempDir())
	require.Error(t, err)
}

func TestFindForVideo(t *testing.T) {
	dir := t.TempDir()
	video := filepath.Join(dir, "FDD-2002.mp4")
	require.NoError(t, os.WriteFile(video, []byte("x"), 0o644))

	_, ok := FindForVideo(video)
	require.False(t, ok)

	nfoPath := filepath.Join(dir, "FDD-2002.nfo")
	require.NoError(t, os.WriteFile(nfoPath, []byte("<movie/>"), 0o644))

	found, ok := FindForVideo(video)
	require.True(t, ok)
	require.Equal(t, nfoPath, found)
}

func TestFindForVideoUppercaseExtension(t *testing.T) {
	dir := t.TempDir()
	video := filepath.Join(dir, "FDD-2002.mp4")
	require.NoError(t, os.WriteFile(video, []byte("x"), 0o644))

	// On case-insensitive filesystems the first candidate may match, so compare
	// case-insensitively to keep this test portable.
	nfoPath := filepath.Join(dir, "FDD-2002.NFO")
	require.NoError(t, os.WriteFile(nfoPath, []byte("<movie/>"), 0o644))

	found, ok := FindForVideo(video)
	require.True(t, ok)
	require.True(t, strings.EqualFold(found, nfoPath), "found %q, want %q", found, nfoPath)
}
