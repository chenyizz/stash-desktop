# 健康度基线扫描

> 状态：基线快照（2026-09-24）。阶段 3 后会变化，重扫而非改旧。
> 方法：`go list` 依赖分析 + `Select-String`（本机无 `rg`）。只读扫描，未改动代码。
> 范围：`backend/`、`internal/`、`main.go`（不含 `frontend/`、`docs/`）。

## 1. 死代码（包级）

`go list` 直接导入统计（PROD = 非测试导入者数，TEST = 测试导入者数）：

| 包 | PROD | TEST | 状态 | 优先级 | 处理时机 | 处理方式 |
|---|---|---|---|---|---|---|
| `plugin/util` | 0 | 0 | 零引用（仅 `examples/` 的 `//go:build plugin_example` 引用） | P1 | 阶段 4 | 插件桥定型时：保留为示例依赖，或删 util+examples |
| `savedfilter` | 1 | 0 | 仅 `manager` 引用（导入/导出用） | P1 | 阶段 3-4 | 评估是否暴露 SavedFilter UI；否则随停用概念阶段 7 删 |
| `group` | 1 | 0 | 仅 `manager` 引用，UI 未暴露 | P1 | 阶段 3 | 决定 Group 是否进核心业务；不进则阶段 7 删 |
| `scraper` | 1 | 0 | 仅 `manager` 引用；AGENTS 标注「暂留不再使用」 | P1 | 阶段 4 | 插件桥设计时决定：改造为插件类型或删除 |
| `match` | 1 | 0 | 仅 `scraper` 引用（连锁死代码） | P1 | 阶段 4 | 随 scraper 一起评估 |
| `signedurl` | 1 | 0 | 仅 `ffmpeg` 引用，且仅用常量（`CIDParam` 等）；桌面端无流签名需求 | P1 | 阶段 3 | 抽查流媒体功能是否保留；不保留则阶段 7 删 |
| `session` | 2 | 0 | `plugin` + `manager`；本地 Store 有创建，主体为插件上下文 | P3 | 阶段 4 | 随插件系统保留；本地残留项阶段 7 清 |
| `javascript` | 2 | 0 | `plugin` + `scraper` | P3 | 阶段 4 | 随插件保留（goja JS 插件） |
| `python` | 2 | 0 | `plugin` + `scraper` | P3 | 阶段 4 | 随插件保留（Python 插件桥） |
| `plugin` | 4 | 3 | 多域引用（scene/gallery/image/manager） | P3 | 阶段 4 | 保留（ADR/AGENTS 已是规划功能） |
| `exec`、`file/image`、`gallery`、`performer`、`pkg`、`signedurl` 等其余 | ≥1 | — | 正常 | P3 | — | 保留 |

已确认**目录不存在**：`stashbox`、`identify`、`dlna`；`autotag` 非独立包（`scraper/autotag.go`）。

### 重点包符号级（导出但仓库内零外部引用，含仅包内使用/API 面）

| 包 | 导出符号数 | 零外部引用 | 备注 |
|---|---|---|---|
| `scraper` | 29 | 17（`Definition`、`GlobalConfig`、`ScrapedContent`、`ScraperSpec`、`QueryType` 等） | 大部分为 scraper 配置/类型，scraper 停用后自然死亡 |
| `plugin` | 46 | 27（`PluginInput/Output`、`RPCRunner`、`ServePlugin` 等） | 阶段 4 使用，勿删 |
| `session` | 11 | 6（`Get/SetCurrentUserID`、`IsLocalRequest` 等） | 桌面端无 HTTP 会话，疑似残留 |
| `match` | 17 | 6 | 连锁死代码（依赖 scraper） |
| `signedurl` | 4 | 4（`SignPrefix`/`VerifyURL`/`DerivePrefix`/`GenerateCredentialID`） | 仅常量被 `ffmpeg` 使用 |
| `group` | 17 | 8 | group 功能未暴露 |
| `javascript` | 10 | 3 | 随插件保留 |
| `savedfilter`、`python` | — | 0 | 无零引用符号 |

## 2. 数据库表

来源：`backend/pkg/sqlite` 的 `*Table = "..."` 与 `tableName:` 常量（共 43 + `*_stash_ids` 5 + `*_custom_fields` 8）。

| 表 | 状态 | 优先级 | 处理时机 | 处理方式 |
|---|---|---|---|---|
| `scenes`、`scene_urls`、`scenes_files`、`scenes_tags`、`scenes_galleries`、`scenes_view_dates`、`scenes_o_dates` | 在用 | P3 | — | 保留 |
| `performers`、`performer_aliases`、`performer_urls`、`tags`、`tags_relations`、`tag_aliases`、`studios`、`studio_aliases`、`studio_urls`、各 `*_tags`、`*_files` | 在用 | P3 | — | 保留 |
| `files`、`files_fingerprints`、`video_files`、`image_files`、`blobs` | 在用 | P3 | — | 保留 |
| `galleries`、`galleries_files`、`galleries_images`、`galleries_tags`、`gallery_urls` | 在用（图片集方向） | P3 | 阶段 4/5 | 保留（DESIGN_MEDIA_SCAN 复用 Gallery） |
| `saved_filters` | 未暴露（`savedfilter` 仅导入导出） | P2 | 阶段 7 | 只记录；删表需新迁移 |
| `groups`、`groups_scenes`、`groups_relations`、`groups_tags`、`group_urls` | 停用/未暴露 | P2 | 阶段 7 | 只记录；删表需新迁移 |
| `scene_markers`、`scene_markers_tags` | 停用/未暴露 | P2 | 阶段 7 | 同上（markers 生成任务仍在，需先评估） |
| `galleries_chapters`、`video_captions` | 低优先 | P2 | 阶段 7 | 评估后处理 |
| `scene_stash_ids`、`performer_stash_ids`、`studio_stash_ids`、`tag_stash_ids` | 领域术语（stash-box 外部 ID） | P2 | 阶段 7 | 与 `StashID` 评估项一起处理 |
| `*_custom_fields` × 8 | 在用（NFO/附件预留） | P3 | — | 保留 |

> 注：删表属迁移（冻结区），一律 P2 阶段 7，本报告不改。

## 3. 硬规则合规

| 项 | PROD | TEST | 状态 | 优先级 | 处理时机 | 处理方式 |
|---|---|---|---|---|---|---|
| `fmt.Printf/Println/Print` | 0 | 0（仅 `logger.go` 注释提及） | 合规 | P3 | — | 保留 |
| 标准库 `log.*` | 1（`main.go:43 log.Fatal`） | 0 | 启动入口，logger 未就绪 | P3 | — | 保留并注明 |
| `os.Exit` | 3（`logger.go:179,189` 为 Fatal 实现；`logger.go:13` 注释） | 1 | 合理 | P3 | — | 保留 |
| `context.TODO()` | 19 | 0 | 惯性占位 | P1 | 阶段 5 | 记 TECH_DEBT；阶段 5 用请求 ctx 替换 |
| `context.Background()` | 17 | 59 | 大部分合理（启动/后台任务） | P3 | — | 保留；必要时阶段 5 收敛 |
| `_ = err` | 0 | 0 | 合规 | P3 | — | — |
| `panic(` | 51 | 7 | 多为 `mustLoaded`/配置不变量/JS VM | P3 | — | 保留；`mustLoaded` 系列属设计约束 |

## 4. 注释

| 项 | 数量 | 文件数 | 优先级 | 处理时机 | 处理方式 |
|---|---|---|---|---|---|
| `TODO` 注释（排除 `context.TODO`） | 132 | 60 | P2 | 阶段 7 | 只记录；不逐个处理 |
| `FIXME` | 0 | 0 | — | — | — |
| `XXX` | 3 | — | P3 | — | 抽查 |
| `deprecated` 标记 | 58 | — | P2 | 阶段 7 | 随停用概念一起清 |
| `#issue` 注释 | 73 | 47 | P3 | — | **禁止删除**（AGENTS 硬规则），随对应代码处理 |

TODO 集中在 `backend/manager`（生成任务/任务返回错误）与 `backend/pkg/scraper`；`#issue` 集中在 manager 扫描/生成与 sqlite 过滤。

## 5. 命名残留

| 项 | 数量 | 分类 | 优先级 | 处理时机 | 处理方式 |
|---|---|---|---|---|---|
| 小写 `stash` | 106 | 领域 `stash_ids`(7)、配置残留(4)、测试夹具/注释/文案(95) | P2 | 5c + 阶段 7 | 5c 处理 `stashignore`；其余注释/夹具阶段 7 批量清 |
| 大写 `Stash` | 9 | `StashBox`(4)、注释(4)、`main.go:20` 描述(1) | P0/P2 | 立即/阶段 7 | `main.go:20` 描述改掉；注释阶段 7 |
| `StashConfig/StashPaths/GetStashFrom/GetStashHomeDirectory` | 0 | — | — | 5a/5b 已完成 | 已清零 |

## 建议的 P0 立即项

| 项 | 位置 | 方式 | 风险 |
|---|---|---|---|
| 用户可见描述含 Stash | `main.go:20` `Description: "...powered by Stash"` | 改为 `...powered by Case` 或精简 | 无（字符串） |
| `plugin/util` 零引用 | `backend/pkg/plugin/util/client.go` + `examples/` | 建议**不删**（示例依赖），降级 P1 | 删会破坏 build-tag 示例 |

> 其余 P1/P2/P3 均建议只记录，不动代码；按上表时机处理。

## 重扫建议

- 阶段 3 完成（分页/搜索/标签页）后重扫一次：死代码与硬规则计数会显著变化。
- 阶段 4 插件桥定型后重扫 Part 1，决定 `scraper`/`match`/`signedurl` 去留。
- 阶段 7 前重扫 Part 2，制定删表迁移清单。
