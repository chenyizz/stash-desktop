# 项目状态

## 最后更新：2026-09-24

## 当前阶段

阶段 3：核心业务（阶段 2 最小闭环已于 2026-09-23 完成）

## 已完成

- Wails v3 + Svelte 5 + TypeScript 项目骨架
- Go 环境：CGO=1，MinGW-w64 16.2.0
- `backend/` 从 Stash 迁移：manager + pkg
- 日志系统：slog + 多 Handler + UI 推送
- config 接入 ServiceStartup
- Manager 初始化：SQLite 创建 + 迁移 + FFmpeg 检测
- NVENC 硬件加速检测：h264_nvenc 可用
- 扫描文件夹 → 写库
- 场景列表显示：1 条记录
- NFO 解析模块：`backend/pkg/metadata/nfo/`（解析 + `SceneMetadata` 映射 + 11 用例）
- 扫描时自动读 NFO 并回填场景元数据（`nfo.Applier`，只填不覆盖）
- `GetScene(id)` 返回完整 `SceneDetailDTO`（元数据 + 结构化 tags/performers + files）
- 详情页 Svelte 组件 + 自研 hash 路由（列表 ↔ 详情）
- 封面图：`backend/pkg/metadata/cover/` 策略链（poster → 同名图 → FFmpeg 20%）+ `/covers/<id>` 中间件 + 详情页展示/占位
- 扫描完成事件：后端 `scan:complete`（`internal/app/events.go`）+ 前端监听，替代 `setTimeout`
- 列表分页：`FindScenes` 返回 `ScenesPageDTO`（`total/page/pageSize`）+ 前端分页器
- 列表标签/演员：`SceneDTO` 扩展结构化 `tags/performers`（本页批量取名称）+ 前端 chips
- 演员/标签/工作室列表页：`FindPerformers` / `FindTags` / `FindStudios`（分页）+ 通用 `TaxonomyList` + 顶部导航
- 封面缩略图 + 列表海报墙：`/covers/<id>?w=320`（imaging 缩放 + 磁盘缓存/ETag），卡片 16/9 海报
- 详情页附件（D2 方案 A）：`backend/pkg/metadata/attachments` + `/attachments/<sceneID>/<index>` + 详情页附件区块（不入库）
- 库路径扫描模式 `LibraryMode`（`Videos/Images/Attachments`，默认全开）：扫描分派与附件开关接入
- 搜索：查询结构体入参（过滤只加字段）；IME 组合态 + 300ms 防抖 `SearchBox`；场景搜索覆盖 tag/演员/工作室

## 当前焦点

**阶段 3：核心业务（演员/标签/工作室/搜索/分页）**

## 下一步

1. 阶段 3 结束前写 `docs/MIGRATIONS.md`（记录当前 schema、每张表作用、停用表、删除顺序）

## 已知问题（阶段 3 相关）

- 无虚拟滚动
- 同场景多文件首次扫描时标量字段由并行顺序决定（阶段 3+ 再设计优先级）
- 取消扫描时后端不发 `scan:complete`，前端需手动刷新（用刷新按钮兜底）

> 其余技术债（封面同步/缩略图/缓存/NFO 编码/测试缺口/工程化等）统一见 `docs/TECH_DEBT.md`，此处不再重复。

## 最近变更

- 搜索：`ScenesQuery` 等结构体入参 + `normalizeQuery`（rune 上限）；`SearchBox.svelte`（IME+防抖）接入两个列表；场景搜索覆盖 tag/演员/工作室
- 库路径 `LibraryMode`：config 结构/默认/校验 + `useAs*`/`scanFilter`/附件按 mode 分派
- 列表分页 + 标签/演员：`FindScenes` 返回 `ScenesPageDTO`；`SceneDTO` 含 `tags`/`performers`（本页批量取名称）
- 扫描完成事件：`internal/app/events.go`（watcher + 契约 `scan:complete` + `{at}`）替代前端 `setTimeout`
- 封面收尾：`CoverURL` 缓存参数改用 `md5(bytes)[:8]`（`UpdateCover` 不改 `updated_at`）；解码失败置空走占位
- 新增封面：blob 存储 + 策略链解析 + `/covers/<id>` 中间件 + 详情页封面/占位 + DTO 三个封面字段
- 新增详情页 + hash 路由（`frontend/src/lib/`）：SceneList/SceneDetail、router、format、external
- 新增 ADR-006（自研 hash 路由）、ADR-007（详情数据用 `$state` + onMount）
- 新增 `GetScene` 详情 DTO：tags/performers 结构化（ID+Name），files 含 duration/分辨率/codec
- 扫描接入 NFO：scene post-commit hook 调用 `nfo.Applier`，只填不覆盖、关系只增
- 完成 NFO 解析模块（Go 原生，含映射与测试）
- 完成扫描 → 写库
- 完成场景列表显示