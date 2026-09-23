package app

import (
	"fmt"
	"os"
	"path/filepath"
)

// Layout 描述应用的目录布局
type Layout struct {
	BaseDir string // 根目录
	DataDir string // 数据目录（config.yml、case.db 等）
	LogDir  string // 日志目录
}

// ResolveLayout 确定应用的目录布局。
//
// 策略：
//  1. 环境变量 CASE_DATA_DIR（用户强制指定）
//  2. exe 同目录（便携模式，首选）
//  3. %LOCALAPPDATA%\case（降级，exe 目录不可写时）
//
// 返回的目录保证存在且可写。
func ResolveLayout() (*Layout, error) {
	// ① 环境变量强制指定
	if dir := os.Getenv("CASE_DATA_DIR"); dir != "" {
		if err := ensureWritable(dir); err != nil {
			return nil, fmt.Errorf("CASE_DATA_DIR 不可写 (%s): %w", dir, err)
		}
		return buildLayout(dir)
	}

	// ② exe 同目录
	if exeDir := ExeDir(); exeDir != "" {
		if err := ensureWritable(exeDir); err == nil {
			return buildLayout(exeDir)
		}
	}

	// ③ 降级到 %LOCALAPPDATA%\case
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		if dir, err := os.UserCacheDir(); err == nil {
			localAppData = dir
		}
	}
	if localAppData == "" {
		return nil, fmt.Errorf("无法确定用户数据目录")
	}

	dir := filepath.Join(localAppData, "case")
	if err := ensureWritable(dir); err != nil {
		return nil, fmt.Errorf("用户数据目录不可写: %w", err)
	}
	return buildLayout(dir)
}

// ExeDir 返回 exe 所在目录（解析符号链接）。
func ExeDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exePath)
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}
	return dir
}

// buildLayout 从根目录构造布局，并创建所有需要的子目录。
func buildLayout(baseDir string) (*Layout, error) {
	layout := &Layout{
		BaseDir: baseDir,
		DataDir: filepath.Join(baseDir, "data"),
		LogDir:  filepath.Join(baseDir, "log"),
	}

	// ★ 关键：创建两个子目录，否则后面 config.Initialize() 会失败
	for _, dir := range []string{layout.DataDir, layout.LogDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("创建目录 %s 失败: %w", dir, err)
		}
	}

	return layout, nil
}

// ensureWritable 确保目录存在且可写。
// 通过真实写文件测试，而不是依赖 stat 权限位。
func ensureWritable(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	testFile := filepath.Join(dir, ".case_write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return err
	}
	f.Close()
	return os.Remove(testFile)
}
