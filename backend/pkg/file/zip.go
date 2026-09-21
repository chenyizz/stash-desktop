package file

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"unicode/utf8"

	"case/backend/pkg/models"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	ErrNotReaderAt  = errors.New("invalid reader: does not implement io.ReaderAt")
	errZipFSOpenZip = errors.New("cannot open zip file inside zip file")
)

// ZipFS is a file system backed by a zip file.
type zipFS struct {
	*zip.Reader
	zipFileCloser io.Closer
	zipPath       string
}

func newZipFS(fs models.FS, path string, size int64) (*zipFS, error) {
	reader, err := fs.Open(path)
	if err != nil {
		return nil, err
	}

	asReaderAt, _ := reader.(io.ReaderAt)
	if asReaderAt == nil {
		reader.Close()
		return nil, ErrNotReaderAt
	}

	zipReader, err := zip.NewReader(asReaderAt, size)
	if err != nil {
		reader.Close()
		return nil, err
	}

	// 解码所有文件名。优先 UTF-8，其次 GBK，最后 Shift-JIS。
	// 中文 Windows 创建的 zip 通常是 GBK；日文 Windows 通常是 Shift-JIS。
	for _, f := range zipReader.File {
		f.Name = decodeZipFilename(f.Name)
	}

	return &zipFS{
		Reader:        zipReader,
		zipFileCloser: reader,
		zipPath:       path,
	}, nil
}

func (f *zipFS) rel(name string) (string, error) {
	if f.zipPath == name {
		return ".", nil
	}

	relName, err := filepath.Rel(f.zipPath, name)
	if err != nil {
		return "", fs.ErrNotExist
	}

	return filepath.ToSlash(relName), nil
}

func (f *zipFS) Stat(name string) (fs.FileInfo, error) {
	reader, err := f.Open(name)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return reader.Stat()
}

func (f *zipFS) Lstat(name string) (fs.FileInfo, error) {
	return f.Stat(name)
}

func (f *zipFS) OpenZip(name string, size int64) (models.ZipFS, error) {
	return nil, errZipFSOpenZip
}

func (f *zipFS) IsPathCaseSensitive(path string) (bool, error) {
	return true, nil
}

type zipReadDirFile struct {
	fs.File
}

func (f *zipReadDirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	asReadDirFile, _ := f.File.(fs.ReadDirFile)
	if asReadDirFile == nil {
		return nil, fmt.Errorf("internal error: not a ReadDirFile")
	}

	return asReadDirFile.ReadDir(n)
}

func (f *zipFS) Open(name string) (fs.ReadDirFile, error) {
	relName, err := f.rel(name)
	if err != nil {
		return nil, err
	}

	r, err := f.Reader.Open(relName)
	if err != nil {
		return nil, err
	}

	return &zipReadDirFile{File: r}, nil
}

func (f *zipFS) Close() error {
	return f.zipFileCloser.Close()
}

// openOnly returns a ReadCloser where calling Close will close the zip fs as well.
func (f *zipFS) OpenOnly(name string) (io.ReadCloser, error) {
	r, err := f.Open(name)
	if err != nil {
		return nil, err
	}

	return &wrappedReadCloser{
		ReadCloser: r,
		outer:      f,
	}, nil
}

type wrappedReadCloser struct {
	io.ReadCloser
	outer io.Closer
}

func (f *wrappedReadCloser) Close() error {
	_ = f.ReadCloser.Close()
	return f.outer.Close()
}

// decodeZipFilename 把 zip 里的文件名从可能的 GBK/Shift-JIS 解码成 UTF-8。
// 优先顺序：UTF-8 → GBK → Shift-JIS。
// 针对中文和日文 Windows 用户创建的 zip 文件。
func decodeZipFilename(name string) string {
	// UTF-8 有效，直接用
	if utf8.ValidString(name) {
		return name
	}

	// 尝试 GBK（简体中文 Windows 的默认编码）
	if result, _, err := transform.String(simplifiedchinese.GBK.NewDecoder(), name); err == nil {
		if utf8.ValidString(result) {
			return result
		}
	}

	// 尝试 Shift-JIS（日文 Windows 的默认编码）
	if result, _, err := transform.String(japanese.ShiftJIS.NewDecoder(), name); err == nil {
		if utf8.ValidString(result) {
			return result
		}
	}

	// 都失败，原样返回
	return name
}
