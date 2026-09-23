# internal/app

## 模块职责

Wails 服务层：只做胶水——初始化目录/日志/配置/Manager，向绑定暴露方法，并提供封面中间件。**不写业务逻辑**。

## 文件列表

| 文件 | 说明 |
| --- | --- |
| `app.go` | App 服务主体：`ServiceStartup`/`Shutdown`、`ScanLibrary`/`FindScenes`、列表 DTO `SceneDTO` |
| `scene.go` | `GetScene` + `SceneDetailDTO`/`TagDTO`/`PerformerDTO`/`SceneFileDTO` 与纯映射 |
| `cover.go` | `ensureCover`（按需生成并持久化封面）+ `CoverMiddleware`（`/covers/<id>`） |
| `config.go` | `setupConfig`：设置 `CASE_CONFIG_FILE`、初始化默认配置 |
| `logging.go` | `setupLogging`：文件 Handler + UI Handler |
| `paths.go` | 数据目录策略：`ResolveLayout` / `ExeDir` / `Layout` |
| `wails_emitter.go` | 日志推送到前端的适配器 |
| `scene_test.go` | `toSceneDetailDTO` 纯映射单测 |

## 导出 API

**Wails 绑定方法**（首字母大写才会暴露给前端）：

| 方法 | 签名 | 用途 |
| --- | --- | --- |
| `New` | `func New() *App` | 构造服务 |
| `SetApplication` | `func (a *App) SetApplication(wailsApp *application.App)` | 注入 Wails 应用（日志推送用） |
| `ServiceStartup` | `func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error` | 初始化布局/日志/config/Manager |
| `ServiceShutdown` | `func (a *App) ServiceShutdown() error` | 关闭 Manager |
| `GetDataDir` | `func (a *App) GetDataDir() (string, error)` | 数据目录 |
| `GetVersion` | `func (a *App) GetVersion() string` | 版本号 |
| `Ping` | `func (a *App) Ping() string` | 连通性测试 |
| `GetSystemStatus` | `func (a *App) GetSystemStatus() map[string]any` | 状态查询 |
| `ScanLibrary` | `func (a *App) ScanLibrary(path string) (int, error)` | 触发扫描，返回 Job ID |
| `FindScenes` | `func (a *App) FindScenes(page int, pageSize int) ([]SceneDTO, error)` | 场景列表（分页） |
| `GetScene` | `func (a *App) GetScene(id int) (*SceneDetailDTO, error)` | 场景详情（含封面字段） |

**包级导出（不绑定给前端）**：

| 函数/类型 | 签名 | 用途 |
| --- | --- | --- |
| `CoverMiddleware` | `func CoverMiddleware(a *App) application.Middleware` | 服务 `/covers/<id>`（在 `main.go` 注册） |
| `ResolveLayout` | `func ResolveLayout() (*Layout, error)` | 解析数据目录布局 |
| `ExeDir` | `func ExeDir() string` | exe 所在目录 |
| `Layout` | `type Layout struct { BaseDir, DataDir, LogDir string }` | 目录布局 |

**DTO**：`SceneDTO`（列表）、`SceneDetailDTO`（详情）、`TagDTO`、`PerformerDTO`、`SceneFileDTO`。

## 使用示例

```ts
// 前端：通过自动生成的 bindings 调用
import { FindScenes, GetScene, ScanLibrary } from "../bindings/case/internal/app";

const jobID = await ScanLibrary("E:\\test");
const scenes = (await FindScenes(1, 50)) ?? [];
const detail = await GetScene(scenes[0].id);
```

```go
// Go：目录布局（不依赖 Manager，可在启动早期使用）
layout, err := app.ResolveLayout()
if err != nil {
    return err
}
logger.Infof("data dir: %s", layout.DataDir)
```

## 前置条件

- 必须先 `SetApplication` + `ServiceStartup` 完成；未初始化时方法返回「manager 未初始化」。
- 前端只依赖 `frontend/bindings/case/internal/app`，不直接依赖 `backend/pkg`。
- 新增绑定方法必须导出（首字母大写）并重新生成 bindings（`wails3 dev` / `wails3 generate bindings`）。

## 已知限制

| 限制 | 说明 | 偿还时机 |
| --- | --- | --- |
| 平铺结构将膨胀 | 阶段 3 预计新增 6+ 文件，不定义会变上帝包 | 阶段 3 前定 `dto/`/`service/`/`middleware/`/`infra/` |
| 封面同步生成 | `GetScene` 内 `ensureCover` 阻塞 100–500ms | 阶段 5：改异步 |
| 列表无真正分页语义 | `FindScenes` 仅透传 `page/pageSize` | 阶段 3 |
| 详情无附件 | 附件机制待决策（DESIGN D2） | 阶段 3 |
| 详情数据用本地 `$state` | ADR-007 的临时方案 | 阶段 5：重评 TanStack Query |

详见 `docs/TECH_DEBT.md`、`docs/ARCHITECTURE.md`（ADR-006/007）。

## 依赖关系

本包依赖：

- `backend/manager`、`backend/manager/config`
- `backend/pkg/logger`、`backend/pkg/models`、`backend/pkg/utils`
- `backend/pkg/metadata/cover`、`backend/pkg/scene/generate`
- `github.com/wailsapp/wails/v3/pkg/application`

谁依赖本包：

- `main.go`：构造 `app.New()`、注册服务与 `CoverMiddleware`
- 前端：`frontend/bindings/case/internal/app`（自动生成，非手写）
