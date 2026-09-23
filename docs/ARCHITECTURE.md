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