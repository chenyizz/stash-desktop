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

决策：当前阶段不做 Scraper。阶段 4 先做插件桥，阶段 4+ 再把 Scraper 设计成插件的一种类型。

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

决策：库路径使用显式 `LibraryMode`（mode + flags，**非三选一枚举**）；图片集**复用 Gallery**（一个最小文件夹 = 一个图片集）；附件机制**待阶段 3 决策**。

理由：

- 三选一枚举会随模式增加膨胀；flags 结构可扩展，且能表达「扫视频 + 带附件」。
- Gallery 已具备图片集所需的模型、关系、封面（`FolderID` 绑定一个文件夹，契合「最小文件夹 = 一集」）。
- 附件当前没有「零成本且不带 hack」的实现，先记录候选方案，不提前欠债。

约束：

- `LibraryMode` 仅影响扫描分派与过滤；默认保持现有行为。
- 图片集沿用 Gallery，不新增表；Gallery NFO 解析为新增包。
- 附件当前**不入库、不索引**；候选：
  - A 文件系统直读（零冻结，代价是每次扫目录）
  - B custom fields 存 JSON 字符串（零冻结，属 hack）
  - C 新建 `scene_attachments` 表（碰冻结区，可搜索/过滤/加元数据）
- 详见 `docs/DESIGN_MEDIA_SCAN.md`。

重评时机：

- 需要例 2（多文件夹 = 一集）→ Gallery 分组或 ImageSet。
- 附件需要元数据/搜索/统计 → 选方案 C。
- 需要音频等其他模式 → 扩展 flags。

参考：`docs/DESIGN_MEDIA_SCAN.md`（草案）。