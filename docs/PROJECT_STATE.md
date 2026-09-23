# 项目状态

## 最后更新：2026-09-23

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

## 当前焦点

**阶段 3：核心业务（演员/标签/工作室/搜索/分页）**

## 下一步

1. 扫描完成事件，替代 `setTimeout`
2. 列表分页 / 标签 / 演员展示
3. 封面缩略图与列表海报墙
4. 阶段 3 结束前写 `docs/MIGRATIONS.md`（记录当前 schema、每张表作用、停用表、删除顺序）

## 已知问题（阶段 3 相关）

- 扫描完成用 `setTimeout` 等，不可靠 → 阶段 3 改事件通知
- 无分页
- 无虚拟滚动
- 列表看不到标签/演员
- 同场景多文件首次扫描时标量字段由并行顺序决定（阶段 3+ 再设计优先级）

> 其余技术债（封面同步/缩略图/缓存/NFO 编码/测试缺口/工程化等）统一见 `docs/TECH_DEBT.md`，此处不再重复。

## 最近变更

- 封面收尾：`CoverURL` 缓存参数改用 `md5(bytes)[:8]`（`UpdateCover` 不改 `updated_at`）；解码失败置空走占位
- 新增封面：blob 存储 + 策略链解析 + `/covers/<id>` 中间件 + 详情页封面/占位 + DTO 三个封面字段
- 新增详情页 + hash 路由（`frontend/src/lib/`）：SceneList/SceneDetail、router、format、external
- 新增 ADR-006（自研 hash 路由）、ADR-007（详情数据用 `$state` + onMount）
- 新增 `GetScene` 详情 DTO：tags/performers 结构化（ID+Name），files 含 duration/分辨率/codec
- 扫描接入 NFO：scene post-commit hook 调用 `nfo.Applier`，只填不覆盖、关系只增
- 完成 NFO 解析模块（Go 原生，含映射与测试）
- 完成扫描 → 写库
- 完成场景列表显示