package app

import (
	"fmt"
	"os"
	"path/filepath"

	"case/backend/manager/config"
	"case/backend/pkg/logger"
)

// setupConfig 加载或初始化配置。返回值 cfg 一定是可用的。
func setupConfig(layout *Layout) (*config.Config, error) {
	configFile := filepath.Join(layout.DataDir, "config.yml")
	// 只写新名；旧名 STASH_CONFIG_FILE 仅作为兼容读的回退
	os.Setenv("CASE_CONFIG_FILE", configFile)

	cfg, err := config.Initialize()
	if err != nil {
		return nil, fmt.Errorf("配置初始化失败: %w", err)
	}

	if cfg.IsNewSystem() {
		if err := initializeNewConfig(cfg, layout); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// initializeNewConfig 首次运行时创建默认配置。
func initializeNewConfig(cfg *config.Config, layout *Layout) error {
	dataDir := layout.DataDir

	cfg.SetConfigFile(filepath.Join(dataDir, "config.yml"))
	cfg.SetString(config.DataDir, dataDir)
	cfg.SetString(config.Database, filepath.Join(dataDir, "case.db"))
	cfg.SetString(config.Generated, filepath.Join(dataDir, "generated"))
	cfg.SetString(config.Cache, filepath.Join(dataDir, "cache"))
	cfg.SetString(config.BlobsPath, filepath.Join(dataDir, "blobs"))
	cfg.SetString(config.Metadata, filepath.Join(dataDir, "metadata"))

	// ★ Blob 存文件系统，避免数据库膨胀
	cfg.SetInterface(config.BlobsStorage, config.BlobStorageTypeFilesystem)

	if err := cfg.SetInitialConfig(); err != nil {
		return fmt.Errorf("设置初始配置失败: %w", err)
	}

	// 创建子目录
	for _, sub := range []string{"generated", "cache", "blobs", "metadata"} {
		if err := os.MkdirAll(filepath.Join(dataDir, sub), 0o755); err != nil {
			return fmt.Errorf("创建 %s 目录失败: %w", sub, err)
		}
	}

	if err := cfg.Write(); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	cfg.FinalizeSetup()

	logger.Infof("已创建默认配置: %s", cfg.GetConfigFile())
	return nil
}
