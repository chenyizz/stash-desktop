package nfo

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MaxFileSize is the maximum NFO file size accepted by ParseFile. NFO files are
// tiny in practice; this guards against accidentally reading a huge or
// malicious file.
const MaxFileSize = 10 << 20 // 10 MiB

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Parse decodes an NFO document from r. The reader is expected to be UTF-8
// (with or without a leading BOM). A non-<movie> root element is an error.
//
// The decoder runs in strict mode and only understands the XML predeclared
// entities plus the common HTML entities. DTD/doctype entities are never
// expanded by encoding/xml, so NFO parsing is not vulnerable to XXE or
// "billion laughs" entity expansion.
func Parse(r io.Reader) (*Movie, error) {
	br := bufio.NewReader(r)
	if err := skipBOM(br); err != nil {
		return nil, err
	}

	dec := xml.NewDecoder(br)
	dec.Strict = true
	dec.Entity = xml.HTMLEntity

	var m Movie
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("decode nfo: %w", err)
	}

	return &m, nil
}

// ParseFile reads and parses the NFO file at path.
func ParseFile(path string) (*Movie, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat nfo %q: %w", path, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("nfo path %q is a directory", path)
	}
	if info.Size() > MaxFileSize {
		return nil, fmt.Errorf("nfo file %q too large: %d bytes (max %d)", path, info.Size(), MaxFileSize)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open nfo %q: %w", path, err)
	}
	defer f.Close()

	m, err := Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %w", path, err)
	}

	return m, nil
}

// FindForVideo returns the path of the NFO sidecar that shares the video's
// basename (e.g. "FDD-2002.mp4" -> "FDD-2002.nfo"), and whether it exists.
func FindForVideo(videoPath string) (string, bool) {
	base := strings.TrimSuffix(videoPath, filepath.Ext(videoPath))

	for _, candidate := range []string{base + ".nfo", base + ".NFO"} {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, true
		}
	}

	return "", false
}

func skipBOM(br *bufio.Reader) error {
	head, err := br.Peek(len(utf8BOM))
	if err != nil {
		// A file shorter than the BOM cannot start with one.
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	if bytes.Equal(head, utf8BOM) {
		_, _ = br.Discard(len(utf8BOM))
	}

	return nil
}
