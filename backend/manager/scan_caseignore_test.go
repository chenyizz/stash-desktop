//go:build integration
// +build integration

package manager

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"case/backend/pkg/file"

	// Necessary to register custom migrations.
	_ "case/backend/pkg/sqlite/migrations"
)

// caseIgnorePathFilter wraps CaseIgnoreFilter to implement PathFilter for testing.
// It provides a fixed library root for the filter.
type caseIgnorePathFilter struct {
	filter      *file.CaseIgnoreFilter
	libraryRoot string
}

func (f *caseIgnorePathFilter) Accept(ctx context.Context, path string, info fs.FileInfo, zipFilePath string) bool {
	return f.filter.Accept(ctx, path, info, f.libraryRoot, zipFilePath)
}

// createTestFileOnDisk creates a file with some content.
func createTestFileOnDisk(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create directory for %s: %v", path, err)
	}
	// Write some content so the file has a non-zero size.
	if err := os.WriteFile(path, []byte("test content for "+name), 0644); err != nil {
		t.Fatalf("failed to create file %s: %v", path, err)
	}
	return path
}

// createCaseIgnoreFile creates a .caseignore file with the given content.
func createCaseIgnoreFile(t *testing.T, dir, content string) {
	t.Helper()
	path := filepath.Join(dir, ".caseignore")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create .caseignore: %v", err)
	}
}

func TestScannerWithCaseIgnore(t *testing.T) {
	// Create temp directory structure.
	tmpDir := t.TempDir()

	// Create test files.
	createTestFileOnDisk(t, tmpDir, "video1.mp4")
	createTestFileOnDisk(t, tmpDir, "video2.mp4")
	createTestFileOnDisk(t, tmpDir, "ignore_me.mp4")
	createTestFileOnDisk(t, tmpDir, "subdir/video3.mp4")
	createTestFileOnDisk(t, tmpDir, "subdir/skip_this.mp4")
	createTestFileOnDisk(t, tmpDir, "excluded_dir/video4.mp4")
	createTestFileOnDisk(t, tmpDir, "temp/processing.mp4")

	// Create .caseignore file.
	caseignore := `# Ignore specific files
ignore_me.mp4
subdir/skip_this.mp4

# Ignore directories
excluded_dir/
temp/
`
	createCaseIgnoreFile(t, tmpDir, caseignore)

	// Create caseignore filter with library root.
	caseIgnoreFilter := &caseIgnorePathFilter{
		filter:      file.NewCaseIgnoreFilter(),
		libraryRoot: tmpDir,
	}

	// Create scanner.
	scanner := &file.Scanner{
		ScanFilters: []file.PathFilter{caseIgnoreFilter},
	}

	testScenarios := []struct {
		path     string
		accepted bool
	}{
		{filepath.Join(tmpDir, "video1.mp4"), true},
		{filepath.Join(tmpDir, "video2.mp4"), true},
		{filepath.Join(tmpDir, "ignore_me.mp4"), false},
		{filepath.Join(tmpDir, "subdir/video3.mp4"), true},
		{filepath.Join(tmpDir, "subdir/skip_this.mp4"), false},
		{filepath.Join(tmpDir, "excluded_dir/video4.mp4"), false},
		{filepath.Join(tmpDir, "temp/processing.mp4"), false},
	}

	ctx := context.Background()

	for _, scenario := range testScenarios {
		info, err := os.Stat(scenario.path)
		if err != nil {
			t.Fatalf("failed to stat file %s: %v", scenario.path, err)
		}
		accepted := scanner.AcceptEntry(ctx, scenario.path, info, "")

		if accepted != scenario.accepted {
			t.Errorf("unexpected accept result for %s: expected %v, got %v",
				scenario.path, scenario.accepted, accepted)
		}
	}
}

func TestScannerWithNestedCaseIgnore(t *testing.T) {
	// Create temp directory structure.
	tmpDir := t.TempDir()

	// Create test files.
	createTestFileOnDisk(t, tmpDir, "root.mp4")
	createTestFileOnDisk(t, tmpDir, "root.tmp")
	createTestFileOnDisk(t, tmpDir, "subdir/sub.mp4")
	createTestFileOnDisk(t, tmpDir, "subdir/sub.log")
	createTestFileOnDisk(t, tmpDir, "subdir/sub.tmp")

	// Root .caseignore excludes *.tmp.
	createCaseIgnoreFile(t, tmpDir, "*.tmp\n")

	// Subdir .caseignore excludes *.log.
	createCaseIgnoreFile(t, filepath.Join(tmpDir, "subdir"), "*.log\n")

	// Create caseignore filter with library root.
	caseIgnoreFilter := &caseIgnorePathFilter{
		filter:      file.NewCaseIgnoreFilter(),
		libraryRoot: tmpDir,
	}

	// Create scanner.
	scanner := &file.Scanner{
		ScanFilters: []file.PathFilter{caseIgnoreFilter},
	}

	testScenarios := []struct {
		path     string
		accepted bool
	}{
		{filepath.Join(tmpDir, "root.mp4"), true},
		{filepath.Join(tmpDir, "root.tmp"), false},
		{filepath.Join(tmpDir, "subdir/sub.mp4"), true},
		{filepath.Join(tmpDir, "subdir/sub.log"), false},
		{filepath.Join(tmpDir, "subdir/sub.tmp"), false},
	}

	ctx := context.Background()

	for _, scenario := range testScenarios {
		info, err := os.Stat(scenario.path)
		if err != nil {
			t.Fatalf("failed to stat file %s: %v", scenario.path, err)
		}
		accepted := scanner.AcceptEntry(ctx, scenario.path, info, "")

		if accepted != scenario.accepted {
			t.Errorf("unexpected accept result for %s: expected %v, got %v",
				scenario.path, scenario.accepted, accepted)
		}
	}
}

func TestScannerWithoutCaseIgnore(t *testing.T) {
	// Create temp directory structure (no .caseignore).
	tmpDir := t.TempDir()

	// Create test files.
	createTestFileOnDisk(t, tmpDir, "video1.mp4")
	createTestFileOnDisk(t, tmpDir, "video2.mp4")
	createTestFileOnDisk(t, tmpDir, "subdir/video3.mp4")

	// Create caseignore filter with library root (but no .caseignore file exists).
	caseIgnoreFilter := &caseIgnorePathFilter{
		filter:      file.NewCaseIgnoreFilter(),
		libraryRoot: tmpDir,
	}

	// Create scanner.
	scanner := &file.Scanner{
		ScanFilters: []file.PathFilter{caseIgnoreFilter},
	}

	testScenarios := []struct {
		path     string
		accepted bool
	}{
		{filepath.Join(tmpDir, "video1.mp4"), true},
		{filepath.Join(tmpDir, "video2.mp4"), true},
		{filepath.Join(tmpDir, "subdir/video3.mp4"), true},
	}

	ctx := context.Background()

	for _, scenario := range testScenarios {
		info, err := os.Stat(scenario.path)
		if err != nil {
			t.Fatalf("failed to stat file %s: %v", scenario.path, err)
		}
		accepted := scanner.AcceptEntry(ctx, scenario.path, info, "")

		if accepted != scenario.accepted {
			t.Errorf("unexpected accept result for %s: expected %v, got %v",
				scenario.path, scenario.accepted, accepted)
		}
	}
}

func TestScannerWithNegationPattern(t *testing.T) {
	// Create temp directory structure.
	tmpDir := t.TempDir()

	// Create test files.
	createTestFileOnDisk(t, tmpDir, "file1.tmp")
	createTestFileOnDisk(t, tmpDir, "file2.tmp")
	createTestFileOnDisk(t, tmpDir, "keep_this.tmp")
	createTestFileOnDisk(t, tmpDir, "video.mp4")

	// Create .caseignore with negation.
	caseignore := `*.tmp
!keep_this.tmp
`
	createCaseIgnoreFile(t, tmpDir, caseignore)

	// Create caseignore filter with library root.
	caseIgnoreFilter := &caseIgnorePathFilter{
		filter:      file.NewCaseIgnoreFilter(),
		libraryRoot: tmpDir,
	}

	// Create scanner.
	scanner := &file.Scanner{
		ScanFilters: []file.PathFilter{caseIgnoreFilter},
	}

	testScenarios := []struct {
		path     string
		accepted bool
	}{
		{filepath.Join(tmpDir, "file1.tmp"), false},
		{filepath.Join(tmpDir, "file2.tmp"), false},
		{filepath.Join(tmpDir, "keep_this.tmp"), true},
		{filepath.Join(tmpDir, "video.mp4"), true},
	}

	ctx := context.Background()

	for _, scenario := range testScenarios {
		info, err := os.Stat(scenario.path)
		if err != nil {
			t.Fatalf("failed to stat file %s: %v", scenario.path, err)
		}
		accepted := scanner.AcceptEntry(ctx, scenario.path, info, "")

		if accepted != scenario.accepted {
			t.Errorf("unexpected accept result for %s: expected %v, got %v",
				scenario.path, scenario.accepted, accepted)
		}
	}
}
