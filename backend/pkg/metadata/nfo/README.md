# metadata/nfo

## 模块职责

解析本地 NFO（Kodi `<movie>` schema），映射为场景元数据，并在扫描后回填到 Scene。

## 文件列表

| 文件 | 说明 |
| --- | --- |
| `movie.go` | NFO `<movie>` 的 XML DTO（`Movie` / `Actor` / `Set`） |
| `nfo.go` | 安全解析入口：`Parse` / `ParseFile` / `FindForVideo`（Strict、禁 DTD、10 MiB、UTF-8 BOM） |
| `mapping.go` | `Movie → SceneMetadata` 纯映射（日期回退、评分换算、名字 NFKC 去重） |
| `applier.go` | `Applier.Apply`：扫描后读 NFO 并回填 Scene（只填不覆盖） |
| `nfo_test.go` / `mapping_test.go` / `applier_test.go` | 表驱动测试与样例 |
| `testdata/FDD-2002.nfo` | 样例 NFO |

## 导出 API

| 函数/类型 | 签名 | 用途 |
| --- | --- | --- |
| `Parse` | `func Parse(r io.Reader) (*Movie, error)` | 从 reader 解析 NFO |
| `ParseFile` | `func ParseFile(path string) (*Movie, error)` | 读取并解析 NFO 文件 |
| `FindForVideo` | `func FindForVideo(videoPath string) (string, bool)` | 查找同名 `.nfo` sidecar |
| `Movie` | `type Movie struct` | Kodi `<movie>` DTO（`Title`/`Num`/`Premiered`/`Actors`/`Tags`/…） |
| `Actor` / `Set` | `type Actor struct` / `type Set struct` | `<actor>` / `<set>` 条目 |
| `SceneMetadata` | `type SceneMetadata struct` | `Title/Code/Details/Director/Date/ProductionDate/Rating/StudioName/PerformerNames/TagNames/GroupNames/URLs/CustomFields` |
| `(*Movie).SceneMetadata` | `func (m *Movie) SceneMetadata() SceneMetadata` | 映射为场景元数据 |
| `Applier` | `type Applier struct { Repo models.Repository }` | 回填器 |
| `(*Applier).Apply` | `func (a *Applier) Apply(ctx context.Context, sceneID int, videoPath string) error` | 事务外解析、单事务回填 Scene |
| `MaxFileSize` | `const MaxFileSize = 10 << 20` | 文件大小上限（10 MiB） |

## 使用示例

```go
movie, err := nfo.ParseFile(filepath.Join(dir, "FDD-2002.nfo"))
if err != nil {
    logger.Warnf("skip nfo: %v", err)
    return
}
meta := movie.SceneMetadata()

applier := &nfo.Applier{Repo: mgr.Repository}
if err := applier.Apply(context.Background(), sceneID, videoPath); err != nil {
    logger.Errorf("apply nfo: %v", err)
}
_ = meta
```

## 前置条件

- 输入须为 UTF-8（可带 BOM）；仅识别 `<movie>` 根节点。
- `Applier.Apply` 需有效的 `models.Repository`；须在文件事务提交后调用（当前挂在 `scene.ScanHandler` 的 post-commit hook）。
- `Movie.SceneMetadata` 无副作用（仅对不可解析日期 `Warnf`）；不要求先加载任何关联。

## 已知限制

| 限制 | 说明 | 偿还时机 |
| --- | --- | --- |
| 仅 UTF-8 + BOM | 无 Shift-JIS/GBK 回退，遇到旧编码 NFO 会整份失败 | 阶段 3：遇到日文 NFO 时评估 |
| `Strict=true` | 未转义 `&` 会导致整份失败 | 阶段 3：遇到脏 NFO 时评估 |
| `Parse(io.Reader)` 无大小保护 | 仅 `ParseFile` 有 10 MiB 上限 | 阶段 5：接入外部数据源前 |
| 只填不覆盖 | 同场景多文件时为「先到先填」，无确定性优先级 | 阶段 3+：需要确定性时设计 |
| `CustomFields` 只映射不落库 | mpaa/runtime 等进入 `SceneMetadata.CustomFields`，Applier 暂不写入 | 阶段 3：详情页需要展示时 |

详见 `docs/TECH_DEBT.md` 与 `docs/HEALTH_SCAN.md`。

## 依赖关系

本包依赖：

- `backend/pkg/models`（`Date` / `ScenePartial` 等）
- `backend/pkg/logger`
- `golang.org/x/text/unicode/norm`、标准库（`encoding/xml` 等）

谁依赖本包：

- `backend/manager/task_scan.go`：构造 `nfo.Applier` 注入 `scene.ScanHandler.MetadataApplier`
- `backend/pkg/scene`：仅通过 `ScanMetadataApplier` 接口调用，无编译期依赖
