# 代码地图

## 最后更新：2026-09-23（复核实际文件清单）

## 维护规则

- 新增、删除、重命名模块或关键文件时，Agent **必须**更新本文件。
- 只放**定位信息**：路径、职责、关键类型摘要、搜索命令。
- 不放实现细节、完整代码和易变行号；控制在 200 行以内，超出时下沉到 `docs/CONVENTIONS.md` / `docs/ARCHITECTURE.md`。

---

## 目录结构总览

```text
Case/
├─ backend/
│  ├─ manager/          # 总协调入口
│  │  ├─ config/        # 配置读写与校验
│  │  └─ task/          # 独立任务实现（迁移/清理等）
│  └─ pkg/              # 业务域实现
│     ├─ models/        # 领域模型
│     ├─ sqlite/        # Repository 实现
│     ├─ scene/         # 场景业务逻辑（ScanHandler / Service）
│     ├─ file/          # 文件扫描（video/ 视频元数据装饰器）
│     ├─ ffmpeg/        # FFmpeg / FFProbe 封装
│     ├─ logger/        # 日志
│     ├─ utils/         # 通用工具
│     ├─ fsutil/        # 文件系统工具
│     └─ metadata/      # nfo 解析 + cover 策略链（阶段 2.4）
├─ internal/app/        # Wails 服务层
├─ frontend/            # bindings/（生成）+ src/（App.svelte、lib/router、lib/components）
└─ docs/
```

---

## 核心模块

### backend/manager/

| 文件            | 职责                                                       |
| --------------- | ---------------------------------------------------------- |
| `manager.go`    | Manager 初始化：SQLite 创建、迁移、FFmpeg 检测、NVENC 检测 |
| `repository.go` | Repository 聚合入口，提供 `WithReadTxn` / `WithTxn`        |

**注意**：冻结区。只允许在 `internal/app` 层调用，不要改内部结构。

### backend/pkg/models/

| 文件                 | 职责                                  |
| -------------------- | ------------------------------------- |
| `model_scene.go`     | `Scene`、`ScenePartial`、`RelatedIDs` |
| `model_performer.go` | `Performer`、`PerformerPartial`       |
| `model_tag.go`       | `Tag`、`TagPartial`                   |
| `model_studio.go`    | `Studio`、`StudioPartial`             |
| `date.go`            | `Date`、`ParseDate`、`DateFromYear`   |

**核心类型摘要**（与 `backend/pkg/models` 一致）：

```go
// Scene（精简，model_scene.go）
type Scene struct {
    ID             int
    Title          string
    Code           string
    Details        string
    Director       string
    Date           *Date          // 精度到日/月/年
    ProductionDate *Date
    Rating         *int           // 1-100
    StudioID       *int
    URLs           RelatedStrings
    TagIDs         RelatedIDs
    PerformerIDs   RelatedIDs
    Groups         RelatedGroups
    StashIDs       RelatedStashIDs
    // transient（不入库）：Files / Path / OSHash / Checksum
}
// RelatedIDs：list 私有，只能经 List()/Add() 访问，由 Load* 懒加载。
// Date（date.go）
func ParseDate(s string) (Date, error)
func DateFromYear(year int) Date
```

**Partial / Input**：`ScenePartial`、`CreateSceneInput`、`TagPartial`、`PerformerPartial`、`StudioPartial`。

### backend/pkg/sqlite/

| 文件           | 职责                                                         |
| -------------- | ------------------------------------------------------------ |
| `scene.go`     | `Scene` 的 Repository 实现，含 `Find` / `FindByID` / `UpdatePartial` |
| `performer.go` | `Performer` Finder/Creator                                   |
| `tag.go`       | `Tag` Finder/Creator                                         |
| `studio.go`    | `Studio` Finder/Creator                                      |

### backend/pkg/scene/

| 文件                                    | 职责                              |
| --------------------------------------- | --------------------------------- |
| `scan.go`                               | `ScanHandler`：匹配/创建/关联场景；post-commit 调 `MetadataApplier` |
| `create.go` / `update.go` / `import.go` | 创建、部分更新、JSON 导入         |
| `query.go` / `service.go`               | 查询与 `Service` 聚合             |

**注意**：冻结区。阶段 2.4.2 接入 NFO 时需人工确认。

### backend/pkg/file/video/

| 文件      | 职责                                                  |
| --------- | ----------------------------------------------------- |
| `scan.go` | `Decorator`：提取时长/分辨率/编解码等视频文件元数据 |

**注意**：冻结区。

### backend/pkg/ffmpeg/ 与 backend/manager/config/

| 模块 / 文件                     | 职责                                      |
| ------------------------------- | ----------------------------------------- |
| `ffmpeg.go` / `ffprobe.go` / `codec*.go` | FFmpeg/FFProbe 封装、硬件加速检测（NVENC 等） |
| `stream*.go` / `types.go`       | 转码、流式传输、公共类型与参数            |
| `config.go` / `enums.go` / `library_config.go` / `tasks.go` / `ui.go` | 配置读写校验、枚举、库路径/任务/UI |

**注意**：`backend/manager/config` 属冻结区。

### backend/pkg/metadata/

| 路径      | 职责                                                         |
| --------- | ------------------------------------------------------------ |
| `nfo/`    | Kodi DTO + `Parse`/`ParseFile`；`mapping`；`applier`（扫描回填） |
| `cover/` / `attachments/` | 封面策略链 + 缩略图；附件 resolver（文件系统直读 fanart/poster/extra） |

### backend/pkg/logger/

| 文件        | 职责                                                         |
| ----------- | ------------------------------------------------------------ |
| `logger.go` | `Infof` / `Debugf` / `Warnf` / `Errorf`，slog + 多 Handler + UI 推送 |

### backend/pkg/utils/ 与 backend/pkg/fsutil/

| 文件                                        | 职责                                     |
| ------------------------------------------- | ---------------------------------------- |
| `utils/strings.go` / `utils/date.go`        | 字符串工具、日期解析（`ParseDateStringAsTime`） |
| `utils/phash.go` / `utils/image.go`         | 感知哈希 / 图像工具                      |
| `utils/func.go` / `utils/map.go` / `utils/url.go` / `utils/vtt.go` | 其他通用工具 |
| `fsutil/dir.go` / `fsutil/file.go`          | 目录 / 文件操作（含平台特定实现）        |

### internal/app/

| 文件               | 职责                                                         |
| ------------------ | ------------------------------------------------------------ |
| `app.go` / `query.go` | Wails `ServiceStartup`、`ScanLibrary` / `FindScenes`（分页+搜索）、`SceneDTO`/`ScenesQuery` 等 |
| `scene.go`         | `GetScene` / `SceneDetailDTO` / `TagDTO` / `PerformerDTO` / `SceneFileDTO` |
| `cover.go` / `thumbnail.go` / `assets.go` / `attachment.go` | 封面/缩略图与附件端点；`AssetMiddleware` 按前缀分发 |
| `events.go`        | `watchScanEvents`（扫描完成推送 `scan:complete`）             |
| `taxonomy.go`      | `FindPerformers`/`FindTags`/`FindStudios` + 分类 Page DTO     |
| `config.go` / `logging.go` / `paths.go` / `wails_emitter.go` | 配置、日志、目录布局、事件/日志推送适配 |

**规则**：只做胶水，不写业务逻辑。导出方法首字母大写才会暴露给前端。

### frontend/src/

| 路径                      | 职责                                              |
| ------------------------- | ------------------------------------------------- |
| `App.svelte` / `main.ts`  | 路由 shell（列表 ↔ 详情）与入口挂载              |
| `lib/`                    | `router.svelte.ts`（hash 路由）、`Pager.svelte`、`components/`（SceneList/SceneDetail/TaxonomyList/SearchBox）、`format.ts`、`external.ts`、`events.ts` |

---

## 常用搜索命令

```bash
rg "type Scene struct" backend/            # 找类型定义
rg "func.*FindScenes" backend/ internal/   # 找函数定义
rg "FindScenes\(" backend/ internal/       # 找调用点
rg "SceneDTO" internal/ frontend/src/      # 找 DTO 字段
git diff --name-only HEAD -- backend/manager backend/pkg  # 找冻结区改动
```

---

## 探索规则

1. **优先查本文件**，再决定读哪些源文件。
2. **禁止** `Get-ChildItem -Recurse` 遍历整个目录。
3. 找定义/调用用 `rg`，确认行号后再精确 `read_file`。
4. 一次任务最多读 5 个文件、跑 3 条探索命令，然后必须出计划。
5. 只读 `backend/pkg` 中与当前任务相关的包，不要全读。

---

## 变更日志（本文件）

| 日期       | 变更                                                         |
| ---------- | ------------------------------------------------------------ |
| 2026-09-24 | 阶段 3.7-3.10：缩略图/海报墙；附件；`LibraryMode`；搜索（查询结构体 + `SearchBox`） |
| 2026-09-24 | 阶段 3.4-3.6：`Pager.svelte`、router 前缀路由 + `NAV_ITEMS`、`taxonomy.go` + `TaxonomyList.svelte` |
| 2026-09-24 | 阶段 3.1-3.3：`scan:complete`；`ScenesPageDTO`（分页）；`SceneDTO` 加 tags/performers |
| 2026-09-23 | 命名清理（阶段 5）：`StashConfig→LibraryConfig`、`stash_config.go→library_config.go`、`stashignore.go→caseignore.go`、`GetStashHomeDirectory→GetCaseHomeDirectory`、`STASH_*`→`CASE_*` |
| 2026-09-23 | 阶段 2.4：NFO applier 接入；详情页 + hash 路由；`metadata/cover/` + `internal/app/cover.go` |
| 2026-09-23 | 复核实际文件并创建：修正 models 类型摘要，补 `ffmpeg`/`manager/config`/`scene`/`metadata` 模块 |
