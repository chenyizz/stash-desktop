package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"

	"case/backend/pkg/logger"
)

// setupLogging 初始化日志系统。
// 可以在启动早期调用（用临时目录），也可以在配置加载后重新调用（用正式目录）。
func setupLogging(logDir string, wailsApp *application.App) error {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	logFile, err := os.OpenFile(
		filepath.Join(logDir, "case.log"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0o644,
	)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	fileHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug - 4, // 包含 Trace
	})
	uiHandler := logger.NewUIHandler(&wailsEmitter{app: wailsApp}, slog.LevelInfo)

	logger.Logger = logger.NewSlogLogger(
		logger.NewMultiHandler(fileHandler, uiHandler),
	)
	return nil
}
