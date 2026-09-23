# 项目状态

## 最后更新：2026-09-23

## 当前阶段

阶段 2：最小闭环（扫描 → 列表 → 详情）

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

## 当前焦点

**阶段 2.4：场景详情页**

## 下一步

1. 详情页 Svelte 组件
2. 封面图：AssetServer

## 已知问题

- 同场景多文件首次扫描时，标量字段由并行扫描完成顺序决定，后续扫描不改变。如需确定性优先级，阶段 3+ 再设计
- 扫描完成用 `setTimeout` 等，不可靠
- 无分页
- 无虚拟滚动
- 列表看不到标签/演员

## 最近变更

- 新增 `GetScene` 详情 DTO：tags/performers 结构化（ID+Name），files 含 duration/分辨率/codec
- 扫描接入 NFO：scene post-commit hook 调用 `nfo.Applier`，只填不覆盖、关系只增
- 完成 NFO 解析模块（Go 原生，含映射与测试）
- 完成扫描 → 写库
- 完成场景列表显示