package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wailsapp/wails/v3/pkg/application"

	"case/backend/manager"
	"case/backend/manager/config"
	"case/backend/pkg/logger"
	"case/backend/pkg/models"
)

type App struct {
	app    *application.App
	cfg    *config.Config
	mgr    *manager.Manager
	logger *slog.Logger

	cancel context.CancelFunc
}

type SceneDTO struct {
	ID         int            `json:"id"`
	Title      string         `json:"title"`
	Path       string         `json:"path"`
	OSHash     string         `json:"oshash"`
	Checksum   string         `json:"checksum"`
	Organized  bool           `json:"organized"`
	Tags       []TagDTO       `json:"tags"`
	Performers []PerformerDTO `json:"performers"`
	CreatedAt  string         `json:"createdAt"`
	UpdatedAt  string         `json:"updatedAt"`
}

// ScenesPageDTO 是场景列表的分页结果。
type ScenesPageDTO struct {
	Scenes   []SceneDTO `json:"scenes"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

// normalizePage 约束分页参数，避免非法/过大请求。
func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	switch {
	case pageSize < 1:
		pageSize = 50
	case pageSize > 200:
		pageSize = 200
	}

	return page, pageSize
}

func intSetKeys(set map[int]struct{}) []int {
	keys := make([]int, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}

	return keys
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
	slogLogger, err := setupLogging(layout.LogDir, a.app)
	if err != nil {
		return err
	}
	a.logger = slogLogger

	logger.Infof("Case 启动")
	logger.Infof("根目录: %s", layout.BaseDir)
	logger.Infof("数据目录: %s", layout.DataDir)
	logger.Infof("日志目录: %s", layout.LogDir)

	// 3. 加载或初始化配置
	//    注意：setupConfig 会把新系统也完成初始化，
	//    所以之后 cfg.IsNewSystem() 会返回 false
	cfg, err := setupConfig(layout)
	if err != nil {
		return fmt.Errorf("配置初始化失败: %w", err)
	}
	a.cfg = cfg

	logger.Infof("配置文件: %s", cfg.GetConfigFile())
	logger.Infof("数据库: %s", cfg.GetDatabasePath())

	// 4. 初始化 Manager
	//    此时 cfg.IsNewSystem() 已经是 false，
	//    所以 Initialize 内部会自动调用 postInit：
	//    打开 SQLite、初始化 FFmpeg、创建 StreamManager 等
	mgr, err := manager.Initialize(cfg, slogLogger)
	if err != nil {
		logger.Errorf("Manager 初始化失败: %v", err)
		return fmt.Errorf("Manager 初始化失败: %w", err)
	}
	a.mgr = mgr

	logger.Info("Manager 初始化完成")

	// 订阅扫描完成信号并推送到前端（事件契约：scan:complete + {at}）
	watchCtx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	go watchScanEvents(watchCtx, mgr, (&wailsEmitter{app: a.app}).Emit)
	go a.sweepThumbnails(watchCtx)

	return nil
}

// ServiceShutdown 应用关闭时调用
func (a *App) ServiceShutdown() error {
	logger.Info("Case 关闭")
	if a.cancel != nil {
		a.cancel()
	}
	if a.mgr != nil {
		a.mgr.Shutdown()
	}
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

// GetVersion 返回版本号
func (a *App) GetVersion() string {
	return "0.1.0-dev"
}

// Ping 连通性测试
func (a *App) Ping() string {
	return "pong"
}

// GetSystemStatus 简单状态查询（验证 Manager 是否活着）
func (a *App) GetSystemStatus() map[string]any {
	if a.cfg == nil {
		return map[string]any{"status": "config_not_ready"}
	}

	status := map[string]any{
		"isNewSystem": a.cfg.IsNewSystem(),
		"configFile":  a.cfg.GetConfigFile(),
		"database":    a.cfg.GetDatabasePath(),
	}

	if a.mgr != nil {
		status["manager"] = "ready"
	} else {
		status["manager"] = "not_initialized"
	}

	return status
}

// ScanLibrary 触发扫描指定文件夹。返回 Job ID，前端可通过事件监听进度。
func (a *App) ScanLibrary(path string) (int, error) {
	if a.mgr == nil {
		return 0, fmt.Errorf("manager 未初始化")
	}

	input := manager.ScanMetadataInput{
		Paths:  []string{path},
		Rescan: false,
	}

	jobID, err := a.mgr.Scan(context.Background(), input)
	if err != nil {
		logger.Errorf("扫描失败: %v", err)
		return 0, fmt.Errorf("扫描失败: %w", err)
	}

	logger.Infof("扫描任务已启动，Job ID: %d, 路径: %s", jobID, path)
	return jobID, nil
}

// FindScenes 分页查询场景列表，返回分页结果与总数。
func (a *App) FindScenes(page int, pageSize int) (*ScenesPageDTO, error) {
	if a.mgr == nil {
		return nil, fmt.Errorf("manager 未初始化")
	}

	page, pageSize = normalizePage(page, pageSize)

	var dtos []SceneDTO
	total := 0

	err := a.mgr.Repository.WithReadTxn(context.Background(), func(ctx context.Context) error {
		findFilter := &models.FindFilterType{
			Page:    &page,
			PerPage: &pageSize,
		}

		result, err := a.mgr.Repository.Scene.Query(ctx, models.SceneQueryOptions{
			QueryOptions: models.QueryOptions{
				FindFilter: findFilter,
				Count:      true,
			},
		})
		if err != nil {
			return err
		}
		total = result.Count

		scenes, err := result.Resolve(ctx)
		if err != nil {
			return err
		}

		// 先收集本页关联 ID，再一次性批量取名称（避免逐行查询名称）
		tagIDSet := make(map[int]struct{})
		performerIDSet := make(map[int]struct{})
		for _, s := range scenes {
			if err := s.LoadTagIDs(ctx, a.mgr.Repository.Scene); err != nil {
				return fmt.Errorf("加载标签关联失败: %w", err)
			}
			if err := s.LoadPerformerIDs(ctx, a.mgr.Repository.Scene); err != nil {
				return fmt.Errorf("加载演员关联失败: %w", err)
			}
			for _, id := range s.TagIDs.List() {
				tagIDSet[id] = struct{}{}
			}
			for _, id := range s.PerformerIDs.List() {
				performerIDSet[id] = struct{}{}
			}
		}

		tags, err := a.findTags(ctx, intSetKeys(tagIDSet))
		if err != nil {
			return err
		}
		performers, err := a.findPerformers(ctx, intSetKeys(performerIDSet))
		if err != nil {
			return err
		}

		for _, s := range scenes {
			dtos = append(dtos, SceneDTO{
				ID:         s.ID,
				Title:      s.GetTitle(),
				Path:       s.Path,
				OSHash:     s.OSHash,
				Checksum:   s.Checksum,
				Organized:  s.Organized,
				Tags:       toTagDTOs(s.TagIDs.List(), tags),
				Performers: toPerformerDTOs(s.PerformerIDs.List(), performers),
				CreatedAt:  s.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt:  s.UpdatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("查询场景失败: %w", err)
	}

	logger.Infof("查询场景: page=%d, pageSize=%d, total=%d, 返回 %d 条", page, pageSize, total, len(dtos))

	return &ScenesPageDTO{
		Scenes:   dtos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
