package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type App struct {
	app *application.App
}

func New() *App {
	return &App{}
}

// ServiceStartup 应用启动时调用
func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
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
