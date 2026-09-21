package manager

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"case/backend/manager/config"
	"case/backend/pkg/ffmpeg"
	"case/backend/pkg/fsutil"
	"case/backend/pkg/gallery"
	"case/backend/pkg/group"
	"case/backend/pkg/image"
	"case/backend/pkg/job"
	"case/backend/pkg/logger"
	"case/backend/pkg/models/paths"
	"case/backend/pkg/plugin"
	"case/backend/pkg/scene"
	"case/backend/pkg/scraper"
	"case/backend/pkg/session"
	"case/backend/pkg/sqlite"
	"case/backend/pkg/utils"

	"github.com/remeh/sizedwaitgroup"
)

// Called at startup
func Initialize(cfg *config.Config, l *slog.Logger) (*Manager, error) {
	ctx := context.TODO()

	db := sqlite.NewDatabase()
	repo := db.Repository()

	// start with empty paths
	mgrPaths := &paths.Paths{}

	scraperRepository := scraper.NewRepository(repo)
	scraperCache := scraper.NewCache(cfg, scraperRepository)

	pluginCache := plugin.NewCache(cfg)

	sceneService := &scene.Service{
		File:             db.File,
		Repository:       db.Scene,
		MarkerRepository: db.SceneMarker,
		PluginCache:      pluginCache,
		Paths:            mgrPaths,
		Config:           cfg,
	}

	imageService := &image.Service{
		File:       db.File,
		Repository: db.Image,
	}

	galleryService := &gallery.Service{
		Repository:   db.Gallery,
		ImageFinder:  db.Image,
		ImageService: imageService,
		File:         db.File,
		Folder:       db.Folder,
	}

	groupService := &group.Service{
		Repository: db.Group,
	}

	mgr := &Manager{
		Config: cfg,
		Logger: l,

		Paths: mgrPaths,

		ImageThumbnailGenerateWaitGroup: sizedwaitgroup.New(1),

		JobManager:      initJobManager(cfg),
		ReadLockManager: fsutil.NewReadLockManager(),

		DownloadStore: NewDownloadStore(),

		PluginCache:  pluginCache,
		ScraperCache: scraperCache,

		// DLNAService 已移除 - Wails 不需要 DLNA

		Database:   db,
		Repository: repo,

		SceneService:   sceneService,
		ImageService:   imageService,
		GalleryService: galleryService,
		GroupService:   groupService,

		scanSubs: &subscriptionManager{},
	}

	if !cfg.IsNewSystem() {
		logger.Infof("using config file: %s", cfg.GetConfigFile())

		err := cfg.Validate()
		if err != nil {
			return nil, fmt.Errorf("invalid configuration: %w", err)
		}

		if err := mgr.postInit(ctx); err != nil {
			return nil, err
		}
	} else {
		cfgFile := cfg.GetConfigFile()
		if cfgFile != "" {
			cfgFile += " "
		}

		// create temporary session store - this will be re-initialised
		// after config is complete
		mgr.SessionStore = session.NewStore(cfg)

		logger.Warnf("config file %snot found. Assuming new system...", cfgFile)
	}

	instance = mgr
	return mgr, nil
}

func formatDuration(t time.Duration) string {
	switch {
	case t >= time.Minute:
		t = t.Round(time.Second)
	case t >= time.Second:
		t = t.Round(10 * time.Millisecond)
	default:
		t = t.Round(time.Millisecond)
	}

	return t.String()
}

func initJobManager(cfg *config.Config) *job.Manager {
	ret := job.NewManager()

	// 桌面通知：Wails 下暂不实现
	ctx := context.Background()
	c := ret.Subscribe(context.Background())
	go func() {
		for {
			select {
			case j := <-c.RemovedJob:
				_ = j // 暂时忽略
			case <-ctx.Done():
				return
			}
		}
	}()

	return ret
}

// postInit initialises the paths, caches and database after the initial
// configuration has been set. Should only be called if the configuration
// is valid.
func (s *Manager) postInit(ctx context.Context) error {
	s.RefreshConfig()

	s.SessionStore = session.NewStore(s.Config)
	s.PluginCache.RegisterSessionStore(s.SessionStore)

	s.RefreshPluginCache()
	s.RefreshPluginSourceManager()

	s.RefreshScraperCache()
	s.RefreshScraperSourceManager()

	// DLNA 已移除

	s.SetBlobStoreOptions()

	// writeStashIcon 已移除 - 由 Wails 处理图标

	// clear the downloads and tmp directories
	if s.Config.GetGeneratedPath() != "" {
		const deleteTimeout = 1 * time.Second

		utils.Timeout(func() {
			if err := fsutil.EmptyDir(s.Paths.Generated.Downloads); err != nil {
				logger.Warnf("could not empty downloads directory: %v", err)
			}
			if err := fsutil.EnsureDir(s.Paths.Generated.Tmp); err != nil {
				logger.Warnf("could not create temporary directory: %v", err)
			} else {
				if err := fsutil.EmptyDir(s.Paths.Generated.Tmp); err != nil {
					logger.Warnf("could not empty temporary directory: %v", err)
				}
			}
		}, deleteTimeout, func(done chan struct{}) {
			logger.Info("Please wait. Deleting temporary files...")
			<-done
			logger.Info("Temporary files deleted.")
		})
	}

	if err := s.Database.Open(s.Config.GetDatabasePath()); err != nil {
		var migrationNeededErr *sqlite.MigrationNeededError
		if errors.As(err, &migrationNeededErr) {
			logger.Warn(err)
		} else {
			return err
		}
	}

	// Set the proxy if defined in config
	if s.Config.GetProxy() != "" {
		os.Setenv("HTTP_PROXY", s.Config.GetProxy())
		os.Setenv("HTTPS_PROXY", s.Config.GetProxy())
		os.Setenv("NO_PROXY", s.Config.GetNoProxy())
		logger.Info("Using HTTP proxy")
	}

	s.RefreshFFMpeg(ctx)
	s.RefreshStreamManager()

	return nil
}

// writeStashIcon 已删除 - 用 Wails 的图标系统替代

func (s *Manager) RefreshFFMpeg(ctx context.Context) {
	configDirectory := s.Config.GetConfigPathAbs()
	stashHomeDir := paths.GetStashHomeDirectory()

	ffmpegPath := s.Config.GetFFMpegPath()
	ffprobePath := s.Config.GetFFProbePath()

	if ffmpegPath != "" {
		if err := ffmpeg.ValidateFFMpeg(ffmpegPath); err != nil {
			logger.Errorf("invalid ffmpeg path: %v", err)
			return
		}
		if err := ffmpeg.ValidateFFMpegCodecSupport(ffmpegPath); err != nil {
			logger.Warn(err)
		}
	} else {
		ffmpegPath = ffmpeg.ResolveFFMpeg(configDirectory, stashHomeDir)
	}

	if ffprobePath != "" {
		if err := ffmpeg.ValidateFFProbe(ffmpegPath); err != nil {
			logger.Errorf("invalid ffprobe path: %v", err)
			return
		}
	} else {
		ffprobePath = ffmpeg.ResolveFFProbe(configDirectory, stashHomeDir)
	}

	if ffmpegPath == "" {
		logger.Warn("Couldn't find FFmpeg")
	}
	if ffprobePath == "" {
		logger.Warn("Couldn't find FFprobe")
	}

	if ffmpegPath != "" && ffprobePath != "" {
		logger.Debugf("using ffmpeg: %s", ffmpegPath)
		logger.Debugf("using ffprobe: %s", ffprobePath)

		s.FFMpeg = ffmpeg.NewEncoder(ffmpegPath)
		s.FFProbe = ffmpeg.NewFFProbe(ffprobePath)

		s.FFMpeg.InitHWSupport(context.Background())
	}
}
