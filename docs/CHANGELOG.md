# CHANGELOG

## 2026-09-23（阶段 2 收尾）

### 修改

- `internal/app/cover.go`：`CoverURL` 缓存参数由 `?v=<updatedAt>` 改为 `?v=<md5(bytes)[:8]>`（`UpdateCover` 不更新 `scenes.updated_at`）；`image.DecodeConfig` 失败时置空 URL，走占位 box
- `internal/app/scene.go`：移除不再需要的 `updatedAt` 传递
- `docs/PROJECT_STATE.md`：标记阶段 2 完成，焦点转阶段 3；已知问题精简为阶段 3 相关，其余指向 `docs/TECH_DEBT.md`
- `docs/ROADMAP.md`：阶段 2 全部勾选；封面缩略图顺延到阶段 3

### 新增

- `backend/pkg/metadata/cover/cover_test.go`：新增真实 FFmpeg E2E 用例 `TestResolve_FFmpegRealFrame`（生成测试视频 → 20% 截帧 → 解码校验；无 ffmpeg 时 skip）

### 验证

- `gofmt` / `go vet ./internal/app/... ./backend/pkg/metadata/cover/...`：通过
- `go test ./backend/pkg/metadata/cover/...`：7 个用例（含真实 FFmpeg E2E）全部 PASS
- `go test ./internal/app/... ./backend/pkg/metadata/nfo/... ./backend/pkg/scene/...`：全部 PASS
- `go build ./...`：通过
- `wails3 dev`：构建成功、AssetServer middleware 生效、WebView2 启动成功

### 说明

- 端到端「扫描 → 列表 → 详情 → 封面 → 外部链接 → 返回」中的 **GUI 交互部分需人工点按确认**；自动化覆盖：构建、启动、封面策略链（含真实 FFmpeg）、NFO/场景单测。

## 2026-09-23（阶段 2.4.5 - 封面图）

### 新增

- `backend/pkg/metadata/cover/cover.go`：封面策略链 `poster → 同名图 → FFmpeg 20%`，含 6 个用例
- `internal/app/cover.go`：`ensureCover`（按需生成并持久化）+ `CoverMiddleware`（`/covers/<sceneID>` 输出，包级函数避免被 Wails 绑定）
- 详情页封面展示 + 无图占位（首字母 + `hsl(id*47 % 360, 20%, 15%)`，`aspect-ratio: 2/3`）
- `docs/TECH_DEBT.md`：分类技术债清单

### 修改

- `internal/app/scene.go`：`SceneDetailDTO` 增 `CoverURL` / `CoverWidth` / `CoverHeight`；`GetScene` 在读取事务外按需生成封面
- `main.go`：`AssetOptions.Middleware` 注册 `app.CoverMiddleware`
- `docs/PROJECT_STATE.md`：新增 5 条封面技术债 + 指向 `docs/TECH_DEBT.md`

### 修复

- 无

### 验证

- `gofmt` / `go vet ./backend/pkg/metadata/cover/... ./internal/app/...`：通过
- `go test ./backend/pkg/metadata/cover/... ./backend/pkg/metadata/nfo/... ./backend/pkg/scene/... ./internal/app/...`：全部 PASS
- `go build ./...`：通过
- `npm run check`：0 errors / 0 warnings；`npm run build`：通过
- `wails3 dev`：构建成功、AssetServer middleware 生效、WebView2 启动成功

### 决策

- 封面复用 Stash blob store（`Scene.UpdateCover`），文件落 `data/blobs/xx/yy/md5`，DB 记 `cover_blob` checksum，**不新增迁移**
- 详情页按需生成（零冻结区改动）；截帧 20%（与 `GenerateCoverTask` 一致）
- `CoverURL` 带 `?v=<updatedAt>` 处理浏览器缓存

## 2026-09-23（阶段 2.4.4 - 详情页 + hash 路由）

### 新增

- `frontend/src/lib/router.svelte.ts`：自研 hash 路由（`#/`、`#/scenes/:id`）
- `frontend/src/lib/format.ts`：`formatDuration` / `formatBitrate`
- `frontend/src/lib/external.ts`：`openExternal`（`@wailsio/runtime` 的 `Browser.OpenURL`，失败仅 `console.error`）
- `frontend/src/lib/components/SceneList.svelte`：扫描区 + 列表（由 App.svelte 原样迁出，卡片可点击跳详情）
- `frontend/src/lib/components/SceneDetail.svelte`：详情页（加载三态 + `requestId` 竞态处理 + 返回 + 外部链接）
- `docs/ARCHITECTURE.md`：ADR-006（自研 hash 路由）、ADR-007（详情数据 `$state` + `onMount`）

### 修改

- `frontend/src/App.svelte`：改为路由 shell，按 route 分发 SceneList / SceneDetail

### 修复

- 无

### 验证

- `npm run check`（svelte-check）：0 errors / 0 warnings
- `npm run build`：通过
- `wails3 dev`：分别在第 3 步（列表迁移后）与第 4 步（详情接入后）冒烟，构建成功、WebView2 启动成功
- 手工冒烟点：列表 → 详情 → 返回、深链 `#/scenes/<id>`

## 2026-09-23（阶段 2.4.3 - GetScene 完整 DTO）

### 新增

- `internal/app/scene.go`：`GetScene(id)` 返回 `SceneDetailDTO`（标量元数据 + `StudioName` + 结构化 `Tags []TagDTO` / `Performers []PerformerDTO` + URLs + 主文件便利字段 + `files[]`）
- `internal/app/scene_test.go`：`toSceneDetailDTO` 纯映射 5 个用例（完整场景、空关联/nil 字段、Inf/NaN、竖屏分辨率、resolutionLabel）

### 修改

- 无（未触碰冻结区；`SceneDTO` / `FindScenes` 保持不变，列表仍轻量）

### 修复

- 无

### 验证

- `gofmt -l internal/app`：`scene.go` / `scene_test.go` 无输出（其余 `internal/app` 既有文件未格式化，与本次无关）
- `go vet ./internal/app/...`：通过
- `go test ./internal/app/... -v`：5 个用例全部 PASS
- `go build ./...`：通过
- `wails3 dev`：构建成功、绑定生成包含 `GetScene` / `SceneDetailDTO` / `TagDTO` / `PerformerDTO`、WebView2 启动成功

### 决策

- `Rating` 用 `int`：0 表示「无评分」（有效范围 1-100）
- `Duration`/`FrameRate` 用 `float64`：`DurationFinite()`/`FrameRateFinite()` 将 Inf/NaN 归零，0 表示「未知」
- `resolution` 由 `min(width,height)` 派生为 `"1080p"` 形式，0 维度返回空串

## 2026-09-23（阶段 2.4.2 - 扫描时自动读 NFO）

### 新增

- `backend/pkg/metadata/nfo/applier.go`：`Applier.Apply` 读取视频同名 NFO 并回填场景元数据；文件 I/O 在事务外，写库在单个 `WithTxn` 内；解析失败仅告警跳过
- `backend/pkg/metadata/nfo/applier_test.go`：基于 `models/mocks` 的 5 个用例（无 NFO、坏 NFO、填充空场景、不覆盖已有字段、复用已有标签/演员）

### 修改

- `backend/pkg/scene/scan.go`：新增可选接口 `ScanMetadataApplier` 与 `ScanHandler.MetadataApplier` 字段；在既有 post-commit hook 内 nil 门控调用；`validate()` 不变
- `backend/manager/task_scan.go`：`getScanHandlers` 注入 `&nfo.Applier{Repo: mgr.Repository}`

### 修复

- 无

### 验证

- `gofmt -l backend/pkg/metadata/nfo backend/pkg/scene backend/manager`：新增/改动文件无输出（`scene/generate/sprite.go`、`manager/task_import.go` 为既有未格式化文件）
- `go vet ./backend/pkg/metadata/nfo/...`：通过（`backend/manager` 既有 vet 告警与本次无关）
- `go test ./backend/pkg/metadata/nfo/... ./backend/pkg/scene/...`：全部 PASS
- `go build ./...`：通过
- `wails3 dev`：CGO=1 构建成功、连接前端 dev server、WebView2 启动成功

## 2026-09-23（阶段 2.4 - NFO 解析模块）

### 新增

- `backend/pkg/metadata/nfo/`：Go 原生 NFO 解析模块（ADR-004）
  - `movie.go`：标准 Kodi `<movie>` XML DTO
  - `nfo.go`：`Parse` / `ParseFile` / `FindForVideo`，严格解码、拒绝 DTD、10 MiB 上限、UTF-8 BOM 兼容
  - `mapping.go`：`Movie → SceneMetadata` 纯映射（日期回退、评分 0-10 → 1-100、名字 NFKC 归一化与去重、类型过滤）
  - `nfo_test.go` / `mapping_test.go` / `testdata/FDD-2002.nfo`

### 修改

- `mapping.go`：`normalizeName` 由 NFC 改为 NFKC，全角/半角名字可正确去重
- `mapping.go`：`firstDate` 的 `title` 参数重命名为 `logContext`（仅用于日志）

### 修复

- 无

### 验证

- `gofmt -l backend/pkg/metadata/nfo`：无输出
- `go vet ./backend/pkg/metadata/nfo/...`：通过
- `go test ./backend/pkg/metadata/nfo/... -v`：15 个用例全部 PASS（新增全角去重、已归一化评分、目录报错、大写 `.NFO` 扩展名）
- `go build ./...`：通过
- `wails3 dev`：生成绑定、构建 `bin/stash-desktop.exe`、连接前端 dev server、WebView2 启动成功

## 2026-09-23

### 新增

- 创建 Agent 文档体系：AGENTS.md、docs/PROJECT_STATE.md、docs/WORKFLOW.md、docs/AGENT_TOOLS.md、docs/TROUBLESHOOTING.md、docs/ARCHITECTURE.md、docs/CONVENTIONS.md、docs/ROADMAP.md

### 修改

- 无

### 修复

- 无

### 验证

- 文档结构确认