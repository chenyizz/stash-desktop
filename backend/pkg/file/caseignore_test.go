package file

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// Helper to create an empty file.
func createTestFile(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create file %s: %v", path, err)
	}
}

// Helper to create a file with content.
func createTestFileWithContent(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file %s: %v", path, err)
	}
}

// Helper to create a directory.
func createTestDir(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

// walkAndFilter walks the directory tree and returns paths accepted by the filter.
// Returns paths relative to root for easier assertion.
func walkAndFilter(t *testing.T, root string, filter *CaseIgnoreFilter) []string {
	t.Helper()
	var accepted []string
	ctx := context.Background()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself.
		if path == root {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if filter.Accept(ctx, path, info, root, "") {
			relPath, _ := filepath.Rel(root, path)
			accepted = append(accepted, filepath.ToSlash(relPath))
		} else if info.IsDir() {
			// If directory is rejected, skip it.
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}

	sort.Strings(accepted)
	return accepted
}

// assertPathsEqual checks that the accepted paths match expected.
func assertPathsEqual(t *testing.T, expected, actual []string) {
	t.Helper()
	sort.Strings(expected)

	if len(expected) != len(actual) {
		t.Errorf("path count mismatch:\nexpected %d: %v\nactual %d: %v", len(expected), expected, len(actual), actual)
		return
	}

	for i := range expected {
		if expected[i] != actual[i] {
			t.Errorf("path mismatch at index %d:\nexpected: %s\nactual: %s", i, expected[i], actual[i])
		}
	}
}

func TestCaseIgnore_ExactFilename(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files.
	createTestFile(t, tmpDir, "video1.mp4")
	createTestFile(t, tmpDir, "video2.mp4")
	createTestFile(t, tmpDir, "ignore_me.mp4")

	// Create .caseignore that excludes exact filename.
	createTestFileWithContent(t, tmpDir, ".caseignore", "ignore_me.mp4\n")

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	expected := []string{
		".caseignore",
		"video1.mp4",
		"video2.mp4",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_WildcardPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files.
	createTestFile(t, tmpDir, "video1.mp4")
	createTestFile(t, tmpDir, "video2.mp4")
	createTestFile(t, tmpDir, "temp1.tmp")
	createTestFile(t, tmpDir, "temp2.tmp")
	createTestFile(t, tmpDir, "notes.log")

	// Create .caseignore that excludes by extension.
	createTestFileWithContent(t, tmpDir, ".caseignore", "*.tmp\n*.log\n")

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	expected := []string{
		".caseignore",
		"video1.mp4",
		"video2.mp4",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_DirectoryExclusion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files.
	createTestFile(t, tmpDir, "video1.mp4")
	createTestDir(t, tmpDir, "excluded_dir")
	createTestFile(t, tmpDir, "excluded_dir/video2.mp4")
	createTestFile(t, tmpDir, "excluded_dir/video3.mp4")
	createTestDir(t, tmpDir, "included_dir")
	createTestFile(t, tmpDir, "included_dir/video4.mp4")

	// Create .caseignore that excludes a directory.
	createTestFileWithContent(t, tmpDir, ".caseignore", "excluded_dir/\n")

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	expected := []string{
		".caseignore",
		"included_dir",
		"included_dir/video4.mp4",
		"video1.mp4",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_NegationPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files.
	createTestFile(t, tmpDir, "file1.tmp")
	createTestFile(t, tmpDir, "file2.tmp")
	createTestFile(t, tmpDir, "keep_this.tmp")

	// Create .caseignore that excludes *.tmp but keeps one.
	createTestFileWithContent(t, tmpDir, ".caseignore", "*.tmp\n!keep_this.tmp\n")

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	expected := []string{
		".caseignore",
		"keep_this.tmp",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_CommentsAndEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files.
	createTestFile(t, tmpDir, "video1.mp4")
	createTestFile(t, tmpDir, "ignore_me.mp4")

	// Create .caseignore with comments and empty lines.
	caseignore := `# This is a comment
ignore_me.mp4

# Another comment

`
	createTestFileWithContent(t, tmpDir, ".caseignore", caseignore)

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	expected := []string{
		".caseignore",
		"video1.mp4",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_NestedCaseIgnoreFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files.
	createTestFile(t, tmpDir, "root_video.mp4")
	createTestFile(t, tmpDir, "root_ignore.tmp")
	createTestDir(t, tmpDir, "subdir")
	createTestFile(t, tmpDir, "subdir/sub_video.mp4")
	createTestFile(t, tmpDir, "subdir/sub_ignore.log")
	createTestFile(t, tmpDir, "subdir/also_tmp.tmp")

	// Root .caseignore excludes *.tmp.
	createTestFileWithContent(t, tmpDir, ".caseignore", "*.tmp\n")

	// Subdir .caseignore excludes *.log.
	createTestFileWithContent(t, tmpDir, "subdir/.caseignore", "*.log\n")

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	// *.tmp from root should apply everywhere.
	// *.log from subdir should only apply in subdir.
	expected := []string{
		".caseignore",
		"root_video.mp4",
		"subdir",
		"subdir/.caseignore",
		"subdir/sub_video.mp4",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_PathPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files.
	createTestFile(t, tmpDir, "video1.mp4")
	createTestDir(t, tmpDir, "subdir")
	createTestFile(t, tmpDir, "subdir/video2.mp4")
	createTestFile(t, tmpDir, "subdir/skip_this.mp4")

	// Create .caseignore that excludes a specific path.
	createTestFileWithContent(t, tmpDir, ".caseignore", "subdir/skip_this.mp4\n")

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	expected := []string{
		".caseignore",
		"subdir",
		"subdir/video2.mp4",
		"video1.mp4",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_DoubleStarPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files.
	createTestFile(t, tmpDir, "video1.mp4")
	createTestDir(t, tmpDir, "a")
	createTestFile(t, tmpDir, "a/video2.mp4")
	createTestDir(t, tmpDir, "a/temp")
	createTestFile(t, tmpDir, "a/temp/video3.mp4")
	createTestDir(t, tmpDir, "a/b")
	createTestDir(t, tmpDir, "a/b/temp")
	createTestFile(t, tmpDir, "a/b/temp/video4.mp4")

	// Create .caseignore that excludes temp directories at any level.
	createTestFileWithContent(t, tmpDir, ".caseignore", "**/temp/\n")

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	expected := []string{
		".caseignore",
		"a",
		"a/b",
		"a/video2.mp4",
		"video1.mp4",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_LeadingSlashPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files.
	createTestFile(t, tmpDir, "ignore.mp4")
	createTestDir(t, tmpDir, "subdir")
	createTestFile(t, tmpDir, "subdir/ignore.mp4")

	// Create .caseignore that excludes only at root level.
	createTestFileWithContent(t, tmpDir, ".caseignore", "/ignore.mp4\n")

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	// Only root ignore.mp4 should be excluded.
	expected := []string{
		".caseignore",
		"subdir",
		"subdir/ignore.mp4",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_NoCaseIgnoreFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files without any .caseignore.
	createTestFile(t, tmpDir, "video1.mp4")
	createTestFile(t, tmpDir, "video2.mp4")
	createTestDir(t, tmpDir, "subdir")
	createTestFile(t, tmpDir, "subdir/video3.mp4")

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	// All files should be accepted.
	expected := []string{
		"subdir",
		"subdir/video3.mp4",
		"video1.mp4",
		"video2.mp4",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_HiddenDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files including hidden directory.
	createTestFile(t, tmpDir, "video1.mp4")
	createTestDir(t, tmpDir, ".hidden")
	createTestFile(t, tmpDir, ".hidden/video2.mp4")

	// Create .caseignore that excludes hidden directories.
	createTestFileWithContent(t, tmpDir, ".caseignore", ".*\n!.caseignore\n")

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	expected := []string{
		".caseignore",
		"video1.mp4",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_MultiplePatternsSameLine(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files.
	createTestFile(t, tmpDir, "video1.mp4")
	createTestFile(t, tmpDir, "file.tmp")
	createTestFile(t, tmpDir, "file.log")
	createTestFile(t, tmpDir, "file.bak")

	// Each pattern should be on its own line.
	createTestFileWithContent(t, tmpDir, ".caseignore", "*.tmp\n*.log\n*.bak\n")

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	expected := []string{
		".caseignore",
		"video1.mp4",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_TrailingSpaces(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files.
	createTestFile(t, tmpDir, "video1.mp4")
	createTestFile(t, tmpDir, "ignore_me.mp4")

	// Pattern with trailing spaces (should be trimmed).
	createTestFileWithContent(t, tmpDir, ".caseignore", "ignore_me.mp4   \n")

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	expected := []string{
		".caseignore",
		"video1.mp4",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_EscapedHash(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files.
	createTestFile(t, tmpDir, "video1.mp4")
	createTestFile(t, tmpDir, "#filename.mp4")

	// Escaped hash should match literal # character.
	createTestFileWithContent(t, tmpDir, ".caseignore", "\\#filename.mp4\n")

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	expected := []string{
		".caseignore",
		"video1.mp4",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_CaseSensitiveMatching(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files - use distinct names that work on all filesystems.
	createTestFile(t, tmpDir, "video_lower.mp4")
	createTestFile(t, tmpDir, "VIDEO_UPPER.mp4")
	createTestFile(t, tmpDir, "other.avi")

	// Pattern should match exactly (case-sensitive).
	createTestFileWithContent(t, tmpDir, ".caseignore", "video_lower.mp4\n")

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	// Only exact match is excluded.
	expected := []string{
		".caseignore",
		"VIDEO_UPPER.mp4",
		"other.avi",
	}

	assertPathsEqual(t, expected, accepted)
}

func TestCaseIgnore_ComplexScenario(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a complex directory structure.
	createTestFile(t, tmpDir, "video1.mp4")
	createTestFile(t, tmpDir, "video2.avi")
	createTestFile(t, tmpDir, "thumbnail.jpg")
	createTestFile(t, tmpDir, "metadata.nfo")
	createTestDir(t, tmpDir, "movies")
	createTestFile(t, tmpDir, "movies/movie1.mp4")
	createTestFile(t, tmpDir, "movies/movie1.nfo")
	createTestDir(t, tmpDir, "movies/.thumbnails")
	createTestFile(t, tmpDir, "movies/.thumbnails/thumb1.jpg")
	createTestDir(t, tmpDir, "temp")
	createTestFile(t, tmpDir, "temp/processing.mp4")
	createTestDir(t, tmpDir, "backup")
	createTestFile(t, tmpDir, "backup/video1.mp4.bak")

	// Complex .caseignore.
	caseignore := `# Ignore metadata files
*.nfo

# Ignore hidden directories
.*
!.caseignore

# Ignore temp and backup directories
temp/
backup/

# But keep thumbnails in specific location
!movies/.thumbnails/
`
	createTestFileWithContent(t, tmpDir, ".caseignore", caseignore)

	filter := NewCaseIgnoreFilter()
	accepted := walkAndFilter(t, tmpDir, filter)

	expected := []string{
		".caseignore",
		"movies",
		"movies/.thumbnails",
		"movies/.thumbnails/thumb1.jpg",
		"movies/movie1.mp4",
		"thumbnail.jpg",
		"video1.mp4",
		"video2.avi",
	}

	assertPathsEqual(t, expected, accepted)
}
