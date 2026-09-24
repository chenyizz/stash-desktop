# 技术债

记录已知但未处理的取舍与缺口。按类别归档，不要求当前处理。

> 媒体扫描模式 / 附件机制 / 图片集的设计与待决策项见 `docs/DESIGN_MEDIA_SCAN.md`。
> 健康度基线扫描（死代码/表/硬规则/注释/命名）见 `docs/HEALTH_SCAN.md`。

## 偿还时机清单（来自 HEALTH_SCAN.md）

只列有明确偿还时机的条目；其余见 `docs/HEALTH_SCAN.md`。

| 项 | 偿还时机 |
|---|---|
| `internal/app` 结构迁移（`dto/`/`service/`/`middleware/`/`infra/`）未执行，当前平铺 | 阶段 3 |
| `context.TODO()` 19 处（冻结区 `manager`/`ffmpeg`） | 阶段 5 |
| `signedurl` 函数级零引用（仅常量被 `ffmpeg` 用） | 阶段 3 抽查流媒体功能 |
| `group` / `savedfilter` UI 未暴露 | 阶段 3 决定是否进核心业务 |
| `scraper` / `match` 待评估（AGENTS：暂留不再使用） | 阶段 4 插件桥设计时 |
| `plugin/util` 零引用（仅 build-tag 示例依赖） | 阶段 4 |
| `session` 桌面端残留（`Get/SetCurrentUserID` 等） | 阶段 7 |
| 数据库停用表（`saved_filters`、`groups*`、`scene_markers*`、`galleries_chapters`、`video_captions`） | 阶段 7（删表需新迁移） |
| TODO 注释 132 处 / `deprecated` 标记 58 处 | 阶段 7 |
| 小写 `stash` 106 处（多为注释/测试夹具） | 阶段 7 |

相关决策见 `docs/ARCHITECTURE.md` ADR-009。

## 功能性缺口

- 附件机制未定：候选 A 文件系统直读 / B custom fields JSON 字符串 / C `scene_attachments` 表，阶段 3 详情页附件功能实现时决策（详见 `docs/DESIGN_MEDIA_SCAN.md` D2）。当前不做任何附件入库或索引。
- 同场景多文件首次扫描时，标量字段由并行扫描完成顺序决定，后续扫描不改变。如需确定性优先级，阶段 3+ 再设计。
- NFO 编码只支持 UTF-8 + BOM，无 Shift-JIS/GBK 回退。
- NFO 解析 `Strict=true`，未转义 `&` 会导致整份失败。
- `nfo.Parse(io.Reader)` 无文件大小保护，仅 `ParseFile` 有 10 MiB 上限。

## 性能性取舍（有意为之）

- 封面比例适配策略待定：列表用 `16/9` + `object-fit: cover`；若大量竖版 poster 被裁得难看，再考虑动态比例或 `object-fit: contain`。阶段 3+ 评估。
- 缩略图缓存：磁盘上限 256MB + 启动孤儿清理；场景删除/封面更新后的旧缩略图仅在启动清理，**精确随删随清需 hook blob 删除（冻结区）**，阶段 5/7。
- 缩略图同步生成：`/covers/<id>?w=` 首次请求内用 imaging 同步缩放。
- 搜索基于 `LIKE`（无 FTS）；场景搜索新增 tag/演员/工作室 join 仅在有查询时生效，大数据量下成本上升 → 阶段 5 评估 FTS/索引。
- 过滤不持久化（刷新/切换页签重置）；仅覆盖已整理/评分/日期/tag/演员/工作室，其余维度（分辨率/编码/时长/收藏/marker/is_missing）按需再补。含子标签/子工作室为显式勾选（Depth=-1）。
- 附件（方案 A）：无缓存、每次详情请求枚举目录并 `DecodeConfig`；如未来需要搜索/过滤/元数据，升级为方案 C（`scene_attachments` 表，冻结区迁移）。
- `LibraryMode` 目前仅配置文件（无 UI）：切换扫描模式需手改 `config.yml`；UI 与迁移（旧配置补 `mode`）留待后续。
- 封面同步生成：首次打开详情阻塞 100–500ms，阶段 5 考虑异步。
- 无缩略图：列表海报墙加载原图会卡，阶段 3/5 补。
- 封面无失败缓存：FFmpeg 失败每次重试，遇频繁失败时补。
- 封面无内存缓存：每次读 blob 文件，阶段 5 加 LRU。
- 封面无并发去重：同一场景并发打开可能生成两次，遇问题再修。
- 无分页、无虚拟滚动：列表数据量大时卡，阶段 3 补。
- 扫描完成用 `setTimeout` 等，不可靠，阶段 3 改为事件通知。

## 展示策略（前端不暴露）

- Performer 隐藏字段（DTO/UI 不暴露，DB 保留，阶段 7 再决定删否）：`Ethnicity`、`HairColor`、`EyeColor`、`PenisLength`、`Circumcised`、`FakeTits`、`Piercings`、`StashIDs`。
- `StashIDs` 模块：内部领域术语保留（ADR-009），前端 DTO/UI 一律不暴露。
- 标签 / 工作室详情页未做（本阶段只做演员详情）；若需要按同模式补 `GetTag/GetStudio` 与详情组件。

## 一致性 / 命名

- `JavbusID` 字段名与内容不匹配（存的是 URL）。
- `directorFromActors` 做归一化，`m.Director` 不做，风格不一致。
- `Resolution` 用 `min(width,height)+"p"` 派生，非标准（1920×800 → "800p"）。
- `Rating` nil → 0、`DurationFinite` Inf/NaN → 0，均无法区分「无值」与「零值」。

## 测试缺口（已校正）

- 已覆盖：`FindForVideo` 大写扩展名、`ParseFile` 目录报错、`parseRating "85"`、全角/半角去重（2.4.1 修复时补测）。
- 仍缺：`ParseFile` 超大文件、Actor Type 小写输入、封面策略链之外的 E2E。
- 无 E2E 测试，冒烟靠手工 `wails3 dev`。
- 无 CI/CD，验证链手动执行。

## 工程化

- 无 `.gitattributes`，LF/CRLF 转换不可控（git 每次提示）。
- 数据库迁移已用 `golang-migrate`（`backend/pkg/sqlite/migrations`），但迁移为 Go 文件混合、缺少统一 SQL 与回滚说明。
- Wails v3.0.0-beta.23 是 beta 版，上游可能有破坏性变更。
- 冻结区策略导致 `backend/pkg` 长期无法重构，阶段 7 处理。
- `internal/app` 可能随功能增加膨胀，需在阶段 3 前评估拆分。

## 静态检查（go vet）

`go vet ./backend/manager/...` 有 5 条既有告警（均在冻结区 `backend/manager`），暂不修，待单独评估：

- `generator_interactive_heatmap_speed.go:132`：`%d` 用于 `float64`（应 `%f`）
- `task_generate.go:245`：`logger.Infof(logMsg)` 非常量格式串（应 `logger.Infof("%s", logMsg)`）
- `task_generate_clip_preview.go:43`：`logger.Errorf` 使用 `%w`（应 `%v`）
- `task_scan.go:112` / `task_scan.go:290`：`logger.Errorf(string(debug.Stack()))` 非常量格式串（应 `logger.Errorf("%s", debug.Stack())`）

影响：`go test ./backend/manager/...` 默认因 vet 失败，需加 `-vet=off`。

## 命名遗留（Stash）

- **评估项：`StashID` / `StashIDs` / `stash_id`**（1300+ 处，60+ 文件）。当前视为领域术语（stash-box 的外部 ID），**暂不改**。如未来接入其他元数据源，考虑改为 `ExternalID`，届时与 scraper 子系统一起做。
- **长期项：数据库表名/列名** `scene_stash_ids`、`tag_stash_ids`、`performer_stash_ids`、`studio_stash_ids`、`group_stash_ids` 及列 `stash_id` / `endpoint`。暂不改（迁移风险），随 `StashID` 评估项一并处理。
- `StashBox` / `StashBoxInput` / `GetStashBoxes` 保留（stash-box 协议专有名词）。
- 命名清理已分批完成：第一级用户可见字符串、第二级配置键/环境变量（含 `CASE_HW_TEST_TIMEOUT`、`CASE_HW_DRI_DEVICE`、`CASE_SQLITE_CACHE_SIZE`）、第三级 `StashConfig→LibraryConfig`、`GetStashHomeDirectory→GetCaseHomeDirectory`、`stashignore→caseignore`（磁盘文件 `.caseignore`，未保留旧 `.stashignore` 兼容）。
