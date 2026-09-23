# 代码约定

## Go 侧

- `backend/manager/`：总协调，不做具体工作
- `backend/pkg/<domain>/`：业务域实现
- `internal/app/`：Wails 服务层，只做胶水
- 导出方法用 PascalCase
- 未导出方法用 camelCase
- 错误用 `fmt.Errorf("上下文: %w", err)`
- 日志用 `logger.Errorf("...: %v", err)`
- 所有查询走 `WithReadTxn` / `WithTxn`

## 前端侧

- `src/App.svelte`：主界面
- `src/lib/components/`：组件
- `src/lib/stores/`：Zustand
- `src/lib/hooks/`：TanStack Query
- 从 `../bindings/case/internal/app` 导入
- 所有调用 `await`
- 用 `?? []` 处理 null

## 分层规则

正确链路：

```text
Svelte 前端
  -> Wails bindings
  -> internal/app 导出方法
  -> backend/manager
  -> backend/pkg
  -> SQLite / 文件系统
```

前端不直接依赖 `backend/pkg`。  
前端只认 `internal/app` 返回的 DTO。

## internal/app 目标结构

阶段 3 按新结构放置新增文件；已有文件在阶段 3 适当时机迁移，不强制一次完成。

```text
internal/app/
├─ dto/          ← DTO 定义（SceneDetailDTO、TagDTO、PerformerDTO 等）
├─ service/      ← 领域服务（scene.go、performer.go、tag.go、studio.go）
├─ middleware/   ← AssetServer 中间件（CoverMiddleware 等）
├─ infra/        ← 基础设施（config.go、logging.go、paths.go、wails_emitter.go）
└─ app.go        ← ServiceStartup + Wails 方法注册
```

文件迁移规则：见 `docs/FROZEN_RULES.md`（移动已有文件需 Plan + 人工确认）。

## 状态管理

- 客户端状态：Zustand
- 服务端数据：TanStack Query
- 不要混用

## 日志

- Go：`logger.Infof` / `Debugf` / `Warnf` / `Errorf`
- 插件：stderr 输出 JSON
- 前端：监听 `app:log`