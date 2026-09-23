package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"

	"case/backend/pkg/logger"
)

// setupLogging 初始化日志系统，返回底层 slog.Logger 供 Manager 使用。
func setupLogging(logDir string, wailsApp *application.App) (*slog.Logger, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	logFile, err := os.OpenFile(
		filepath.Join(logDir, "case.log"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0o644,
	)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}

	fileHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug - 4, // 包含 Trace
	})
	uiHandler := logger.NewUIHandler(&wailsEmitter{app: wailsApp}, slog.LevelInfo)

	multi := logger.NewMultiHandler(fileHandler, uiHandler)
	slogLogger := slog.New(multi)

	// 同时注册到全局 logger
	logger.Logger = logger.NewSlogLogger(multi)

	return slogLogger, nil
}
