# 架构决策

## ADR-001：用 Wails 替代 GraphQL

决策：去掉 GraphQL 层，用 Wails 方法绑定直接通信。

理由：

- Wails 是进程内通信，无 HTTP 开销
- 单次调用延迟低
- 不需要维护 Schema 和 Resolver

影响：

- 前端数据请求改为 `await App.XXX()`
- 后端不暴露 HTTP 端口
- GraphQL 相关代码删除

## ADR-002：Blob 存储用文件系统

决策：`blobs_storage = FILESYSTEM`，不存 SQLite。

理由：

- Stash 原版数据库可能很大
- 文件系统存储可单独管理、清理、备份
- SQLite 保持小巧

## ADR-003：数据目录策略

决策：exe 同目录可写 → 用 exe 目录；不可写 → `%LOCALAPPDATA%\case`。

理由：

- 便携软件风格
- 装在 Program Files 时自动降级
- 可用 `CASE_DATA_DIR` 强制指定

## ADR-004：NFO 解析用 Go 重写

决策：不用 Python 插件，用 Go 原生解析 NFO。

理由：

- 同进程调用，无 stdin/stdout 开销
- 打包进 exe，零依赖
- 性能更好，调试更简单

## ADR-005：Scraper 与插件分离

决策：当前阶段不做 Scraper。阶段 8 先做插件桥，之后再把 Scraper 设计成插件的一种类型。

理由：

- Scraper 需要 HTTP、站点适配、认证、限流、缓存、匹配、合并。
- 通用插件系统只解决加载与运行，不解决 Scraper 全部问题。
- 当前阶段只需要本地 NFO 解析。

## ADR-006：阶段 2.4 用自研 hash 路由

决策：极简自研，不引入路由库。

理由：

- 只有 2 个路由，Wails WebView hash 模式最稳，YAGNI。

重评时机：

- 阶段 3 页面超 5 个，或需要嵌套路由/守卫/代码分割。

约束：

- 路由逻辑集中在 `frontend/src/lib/router.svelte.ts`，`App.svelte` 只做分发。

## ADR-007：阶段 2.4 详情数据用 `$state` + `onMount`

决策：详情数据不引入 TanStack Query，用 `$state` + `onMount`。

理由：

- 只有 2 个查询，TanStack Query 当前收益 < 成本。

实现约束：

- 竞态用 `requestId` 丢弃过期响应；
- `onDestroy` 时递增 `requestId`，避免已销毁组件 setState；
- loading/error/success 用独立 `$state`。

重评时机：

- 阶段 3 引入分页/虚拟滚动/搜索时重新评估。

## ADR-008：媒体扫描模式

决策：库路径使用显式 `LibraryMode`（**双布尔 flags，非三选一枚举**）：

```go
type LibraryMode struct {
    Videos      bool // 是否扫描视频文件
    Images      bool // 是否扫描图片文件
    Attachments bool // 识别 fanart/extra/poster 作为附件
}
```

图片集**复用 Gallery**（一个最小文件夹 = 一个图片集）；附件机制**待阶段 3 决策**。

理由：

- 媒体类型当前只有视频/图片，双布尔覆盖全部组合（只视频/只图片/混合），且未来加音频、字幕是同一模式扩展；三选一枚举会随模式增加膨胀。
- Gallery 已具备图片集所需的模型、关系、封面（`FolderID` 绑定一个文件夹，契合「最小文件夹 = 一集」）。
- 附件当前没有「零成本且不带 hack」的实现，先记录候选方案，不提前欠债。

约束：

- `LibraryMode` 仅影响扫描分派与过滤；默认 `{Videos:true, Images:true, Attachments:true}` 保持现有行为。
- 配置校验要求 `Videos || Images`，两者皆 false 时返回错误「至少启用一种媒体类型」。
- 图片集沿用 Gallery，不新增表；Gallery NFO 解析为新增包。
- 附件当前**不入库、不索引**；候选：
  - A 文件系统直读（零冻结，代价是每次扫目录）
  - B custom fields 存 JSON 字符串（零冻结，属 hack）
  - C 新建 `scene_attachments` 表（碰冻结区，可搜索/过滤/加元数据）
- 详见 `docs/DESIGN_MEDIA_SCAN.md`。

重评时机：

- 需要例 2（多文件夹 = 一集）→ Gallery 分组或 ImageSet。
- 附件需要元数据/搜索/统计 → 选方案 C。
- 媒体类型超过 5 种或需要互斥组合 → 升级为枚举 + 集合。
- 需要音频等其他模式 → 扩展 flags。

参考：`docs/DESIGN_MEDIA_SCAN.md`（草案）。

## ADR-009：StashID 与停用概念的偿还时机

决策：`StashID/StashIDs/stash_id`、Group、SavedFilter、SceneMarker，以及 `scraper`/`match` 子系统，**当前保留现有代码和数据库表**，不在命名清理步骤中处理。

偿还时机：

- 阶段 3：`group` / `savedfilter` 决定是否进核心业务；`signedurl` 抽查流媒体功能。
- 阶段 8（插件）：`scraper` / `match` 随插件桥设计评估；`plugin/util` 随示例决定。
- 阶段 9（性能）：`context.TODO()` 19 处收敛。
- 阶段 11（重构）：`session` 桌面端残留、数据库停用表（需新迁移）、TODO 132 / deprecated 58、小写 `stash` 注释与夹具。

理由：

- 当前无功能依赖，重构收益为零。
- 现在拍板结构可能拍错（插件桥尚未定型）。
- 明确偿还时机，不是"忘了改"。

完整清单：`docs/TECH_DEBT.md`「偿还时机清单」与 `docs/HEALTH_SCAN.md`。

## ADR-010：编辑数据流与 NFO 回写

决策：

- **DB 是唯一事实源**。前端编辑 → `internal/app` 写方法 → DB。
- **NFO 默认不回写**（保持为导入源）；提供**显式「导出 NFO」**动作（单场景 / 批量），不做自动回写。
- 导入侧保持 **fill-only**（只填不覆盖），因此 DB 编辑不会被重扫冲掉。
- 导出前检查 NFO 是否被外部修改（mtime 或内容 hash）；若被改过，**提示用户选择**（覆盖 / 跳过 / 另存）。
- NFO 含未知标签时，导出**默认不覆盖**（跳过或另存），避免丢数据。

**显式后果（重要取舍）**：改了 DB 但不执行导出，**NFO 文件不会跟着变**。反向亦然——外部直接改 NFO 不会自动进 DB，需重新扫描（且 fill-only 不会覆盖已有值）。

理由：

- 双源写入必然冲突；把 NFO 降级为「导入 + 显式导出」可避免隐式覆盖。
- fill-only 已在 2.4.2 实现，行为一致。

重评时机：

- 需要 NFO 与 DB 自动同步，或 NFO 作为多端共享源时。

## ADR-011：配置持久化与热生效

决策：`config.yml` 是配置源（不是 DB）；Settings 页经 `backend/manager/config` 的 setter 写入并 `Write()` 持久化。

**生效矩阵（决策项，不展开实现）**：

- 热生效：主题、日志级别、UI 偏好（无需重启/重扫）。
- 需重扫 / 重建：库路径、每路径 `LibraryMode`、附件目录、扫描默认项（重建 `paths`、提示或触发重扫）。
- 无需生效：备份路径、数据库优化等一次性操作参数。

库路径变更流程：校验（存在 / 可读 / 去重）→ 需要时创建目录 → 重建 `paths` → 提示重扫。

首次使用：库路径为空时提供引导（空态 CTA），避免必须手改 YAML。

冻结区：`backend/manager/config`（按 `docs/FROZEN_RULES.md` 走 Plan）。

重评时机：引入 DB 配置或多配置源时。

## ADR-012：写路径统一模式

决策：写路径统一为「后端写服务 + 前端统一表单」，不引入乐观更新（保守策略）。

- 后端：`internal/app` 统一写服务模式——参数校验 → Repository 写方法 → 错误映射；不修改已有 Repository 接口（写能力已具备）。
- 前端：统一「表单 + 字段 + 保存态 + 错误态」模式；复用 `SearchBox` / `EntityPicker` 等既有组件。
- 保存失败保留表单内容；不做乐观更新与撤销（本阶段）。
- `Studio.Merge` 缺失：**暂不新增**，等真实重名需求。

理由：

- 配置写与内容写共享同一套写路径基础设施，先建立可减少返工。
- 无乐观更新可避免一致性问题，代价是保存后需刷新。

重评时机：表单数量增长或交互体验出现明显问题时，抽独立组件库 / 引入乐观更新。

参考：`docs/ROADMAP.md`（阶段 4-7）。