package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"

	"case/backend/pkg/logger"
)

type App struct {
	app *application.App
}

func (a *App) SetApplication(wailsApp *application.App) {
	a.app = wailsApp // ← 必须有这个方法
}

func New() *App {
	return &App{}
}

// ServiceStartup 应用启动时调用
func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	// 1. 创建数据目录
	dataDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("获取用户配置目录失败: %w", err)
	}
	caseDir := filepath.Join(dataDir, "case")
	os.MkdirAll(caseDir, 0o755)

	// 2. 打开日志文件
	logFilePath := filepath.Join(caseDir, "case.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 3. 文件 Handler
	fileHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug - 4, // 包含 Trace
	})

	// 4. UI Handler
	uiHandler := logger.NewUIHandler(a.app, slog.LevelInfo)

	// 5. Multi Handler
	multi := logger.NewMultiHandler(fileHandler, uiHandler)

	// 6. 注册全局 logger
	logger.Logger = logger.NewSlogLogger(multi)

	// 7. 测试
	logger.Infof("Case 启动，日志文件：%s", logFilePath)
	logger.Debugf("这是一条 debug 日志，应该只在文件里")
	logger.Infof("这是一条 info 日志，应该在文件和 UI 里都出现")

	return nil
}

// ServiceShutdown 应用关闭时调用
func (a *App) ServiceShutdown() error {
	return nil
}

// GetVersion 返回版本号
func (a *App) GetVersion() string {
	return "0.1.0-dev"
}

// GetDataDir 返回数据目录
func (a *App) GetDataDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}
	dataDir := filepath.Join(configDir, "case")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create data dir: %w", err)
	}
	return dataDir, nil
}

// Ping 连通性测试
func (a *App) Ping() string {
	return "pong"
}
