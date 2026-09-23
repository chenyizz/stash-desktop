package app

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"

	"case/backend/pkg/logger"
)

type App struct {
	app *application.App
}

func New() *App {
	return &App{}
}

func (a *App) SetApplication(wailsApp *application.App) {
	a.app = wailsApp
}

// ServiceStartup 应用启动时调用
func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	// 1. 解析目录布局
	layout, err := ResolveLayout()
	if err != nil {
		return fmt.Errorf("解析目录布局失败: %w", err)
	}

	// 2. 初始化日志
	if err := setupLogging(layout.LogDir, a.app); err != nil {
		return err
	}

	logger.Infof("Case 启动")
	logger.Infof("根目录: %s", layout.BaseDir)
	logger.Infof("数据目录: %s", layout.DataDir)
	logger.Infof("日志目录: %s", layout.LogDir)

	// 3. 加载或初始化配置
	cfg, err := setupConfig(layout)
	if err != nil {
		return fmt.Errorf("配置初始化失败: %w", err)
	}

	logger.Infof("配置文件: %s", cfg.GetConfigFile())
	logger.Infof("数据库: %s", cfg.GetDatabasePath())

	// 4. 初始化 Manager（下一步做）
	// if err := setupManager(cfg); err != nil { return err }

	return nil
}

// GetDataDir 返回数据目录（供前端调用）
func (a *App) GetDataDir() (string, error) {
	layout, err := ResolveLayout()
	if err != nil {
		return "", err
	}
	return layout.DataDir, nil
}

// ServiceShutdown 应用关闭时调用
func (a *App) ServiceShutdown() error {
	logger.Info("Case 关闭")
	return nil
}

// GetVersion 返回版本号
func (a *App) GetVersion() string {
	return "0.1.0-dev"
}

// Ping 连通性测试
func (a *App) Ping() string {
	return "pong"
}
