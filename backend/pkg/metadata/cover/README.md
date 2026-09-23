# metadata/cover

## 模块职责

从本地来源按优先级解析场景封面（`poster` → 同名图 → FFmpeg 截帧），返回图片字节；**不负责持久化**。

## 文件列表

| 文件 | 说明 |
| --- | --- |
| `cover.go` | 策略链 `sources` 与各来源实现：`fromPoster` / `fromSameName` / `fromFFmpeg` |
| `cover_test.go` | 表驱动测试 + 真实 FFmpeg E2E（`TestResolve_FFmpegRealFrame`，无 ffmpeg 时 skip） |

## 导出 API

| 函数/类型 | 签名 | 用途 |
| --- | --- | --- |
| `Resolve` | `func Resolve(ctx context.Context, videoPath string, duration float64, screenshot ScreenshotFunc) (Result, error)` | 按优先级解析封面；首个成功来源即返回 |
| `Result` | `type Result struct { Data []byte; Source string }` | 图片字节与来源名（`poster`/`same-name`/`ffmpeg`） |
| `ScreenshotFunc` | `type ScreenshotFunc func(ctx context.Context, videoPath string, at float64) ([]byte, error)` | 注入的截帧实现（FFmpeg 来源用） |
| `DefaultTimeProportion` | `const DefaultTimeProportion = 0.2` | 截帧位置（时长的 20%） |

来源优先级：`poster.jpg|jpeg|png`（同目录）→ `<视频同名>.jpg|jpeg|png` → FFmpeg 20% 截帧。

## 使用示例

```go
// 方式 1：只用本地图片，不需要 FFmpeg（screenshot 传 nil）
result, err := cover.Resolve(ctx, videoPath, 0, nil)
if err != nil {
    logger.Warnf("resolve cover: %v", err)
    return
}
if len(result.Data) > 0 {
    // result.Data 交给调用方持久化（例如 Scene.UpdateCover）
    logger.Infof("cover from %s (%d bytes)", result.Source, len(result.Data))
}
```

```go
// 方式 2（推荐）：带上 FFmpeg 截帧兜底
result, err := cover.Resolve(ctx, videoPath, duration, screenshotFunc)
if err != nil {
    logger.Warnf("resolve cover: %v", err)
    return
}
if len(result.Data) == 0 {
    // 无封面（非错误），调用方走占位
    return
}
```

## 前置条件

- `videoPath` 非空；本地图片来源要求文件可读且能被 `image.DecodeConfig` 解码（非图片内容会被忽略）。
- FFmpeg 来源要求 `duration > 0` 且 `screenshot` 非 nil，否则自动跳过。
- `Resolve` 返回空 `Result` 表示「没有可用封面」，不是错误。
- 本包不持久化、不生成缩略图；持久化与 DTO 由调用方负责。

## 已知限制

| 限制 | 说明 | 偿还时机 |
| --- | --- | --- |
| 同步生成 | FFmpeg 来源在详情页请求内同步执行，首次约 100–500ms | 阶段 5：改异步 |
| 无缩略图 | 列表海报墙若加载原图会卡 | 阶段 3/5 |
| 失败无缓存 | FFmpeg 失败后每次重试 | 遇频繁失败时 |
| 无内存缓存 | 每次重新读盘/生成 | 阶段 5：加 LRU |
| 无并发去重 | 同一场景并发打开可能生成两次 | 遇问题再修 |
| 格式范围 | 本地图片仅 `jpg/jpeg/png` | 有需要时扩展 |

详见 `docs/TECH_DEBT.md`。

## 依赖关系

本包依赖：

- `backend/pkg/logger`
- 标准库（`image`、`net/http` 内容类型嗅探等）

谁依赖本包：

- `internal/app/cover.go`：`ensureCover` 调用 `Resolve`，并注入基于 `scene/generate` 的 `ScreenshotFunc`
