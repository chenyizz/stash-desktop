# Case - 项目结构说明文档

> 本项目是从 [stashapp/stash](https://github.com/stashapp/stash)（一个自托管的媒体资源管理器）迁移重构的 **Wails v3 + Svelte** 桌面原生应用。
>
> 原项目 stash 是基于 Go 后端 (GraphQL API + HTTP Server) + React 前端的 Web 应用。迁移后，后端核心业务逻辑保留，替换为 Wails 框架直接绑定 Go 与 Svelte 前端，不再依赖 HTTP Server / GraphQL API 层。

---

## 目录树总览

```
Case/
├── main.go                          # Wails 应用入口
├── go.mod / go.sum                  # Go 模块依赖
├── Taskfile.yml                     # Task 构建脚本
├── README.md
├── PROJECT_STRUCTURE.md             # 本文档
│
├── internal/                        # Wails 应用层（适配层）
│   └── app/
│       ├── app.go                   # 应用服务（启动/关闭/版本/日志初始化）
│       └── wails_emitter.go         # Wails 事件发射器（日志→UI）
│
├── frontend/                        # Svelte 前端
│   ├── src/
│   │   ├── App.svelte               # 根组件（当前为脚手架模板）
│   │   ├── main.ts                  # 前端入口
│   │   └── vite-env.d.ts           # Vite 类型声明
│   ├── public/                      # 静态资源（图标、字体、背景图）
│   ├── package.json                 # 前端依赖
│   ├── vite.config.ts               # Vite 构建配置
│   ├── svelte.config.js             # Svelte 配置
│   └── tsconfig.json                # TypeScript 配置
│
├── backend/                         # 后端核心业务逻辑
│   ├── manager/                     # 全局管理器（协调中心）
│   │   ├── config/                  # 配置管理
│   │   ├── task/                    # 初始化/迁移任务
│   │   ├── manager.go               # Manager 主结构体
│   │   ├── init.go                  # 初始化逻辑
│   │   ├── manager_tasks.go         # 扫描/导入/导出/生成/清理任务
│   │   ├── models.go                # 管理器数据模型
│   │   ├── repository.go            # 服务接口定义
│   │   ├── task.go / task_scan.go   # 各类任务实现
│   │   ├── task_generate*.go        # 生成任务（截图/预览/缩略图等）
│   │   ├── task_clean.go            # 清理任务
│   │   ├── task_export.go           # 导出任务
│   │   ├── task_import.go           # 导入任务
│   │   ├── task_migrate_hash.go     # Hash迁移任务
│   │   ├── task_optimise.go         # 数据库优化任务
│   │   ├── task_plugin.go           # 插件任务
│   │   ├── task_transcode.go        # 转码任务
│   │   └── ...                      # 其他辅助文件
│   │
│   └── pkg/                         # 核心功能模块包
│       ├── sqlite/                  # SQLite 数据库层
│       ├── models/                  # 数据模型定义
│       ├── scene/                   # 场景（视频）服务
│       ├── image/                   # 图片服务
│       ├── gallery/                 # 图库服务
│       ├── group/                   # 分组服务
│       ├── performer/               # 演员服务
│       ├── studio/                  # 工作室服务
│       ├── tag/                     # 标签服务
│       ├── ffmpeg/                  # FFmpeg/FFprobe 集成
│       ├── file/                    # 文件系统操作
│       ├── scraper/                 # 数据抓取器
│       ├── plugin/                  # 插件系统
│       ├── job/                     # 任务队列管理
│       ├── logger/                  # 日志系统
│       ├── hash/                    # 哈希计算
│       ├── fsutil/                  # 文件系统工具
│       ├── utils/                   # 通用工具
│       ├── match/                   # 路径匹配
│       ├── session/                 # 会话管理
│       ├── pkg/                     # 包管理器
│       ├── exec/                    # 外部命令执行
│       ├── javascript/              # JavaScript VM (goja)
│       ├── txn/                     # 事务管理
│       ├── sliceutil/               # 切片工具集
│       └── python/                  # Python 环境检测
│
└── build/                           # 构建配置
    ├── config.yml                   # Wails 构建配置
    ├── Taskfile.yml                 # 公共构建任务
    ├── appicon.png / appicon.icon/  # 应用图标
    ├── windows/                     # Windows 打包
    ├── linux/                       # Linux 打包
    ├── darwin/                      # macOS 打包
    └── docker/                      # Docker 配置
```

---

## 一、入口层

### `main.go` - 应用入口

**作用**：Wails v3 应用的主入口，负责创建并启动 Wails 应用实例。

**主要功能**：
- 通过 `//go:embed` 将构建后的前端产物 (`frontend/dist`) 嵌入到 Go 二进制中
- 创建 `application.New()` 实例，配置应用名称 "case"、嵌入的资源文件服务
- 将 `internal/app.App` 注册为 Wails Service（`application.NewService(appService)`）
- 创建主窗口（1280x800 分辨率），URL 指向 `/`
- 调用 `wailsApp.Run()` 进入 Wails 事件循环

**与原项目 stash 的对应关系**：
- 原项目：`main.go` 启动 HTTP Server 监听端口，注册 GraphQL 路由和 UI 路由
- 迁移后：不再启动 HTTP Server，Wails 框架直接管理窗口和前端的双向绑定

---

### `internal/app/` - Wails 适配层

#### `app.go` - 应用服务

**作用**：Wails 框架与后端核心逻辑的桥梁。实现 `application.Service` 接口。

**主要功能**：
1. **ServiceStartup**：应用启动回调
   - 创建用户数据目录 (`$APPDATA/case` 或 `$XDG_CONFIG_HOME/case`)
   - 初始化日志文件 (`case.log`)
   - 创建文件日志 Handler（JSON 格式，Trace 级别）
   - 创建 UI 日志 Handler（Info 级别，通过 Wails Events 推送到前端）
   - 通过 MultiHandler 合并两种 Handler，注册为全局 Logger
2. **ServiceShutdown**：应用关闭回调
3. **GetVersion**：返回应用版本号 (`0.1.0-dev`)
4. **GetDataDir**：返回数据目录路径
5. **Ping**：连通性测试（返回 "pong"）

**与原项目关系**：
- 原项目通过 HTTP Handler 提供服务，前端通过 GraphQL 请求获取数据
- 迁移后，`App` 的方法通过 Wails 的 bindings 机制自动暴露给前端调用

#### `wails_emitter.go` - 事件发射器

**作用**：将 logger 包的日志事件通过 Wails 的 Event 系统推送到前端 UI。

**工作原理**：
- 实现 `logger.EventEmitter` 接口
- `Emit(name, data...)` → `app.Event.Emit(name, data...)`
- 前端可以通过 Wails 的 `Event.on()` 监听日志事件，实现日志在 UI 中实时展示

---

## 二、前端层 (`frontend/`)

**技术栈**：Svelte 5 + TypeScript + Vite 8 + Zustand + TanStack Query

**当前状态**：前端目前处于项目脚手架阶段，仅包含基本的 App.svelte 模板，显示版本、Ping 结果和数据目录。

**关键依赖**：

| 依赖 | 用途 |
|------|------|
| `svelte` | UI 框架 |
| `vite` | 构建工具 |
| `@wailsio/runtime` | Wails 前端运行时（Go 方法绑定、事件系统） |
| `zustand` | 轻量级状态管理 |
| `@tanstack/svelte-query` | 数据获取与缓存 |

**Wails Bindings 机制**：
- 构建时 Wails 自动扫描 `internal/app` 中导出的方法，生成 TypeScript bindings 文件到 `frontend/bindings/`
- 前端通过 `import { App } from "../bindings/case/internal/app"` 直接调用 Go 方法

---

## 三、管理器层 (`backend/manager/`)

Manager 是整个应用的中枢，整合了所有服务和子系统。

### `manager.go` - 管理器主结构

**Manager 结构体包含的核心字段**：

| 字段 | 类型 | 职责 |
|------|------|------|
| `Config` | `*config.Config` | 全局配置管理 |
| `Logger` | `*slog.Logger` | 日志记录器 |
| `Paths` | `*paths.Paths` | 生成文件路径管理 |
| `FFMpeg` / `FFProbe` | `*ffmpeg.*` | 音视频工具绑定 |
| `StreamManager` | `*ffmpeg.StreamManager` | 流媒体管理 |
| `JobManager` | `*job.Manager` | 任务队列管理器 |
| `ReadLockManager` | `*fsutil.ReadLockManager` | 文件读锁管理 |
| `DownloadStore` | `*DownloadStore` | 下载状态存储 |
| `SessionStore` | `*session.Store` | 会话状态存储 |
| `PluginCache` | `*plugin.Cache` | 插件缓存 |
| `ScraperCache` | `*scraper.Cache` | 抓取器缓存 |
| `PluginPackageManager` | `*pkg.Manager` | 插件包管理器 |
| `ScraperPackageManager` | `*pkg.Manager` | 抓取器包管理器 |
| `Database` | `*sqlite.Database` | SQLite 数据库 |
| `Repository` | `models.Repository` | 数据仓库接口 |
| `SceneService` / `ImageService` / `GalleryService` / `GroupService` | 对应接口 | 各实体类型的业务服务 |

**主要方法**：

| 方法 | 功能 |
|------|------|
| `GetInstance()` | 获取全局单例 |
| `Setup(ctx, input)` | 初始系统配置（创建数据库、目录、写入配置） |
| `RefreshConfig()` | 刷新配置并创建所有生成目录 |
| `RefreshFFMpeg(ctx)` | 检测/刷新 FFmpeg 和 FFProbe 路径 |
| `RefreshStreamManager()` | 重建流媒体管理器 |
| `RefreshPluginCache()` | 重新加载插件 |
| `RefreshScraperCache()` | 重新加载抓取器 |
| `GetSystemStatus()` | 获取系统状态（数据库版本、配置路径等） |
| `Shutdown()` | 优雅关闭所有子系统 |
| `AnonymiseDatabase()` | 匿名化数据库 |

### `init.go` - 初始化

**作用**：Manager 的工厂函数，在应用启动时执行完整的依赖注入和初始化。

**初始化流程**：
1. 创建 SQLite 数据库实例和 Repository
2. 初始化 Scene/Image/Gallery/Group 各 Service
3. 组装 Manager 结构体
4. 如果是新系统 → 标记等待配置
5. 如果是已有系统 → 执行 `postInit()`

**postInit 流程**：
1. `RefreshConfig()` — 刷新配置和生成目录
2. 创建 SessionStore
3. 加载插件和抓取器缓存
4. 设置 Blob 存储选项
5. 打开数据库连接（自动执行迁移）
6. 设置 HTTP 代理（如果配置了）
7. 初始化 FFmpeg/FFProbe
8. 创建 StreamManager
9. 清理临时目录

### `manager_tasks.go` - 核心任务调度

**作用**：将用户操作（扫描、导入、导出、生成、清理等）封装为 Job 并推入 JobManager 队列。

**提供的方法**：

| 方法 | 功能 |
|------|------|
| `Scan(ctx, input)` | 扫描媒体目录，发现新文件并提取元数据 |
| `Import(ctx)` | 从 JSON 元数据导入 |
| `Export(ctx)` | 导出所有元数据为 JSON |
| `Generate(ctx, input)` | 生成预览/截图/Sprite/缩略图等 |
| `GenerateDefaultScreenshot(ctx, sceneId)` | 为指定场景生成默认截图 |
| `GenerateScreenshot(ctx, sceneId, at)` | 在指定时间点生成截图 |
| `Clean(ctx, input)` | 清理数据库中不存在于磁盘的文件记录 |
| `OptimiseDatabase(ctx)` | 优化数据库（VACUUM 等） |
| `MigrateHash(ctx)` | 迁移所有场景的哈希命名格式 |
| `RunSingleTask(ctx, t)` | 执行单次任务 |

### `config/` - 配置管理

**作用**：管理应用的全局配置，对应原 stash 项目中的 `config.yml` 配置系统。

**关键文件**：

| 文件 | 功能 |
|------|------|
| `config.go` | 配置主文件：定义所有配置常量、Config 结构体、配置读写 |
| `stash_config.go` | 定义 Stash 配置（扫描路径及排除规则） |
| `enums.go` | 配置相关枚举类型（BlobStorageType 等） |
| `init.go` | 配置初始化 |
| `tasks.go` | 配置相关的默认任务（如 FFmpeg 下载等） |
| `ui.go` | UI 相关配置项 |

**核心配置键（部分）**：

| 配置键 | 用途 |
|--------|------|
| `stash` | 媒体库路径列表 |
| `database` | SQLite 数据库文件路径 |
| `generated` | 生成文件（截图、预览等）的输出目录 |
| `cache` | 缓存目录 |
| `blobs_path` | Blob（封面图等）存储路径 |
| `ffmpeg_path` / `ffprobe_path` | FFmpeg/FFProbe 可执行文件路径 |
| `parallel_tasks` | 并行任务数 |
| `scrapers_path` | 抓取器配置路径 |
| `plugins_path` | 插件路径 |
| `max_transcode_size` | 转码最大尺寸限制 |

### `task/` - 辅助任务

| 文件 | 功能 |
|------|------|
| `clean_generated.go` | 清理生成的支持文件和临时缓存 |
| `download_ffmpeg.go` | 下载 FFmpeg/FFprobe 工具链 |
| `migrate.go` | Blob 迁移任务（filesystem ↔ database） |
| `migrate_blobs.go` | Blob 存储迁移辅助 |
| `migrate_scene_screenshots.go` | 场景截图文件重写（使用新文件命名路径） |
| `packages.go` | 可选包下载与安装（FFmpeg 等） |

---

## 四、数据层

### `backend/pkg/models/` — 数据模型定义

**作用**：定义项目中所有实体类型、关联关系、查询结构、JSON Schema 等。

#### 核心实体模型

| 文件 | 实体 | 关键字段 |
|------|------|----------|
| `model_scene.go` | Scene（场景/视频） | Title, Code, Director, Date, Rating, Files, Tags, Performers, Studio, Galleries, URLs, StashIDs |
| `model_image.go` | Image（图片） | Title, Rating, Files, Tags, Performers, Studio, Galleries, URLs |
| `model_gallery.go` | Gallery（图库） | Title, Date, Rating, Tags, Performers, Studio, URLs |
| `model_performer.go` | Performer（演员） | Name, Gender, Birthdate, Ethnicity, Height, Measurements, Aliases, Tags |
| `model_studio.go` | Studio（工作室） | Name, URL, ParentStudio, Rating, Aliases, Tags |
| `model_tag.go` | Tag（标签） | Name, Description, Aliases, ParentTags, Children |
| `model_group.go` | Group（分组） | Name, Description, SubGroups, FrontImage, BackImage |
| `model_file.go` | File（文件） | BaseName, Size, ModTime, Fingerprints |
| `model_folder.go` | Folder（文件夹） | DirEntry 信息 |
| `model_saved_filter.go` | SavedFilter（已保存的筛选器） | 筛选条件持久化 |
| `model_scene_marker.go` | SceneMarker（场景标记） | SceneID, Time, Title, Tags |
| `model_gallery_chapter.go` | GalleryChapter（图库章节） | GalleryID, ImageID, Title, ImageIndex |
| `model_scraped_item.go` | 抓取结果模型 | 从外部源抓取的数据格式 |

#### 辅助子目录

| 子目录 | 功能 |
|--------|------|
| `json/` | JSON 时间类型序列化 |
| `jsonschema/` | 各实体的 JSON Schema 导入/导出格式定义 |
| `mocks/` | 各实体 ReaderWriter 接口的 Mock 实现（用于测试） |
| `paths/` | 生成文件路径的构建逻辑（截图/Sprite/预览/标记片段的命名规则） |

#### 其他关键文件

| 文件 | 功能 |
|------|------|
| `repository.go` | **Repository 接口总定义**：包含文件、文件夹、场景、图片、图库、分组、演员、工作室、标签、场景标记、已保存筛选器、Blob、图库章节等所有实体的 CRUD 和查询接口 |
| `query.go` | 通用查询选项和结果类型 |
| `filter.go` / `find_filter.go` | 筛选条件的通用结构（FindFilter、Sort、分页等） |
| `filename_parser.go` | 文件名解析器（从文件名中提取元数据） |
| `fingerprint.go` | 文件指纹（哈希）类型定义 |
| `custom_fields.go` | 自定义字段系统 |
| `relationships.go` | 关联关系映射（多对多关联 ID 类型） |
| `resolution.go` | 视频分辨率枚举 |
| `orientation.go` | 图片方向枚举 |
| `rating.go` | 评分系统（1-100 制） |
| `date.go` | 日期类型（含精度：年/月/日） |
| `value.go` | Optional 类型（OptionalString, OptionalInt 等），用于部分更新 |

### `backend/pkg/sqlite/` — SQLite 数据库层

**作用**：项目的数据持久化层，基于 SQLite + [sqlx](https://github.com/jmoiron/sqlx) 实现。

#### 核心结构

| 文件 | 职责 |
|------|------|
| `database.go` | **Database 主结构体**：管理读写连接（readDB / writeDB），提供 Open/Close/Migrate/Optimise 等操作。支持读写分离（1 写连接 + 10 读连接）。 |
| `repository.go` | **通用 Repository 基类**：提供 `getAll`、`destroy`、`exists`、`runCountQuery`、`runIdsQuery`、`executeFindQuery` 等通用数据库操作方法 |
| `driver.go` | SQLite 驱动注册和连接配置 |
| `tx.go` | 事务包装器（dbWrapper），使 sqlite 包内的所有操作通过事务上下文执行 |
| `transaction.go` | 事务管理器 |
| `migrate.go` | 数据库迁移框架 |
| `anonymise.go` | 数据库匿名化（移除个人信息用于分享） |

#### 实体 Store 实现

每个实体类型都有对应的一对文件：`{entity}.go`（CRUD 实现）+ `{entity}_filter.go`（筛选器）：

| Store 文件 | 对应实体 |
|------------|----------|
| `scene.go` + `scene_filter.go` | Scene 场景存储 |
| `image.go` + `image_filter.go` | Image 图片存储 |
| `gallery.go` + `gallery_filter.go` | Gallery 图库存储 |
| `performer.go` + `performer_filter.go` | Performer 演员存储 |
| `studio.go` + `studio_filter.go` | Studio 工作室存储 |
| `tag.go` + `tag_filter.go` | Tag 标签存储 |
| `group.go` + `group_filter.go` | Group 分组存储 |
| `file.go` + `file_filter.go` | File 文件存储 |
| `folder.go` + `folder_filter.go` | Folder 文件夹存储 |
| `scene_marker.go` + `scene_marker_filter.go` | SceneMarker 场景标记存储 |
| `saved_filter.go` | SavedFilter 已保存筛选器存储 |
| `gallery_chapter.go` | GalleryChapter 图库章节存储 |

#### 关键辅助文件

| 文件 | 功能 |
|------|------|
| `filter.go` | 通用筛选器构建逻辑 |
| `filter_hierarchical.go` | 层级筛选器（如标签父子级） |
| `criterion_handlers.go` | 筛选条件的处理函数映射 |
| `relationships.go` | 多对多关联（join 表操作） |
| `group_relationships.go` | 分组关联操作 |
| `custom_fields.go` | 自定义字段的数据库操作 |
| `blob.go` + `blob_migrate.go` | Blob 存储和迁移（filesystem ↔ database） |
| `fingerprint.go` | 文件指纹/哈希存储操作 |
| `phash.go` | 感知哈希存储 |
| `batch.go` | 批量操作 |
| `functions.go` | 自定义 SQLite 函数 |
| `tables.go` / `table.go` | 数据库表定义和 DDL |
| `date.go` | 日期转换工具 |
| `history.go` | 变更历史记录 |
| `query.go` | 通用查询构建器 |
| `timestamp.go` | 时间戳处理 |
| `values.go` | 值转换工具 |

#### `migrations/` — 数据库迁移

**作用**：版本化数据库模式迁移，从 v1 到 v86，涵盖 stash 项目数年的数据库演进。

**迁移文件格式**：
- `NN_description.up.sql` — 向前迁移的 SQL
- `NN_postmigrate.go` — SQL 迁移后执行的 Go 代码（数据修复/迁移）
- `NN_premigrate.go` — SQL 迁移前执行的 Go 代码（数据准备）

**关键迁移里程碑**：
- `1_initial.up.sql` — 初始数据库模式（基础表结构）
- `32_files.up.sql` + `32_postmigrate.go` — 采用抽象文件概念（将 files 与 folders 分离）
- `45_blobs.up.sql` + `45_postmigrate.go` — Blob 存储系统
- `49_saved_filter_refactor.up.sql` — 已保存筛选器重构
- `72_tag_sort_name.up.sql` — 标签排序名
- `84_migrate.go` — 文件夹数据修正

#### `blob/` — Blob 文件系统存储

`blob/fs.go` — Blob 的文件系统存储实现（小文件按哈希路径存储在磁盘上）

### `backend/pkg/models/paths/` — 生成文件路径

**作用**：管理所有生成文件的路径结构。对应原 stash 项目中的 `generated/` 目录。

| 文件 | 功能 |
|------|------|
| `paths.go` | 定义 Generated 路径结构：Screenshots, Vtt, Markers, Transcodes, Downloads, InteractiveHeatmap |
| `paths_scenes.go` | 场景相关文件的路径：截图、预览视频、预览图片、Sprite、转码文件、VTT 字幕、热力图 |
| `paths_scene_markers.go` | 场景标记预览的路径生成 |
| `paths_json.go` | JSON 元数据的文件路径 |
| `paths_generated.go` | 生成文件路径的辅助方法 |

---

## 五、业务服务层

各实体类型的业务服务，封装了该实体的创建、查询、更新、删除、导入、导出等完整业务逻辑。

### `backend/pkg/scene/` — 场景服务

| 文件 | 功能 |
|------|------|
| `service.go` | Scene Service 结构体定义 |
| `create.go` | 创建场景（含文件关联） |
| `update.go` | 更新场景（部分更新 + 关联关系变更） |
| `query.go` | 场景查询 |
| `find.go` | 场景查找 |
| `filter.go` | 场景筛选器 |
| `delete.go` | 删除场景（级联删除标记、生成文件等） |
| `scan.go` | 场景扫描（发现新视频文件） |
| `import.go` | 从 JSON 导入场景 |
| `export.go` | 导出场景为 JSON |
| `hash.go` | 场景哈希管理 |
| `fingerprints.go` | 场景文件指纹管理 |
| `merge.go` | 合并多个场景 |
| `migrate_hash.go` | 场景文件哈希迁移 |
| `migrate_screenshots.go` | 场景截图文件迁移 |
| `filename_parser.go` | 从文件名解析场景元数据 |
| `marker_import.go` / `marker_query.go` | 场景标记的导入和查询 |
| `generate/` | 场景生成子包 |
| ├── `generator.go` | 生成器入口（协调截图/预览/Sprite 的创建） |
| ├── `screenshot.go` | 截图生成 |
| ├── `preview.go` | 视频预览生成 |
| ├── `sprite.go` | Sprite 图生成（视频缩略图网格） |
| ├── `transcode.go` | 视频转码 |
| └── `marker_preview.go` | 标记片段预览生成 |

### `backend/pkg/image/` — 图片服务

| 文件 | 功能 |
|------|------|
| `service.go` | Image Service 结构体 |
| `query.go` | 图片查询 |
| `filter.go` | 图片筛选器 |
| `delete.go` | 删除图片 |
| `scan.go` | 图片扫描（发现新图片文件） |
| `import.go` / `export.go` | 图片导入/导出 |
| `update.go` | 图片更新 |
| `thumbnail.go` | 缩略图生成 |
| `webp.go` | WebP 格式支持 |
| `vips.go` | 使用 libvips 库处理图片 |

### `backend/pkg/gallery/` — 图库服务

| 文件 | 功能 |
|------|------|
| `service.go` | Gallery Service 结构体 |
| `query.go` | 图库查询 |
| `filter.go` | 图库筛选器 |
| `delete.go` | 删除图库 |
| `scan.go` | 图库扫描（从文件夹/ZIP 发现图库） |
| `import.go` / `export.go` | 图库导入/导出 |
| `update.go` | 图库更新 |
| `chapter_import.go` | 图库章节导入 |
| `validation.go` | 图库数据验证（如图片中是否有图库级联约束） |

### `backend/pkg/group/` — 分组服务

| 文件 | 功能 |
|------|------|
| `service.go` | Group Service 结构体 |
| `create.go` | 创建分组 |
| `update.go` | 更新分组 |
| `query.go` | 分组查询 |
| `import.go` / `export.go` | 分组导入/导出 |
| `reorder.go` | 子分组重新排序 |
| `validate.go` | 数据验证 |

### `backend/pkg/performer/` — 演员服务

| 文件 | 功能 |
|------|------|
| `import.go` / `export.go` | 演员数据导入/导出 |
| `query.go` | 演员查询 |
| `validate.go` | 数据验证 |
| `url.go` | URL 处理 |
| `doc.go` | 包文档 |

### `backend/pkg/studio/` / `backend/pkg/tag/` — 工作室/标签服务

结构类似 performer：
- `validate.go` — 数据验证（名称唯一性等）
- `query.go` — 查询（按名称、父子级等）
- `import.go` / `export.go` — JSON 导入/导出

---

## 六、工具与基础设施层

### `backend/pkg/ffmpeg/` — FFmpeg/FFprobe 集成

**作用**：封装 FFmpeg 和 FFProbe 的命令行调用，提供音视频处理能力。

| 文件 | 功能 |
|------|------|
| `ffmpeg.go` | FFmpeg 封装：路径解析、版本检测、可执行文件验证 |
| `ffprobe.go` | FFprobe 封装：调用 ffprobe 提取媒体文件的元数据（分辨率、时长、编码等） |
| `codec.go` | 编码器相关逻辑 |
| `codec_hardware.go` | 硬件加速编码器支持 |
| `container.go` | 容器格式处理 |
| `format.go` | 格式定义 |
| `filter.go` | FFmpeg 滤镜参数构建 |
| `frame_rate.go` | 帧率处理 |
| `generate.go` | 生成命令构建 |
| `options.go` | 生成选项定义 |
| `types.go` | 类型定义 |
| `media_detection.go` | 媒体文件格式检测 |
| `browser.go` | FFmpeg 可执行文件的系统搜索 |
| `downloader.go` | FFmpeg 二进制下载功能 |

#### `stream/` — 流媒体

| 文件 | 功能 |
|------|------|
| `stream.go` | 流媒体播放（HLS 实时转码流） |
| `stream_segmented.go` | 分段流媒体 |
| `stream_transcode.go` | 流媒体转码 |

#### `transcoder/` — 转码器

| 文件 | 功能 |
|------|------|
| `transcode.go` | 视频转码（编码转换、分辨率调整） |
| `splice.go` | 视频片段拼接 |
| `screenshot.go` | 视频截图 |
| `image.go` | 图片处理（缩略图、格式转换） |

### `backend/pkg/file/` — 文件系统操作

**作用**：抽象的文件系统操作层，处理媒体文件的扫描、导入、移动、删除等。

| 文件 | 功能 |
|------|------|
| `file.go` | 文件通用操作 |
| `folder.go` | 文件夹操作 |
| `scan.go` | 文件扫描器：遍历目录，发现新文件并入库 |
| `import.go` | 文件导入逻辑 |
| `move.go` | 文件移动 |
| `delete.go` | 文件删除 |
| `clean.go` | 文件清理器：检测数据库中已不存在的文件记录并删除 |
| `fs.go` | 文件系统抽象（OsFS 等） |
| `handler.go` | Handler/Filter 接口定义 |
| `walk.go` | 目录遍历工具 |
| `zip.go` | ZIP 文件处理 |
| `stashignore.go` | `.stashignore` 文件解析（忽略规则） |
| `folder_rename_detect.go` | 文件夹重命名检测 |

#### `image/` 子包

| 文件 | 功能 |
|------|------|
| `scan.go` | 图片文件扫描（提取 EXIF 信息、分辨率等） |
| `orientation.go` | 图片方向检测 |

#### `video/` 子包

| 文件 | 功能 |
|------|------|
| `scan.go` | 视频文件扫描（通过 FFprobe 提取元数据） |
| `caption.go` | 字幕文件处理 |
| `funscript.go` | 互动脚本文件处理 |

### `backend/pkg/scraper/` — 数据抓取器

**作用**：从外部网站获取元数据（演员信息、场景信息等）的抓取框架。

| 文件 | 功能 |
|------|------|
| `scraper.go` | 抓取器接口定义和核心结构 |
| `cache.go` | 抓取器缓存（加载和管理所有已配置的抓取器） |
| `definition.go` | 抓取器定义解析（YAML 配置文件） |
| `defined_scraper.go` | 已定义抓取器的统一执行 |
| `url.go` | URL 抓取实现 |
| `script.go` | 脚本式抓取器（执行用户脚本） |
| `graphql.go` | GraphQL 式抓取器（从 Stash-box 等 GraphQL 源获取） |
| `xpath.go` | XPath 式抓取器（HTML 解析） |
| `json.go` | JSON 式抓取器（JSON API 解析） |
| `mapped.go` | Mapped 配置式抓取器（声明式字段映射） |
| `freeones.go` | Freeones 专用抓取器 |
| `performer.go` / `scene.go` / `movie.go` / `gallery.go` / `tag.go` / `image.go` / `country.go` | 各内容类型的抓取逻辑 |
| `postprocessing.go` | 抓取结果的后期处理 |
| `cookies.go` | Cookie 管理 |
| `stash.go` | Stash-box 集成 |
| `autotag.go` | 自动标签（根据抓取结果自动匹配） |

### `backend/pkg/plugin/` — 插件系统

**作用**：基于 YAML 配置的插件系统，支持多种运行时。

| 文件 | 功能 |
|------|------|
| `plugins.go` | 插件缓存和管理核心 |
| `config.go` | 插件配置加载 |
| `convert.go` | 插件配置格式转换 |
| `args.go` | 插件参数处理 |
| `task.go` | 插件任务执行 |
| `hooks.go` | 插件钩子系统 |
| `rpc.go` | 插件 RPC 调用 |
| `raw.go` | 原始命令执行插件 |
| `js.go` | JavaScript 插件支持 |
| `setting.go` | 插件设置管理 |

#### `common/` — 插件通用类型

| 文件 | 功能 |
|------|------|
| `rpc.go` | 插件 RPC 消息格式 |
| `msg.go` | 插件消息格式 |
| `doc.go` | 包文档 |

#### `hook/` — 插件钩子

`hooks.go` — 定义插件可注册的钩子点（场景创建后、场景更新后等生命周期事件）

#### `examples/` — 插件示例

包含多种语言的插件示例：
- `goraw/` — Go 原生命令插件
- `gorpc/` — Go RPC 插件
- `js/` — JavaScript 插件
- `python/` — Python 插件（含 stash_interface.py）
- `react-component/` — React UI 组件插件

### `backend/pkg/exec/` — 外部命令执行

| 文件 | 功能 |
|------|------|
| `command.go` | 通用命令执行封装 |
| `shell_windows.go` | Windows Shell 特殊处理 |
| `shell_nonwindows.go` | Unix Shell 处理 |

### `backend/pkg/job/` — 任务队列

**作用**：后台任务的调度、进度追踪和取消管理。

| 文件 | 功能 |
|------|------|
| `job.go` | Job 结构体定义（含进度、状态、取消上下文） |
| `manager.go` | Job Manager：维护任务队列、并发控制、Job 生命周期管理 |
| `progress.go` | 进度追踪（Total/Current 进度条模型） |
| `subscribe.go` | Job 事件订阅（新 Job 添加、Job 完成等通知） |
| `task.go` | 任务接口定义 |

### `backend/pkg/logger/` — 日志系统

**作用**：统一的日志抽象层，支持多 Handler 输出。

| 文件 | 功能 |
|------|------|
| `logger.go` | Logger 接口定义和全局 Logger 变量 |
| `slog_logger.go` | 基于 Go 标准库 `log/slog` 的实现 |
| `multi_handler.go` | 多 Handler 合并（文件 + UI 同时输出） |
| `ui_handler.go` | UI Handler（将日志通过 EventEmitter 推送到前端） |
| `plugin.go` | 插件日志处理器 |

### `backend/pkg/hash/` — 哈希计算

| 文件 | 功能 |
|------|------|
| `key.go` | 哈希键类型定义 |
| `imagephash/phash.go` | 图片感知哈希（Perceptual Hash） |
| `videophash/phash.go` | 视频感知哈希 |
| `md5/md5.go` | MD5 哈希计算 |
| `oshash/oshash.go` | OpenSubtitles 哈希算法 |

### `backend/pkg/fsutil/` — 文件系统工具

| 文件 | 功能 |
|------|------|
| `fs.go` | 通用文件系统操作 |
| `file.go` | 文件操作（跨平台） |
| `dir.go` | 目录操作 |
| `symwalk.go` | 符号链接感知的目录遍历 |
| `lock_manager.go` | 文件读锁管理器（防止并发读取冲突） |
| `trash.go` | 删除到回收站 |

### `backend/pkg/utils/` — 通用工具集

| 文件 | 功能 |
|------|------|
| `boolean.go` | 布尔值工具 |
| `date.go` / `time.go` | 日期时间工具 |
| `func.go` | 函数工具 |
| `http.go` | HTTP 工具 |
| `image.go` | 图片工具（尺寸、格式） |
| `map.go` | Map 工具 |
| `mutex.go` | 互斥锁封装（支持 context 取消） |
| `natural.go` | 自然排序 |
| `phash.go` | 感知哈希工具 |
| `reflect.go` | 反射工具 |
| `resources.go` | 嵌入资源工具 |
| `strings.go` | 字符串工具 |
| `url.go` / `urlmap.go` | URL 处理和映射 |
| `user_agent.go` | User-Agent 管理 |
| `vtt.go` | VTT 字幕文件工具 |

### `backend/pkg/session/` — 会话管理

| 文件 | 功能 |
|------|------|
| `session.go` | Session Store 主实现 |
| `local.go` | 本地会话管理 |
| `plugin.go` | 插件会话管理 |
| `config.go` | 会话配置 |

### `backend/pkg/pkg/` — 包管理器

**作用**：管理远程包（插件、抓取器）的下载、安装和更新。

| 文件 | 功能 |
|------|------|
| `manager.go` | 包管理器主结构 |
| `store.go` | 本地包存储 |
| `repository.go` | 远程仓库接口 |
| `repository_http.go` | HTTP 远程仓库实现 |
| `cache.go` | 包缓存 |
| `pkg.go` | 包定义 |

### `backend/pkg/javascript/` — JavaScript 虚拟机

**作用**：基于 [goja](https://github.com/dop251/goja) 提供 JavaScript 执行环境。

| 文件 | 功能 |
|------|------|
| `vm.go` | VM 封装：创建 goja Runtime、脚本编译执行 |
| `gql.go` | GraphQL 工具（JS 脚本中调用 GraphQL） |
| `log.go` | JS 日志接口 |
| `console.go` | JS console 对象实现 |
| `util.go` | JS 工具函数 |

### `backend/pkg/match/` — 路径匹配

| 文件 | 功能 |
|------|------|
| `path.go` | 路径匹配算法（根据文件路径匹配场景） |
| `cache.go` | 匹配结果缓存 |
| `scraped.go` | 抓取数据匹配 |

### `backend/pkg/txn/` — 事务管理

| 文件 | 功能 |
|------|------|
| `transaction.go` | 事务管理器 |
| `hooks.go` | 事务钩子 |

### `backend/pkg/sliceutil/` — 切片工具集

| 文件 | 功能 |
|------|------|
| `collections.go` | 通用切片工具（Filter、Map、Contains 等） |
| `intslice/int_collections.go` | int 切片专用工具 |
| `stringslice/string_collections.go` | string 切片专用工具 |

### `backend/pkg/python/` — Python 环境

`env.go` — 检测系统 Python 解释器路径

---

## 七、构建配置 (`build/`)

| 文件/目录 | 功能 |
|-----------|------|
| `config.yml` | Wails v3 项目配置：应用名称、版本、开发者模式、文件夹关联等 |
| `Taskfile.yml` | 通用构建任务：`go mod tidy`、前端依赖安装、前端构建、bindings 生成 |
| `appicon.png` | 应用图标源文件 |
| `appicon.icon/` | macOS 图标资源 |
| `windows/` | Windows 打包：NSIS 安装器脚本、MSIX 配置、图标 |
| `linux/` | Linux 打包：AppImage 构建脚本、nfpm 打包配置 |
| `darwin/` | macOS 图标 |
| `docker/` | Docker 构建配置：交叉编译、Server 模式 |

---

## 八、迁移与原项目的主要区别

| 方面 | 原项目 (stash) | 迁移后 (Case) |
|------|----------------|---------------|
| **架构模式** | Go HTTP Server + React SPA | Wails v3 桌面应用 + Svelte |
| **API 层** | GraphQL (gqlgen) + REST | Wails Service bindings (Go ↔ TS 直接调用) |
| **前端框架** | React (TypeScript) | Svelte 5 (TypeScript) |
| **前端状态管理** | Apollo Client GraphQL hooks | Zustand + TanStack Svelte Query |
| **通信方式** | HTTP + WebSocket (GQL subscriptions) | Wails IPC (Go ↔ JS 直接绑定) |
| **DLNA 服务** | 有（网络流媒体共享） | 已移除（桌面应用不需要） |
| **多用户认证** | 有（登录、API Key、会话） | 简化为单用户桌面场景 |
| **Stash-box 集成** | 有（内容订阅、自动标签） | 已移除相关代码，将来用插件替代 |
| **Server 模式** | 原生 HTTP Server | 通过 Taskfile 可选支持 |
| **UI 自定义** | 插件 UI (React 组件注入) | Svelte 原生组件 |
| **图标/资源** | 嵌入的 favicon 等 | Wails 构建系统管理 |

---

## 九、关键数据流

### 扫描流程

```
用户操作 → manager.Scan() → job.Manager.Add(ScanJob)
  → file.Scanner 遍历目录
    → 对每个文件：
      1. video.Decorator (通过 FFprobe 提取视频元数据)
      2. image.Decorator (提取图片 EXIF 等)
      3. fingerprintCalculator (计算文件哈希)
    → 将新发现文件写入 SQLite
    → 发送扫描完成通知 (scanSubs)
```

### 生成流程

```
用户操作 → manager.Generate() → job.Manager.Add(GenerateJob)
  → GenerateCoverTask / GenerateSpriteTask 等
    → 调用 ffmpeg.FFMpeg 执行截图/预览/Sprite 生成
    → 输出文件写入 generated/ 目录
    → 更新数据库中的文件路径
```

### 抓取流程

```
用户触发抓取 → scraper.Cache 根据 URL 匹配抓取器
  → 执行抓取 (Script/XPath/JSON/GraphQL 等)
  → 返回 ScrapedContent (演员/场景/图库等)
  → 用户确认后写入数据库
```

---

*本文档基于项目代码的实际结构和依赖关系生成，随项目迭代可能发生变化。*