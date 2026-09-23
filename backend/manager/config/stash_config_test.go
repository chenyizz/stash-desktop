package config

import (
	"path/filepath"
	"testing"
)

func TestLibraryConfigsGetLibraryFromDirPathReturnsMostSpecificPath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "library")
	nested := filepath.Join(root, "images")

	parent := &LibraryConfig{
		Path:         root,
		ExcludeImage: true,
	}
	child := &LibraryConfig{
		Path:         nested,
		ExcludeVideo: true,
	}

	libs := LibraryConfigs{parent, child}

	got := libs.GetLibraryFromDirPath(filepath.Join(nested, "set"))
	if got != child {
		t.Fatalf("expected nested library config, got %#v", got)
	}
}

func TestLibraryConfigsGetLibraryFromDirPathReturnsMostSpecificPathRegardlessOfOrder(t *testing.T) {
	root := filepath.Join(t.TempDir(), "library")
	nested := filepath.Join(root, "images")

	parent := &LibraryConfig{
		Path:         root,
		ExcludeImage: true,
	}
	child := &LibraryConfig{
		Path:         nested,
		ExcludeVideo: true,
	}

	libs := LibraryConfigs{child, parent}

	got := libs.GetLibraryFromDirPath(filepath.Join(nested, "set"))
	if got != child {
		t.Fatalf("expected nested library config, got %#v", got)
	}
}

func TestLibraryConfigsGetLibraryRootFromDirPathReturnsTopmostPath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "library")
	nested := filepath.Join(root, "images")

	libs := LibraryConfigs{
		{Path: root},
		{Path: nested},
	}

	got := libs.GetLibraryRootFromDirPath(filepath.Join(nested, "set"))
	if got != root {
		t.Fatalf("expected topmost library path %q, got %q", root, got)
	}
}

func TestLibraryConfigsGetLibraryRootFromDirPathReturnsTopmostPathRegardlessOfOrder(t *testing.T) {
	root := filepath.Join(t.TempDir(), "library")
	nested := filepath.Join(root, "images")

	libs := LibraryConfigs{
		{Path: nested},
		{Path: root},
	}

	got := libs.GetLibraryRootFromDirPath(filepath.Join(nested, "set"))
	if got != root {
		t.Fatalf("expected topmost library path %q, got %q", root, got)
	}
}

func TestLibraryConfigsGetLibraryFromPathReturnsMostSpecificPath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "library")
	nested := filepath.Join(root, "images")

	parent := &LibraryConfig{
		Path:         root,
		ExcludeImage: true,
	}
	child := &LibraryConfig{
		Path:         nested,
		ExcludeVideo: true,
	}

	libs := LibraryConfigs{parent, child}

	got := libs.GetLibraryFromPath(filepath.Join(nested, "image.jpg"))
	if got != child {
		t.Fatalf("expected nested library config, got %#v", got)
	}
}

func TestLibraryConfigsGetLibraryFromDirPathReturnsNilOutsideLibraries(t *testing.T) {
	root := filepath.Join(t.TempDir(), "library")
	outside := filepath.Join(t.TempDir(), "outside")

	libs := LibraryConfigs{{Path: root}}

	got := libs.GetLibraryFromDirPath(outside)
	if got != nil {
		t.Fatalf("expected nil library config, got %#v", got)
	}
}

func TestLibraryConfigsGetLibraryRootFromDirPathReturnsEmptyOutsideLibraries(t *testing.T) {
	root := filepath.Join(t.TempDir(), "library")
	outside := filepath.Join(t.TempDir(), "outside")

	libs := LibraryConfigs{{Path: root}}

	got := libs.GetLibraryRootFromDirPath(outside)
	if got != "" {
		t.Fatalf("expected empty library path, got %q", got)
	}
}
