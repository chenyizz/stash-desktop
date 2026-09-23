# 技术债

记录已知但未处理的取舍与缺口。按类别归档，不要求当前处理。

> 媒体扫描模式 / 附件机制 / 图片集的设计与待决策项见 `docs/DESIGN_MEDIA_SCAN.md`。

## 功能性缺口

- 附件机制未定：候选 A 文件系统直读 / B custom fields JSON 字符串 / C `scene_attachments` 表，阶段 3 详情页附件功能实现时决策（详见 `docs/DESIGN_MEDIA_SCAN.md` D2）。当前不做任何附件入库或索引。
- 同场景多文件首次扫描时，标量字段由并行扫描完成顺序决定，后续扫描不改变。如需确定性优先级，阶段 3+ 再设计。
- NFO 编码只支持 UTF-8 + BOM，无 Shift-JIS/GBK 回退。
- NFO 解析 `Strict=true`，未转义 `&` 会导致整份失败。
- `nfo.Parse(io.Reader)` 无文件大小保护，仅 `ParseFile` 有 10 MiB 上限。

## 性能性取舍（有意为之）

- 封面同步生成：首次打开详情阻塞 100–500ms，阶段 5 考虑异步。
- 无缩略图：列表海报墙加载原图会卡，阶段 3/5 补。
- 封面无失败缓存：FFmpeg 失败每次重试，遇频繁失败时补。
- 封面无内存缓存：每次读 blob 文件，阶段 5 加 LRU。
- 封面无并发去重：同一场景并发打开可能生成两次，遇问题再修。
- 无分页、无虚拟滚动：列表数据量大时卡，阶段 3 补。
- 扫描完成用 `setTimeout` 等，不可靠，阶段 3 改为事件通知。

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

## 命名遗留（Stash）

- **评估项：`StashID` / `StashIDs` / `stash_id`**（1300+ 处，60+ 文件）。当前视为领域术语（stash-box 的外部 ID），**暂不改**。如未来接入其他元数据源，考虑改为 `ExternalID`，届时与 scraper 子系统一起做。
- **长期项：数据库表名/列名** `scene_stash_ids`、`tag_stash_ids`、`performer_stash_ids`、`studio_stash_ids`、`group_stash_ids` 及列 `stash_id` / `endpoint`。暂不改（迁移风险），随 `StashID` 评估项一并处理。
- `StashBox` / `StashBoxInput` / `GetStashBoxes` 保留（stash-box 协议专有名词）。
- **环境变量保留旧名**：`STASH_HW_TEST_TIMEOUT`、`STASH_HW_DRI_DEVICE`（`backend/pkg/ffmpeg`）、`STASH_SQLITE_CACHE_SIZE`（`backend/pkg/sqlite`）仍是旧名，位于冻结区，单独做太碎——与 Step 5 第三级标识符清理一起处理（届时加 `CASE_` 回退）。
- **兼容旧名代码（过渡期）**：`STASH_CONFIG_FILE`、`STASH_` 前缀、`~/.stash`、`stash-go.sqlite`、配置键 `stash` 的兼容读取逻辑，属于过渡期代码，待用户升级完成后删除。偿还时机：v1.0 发布后一个版本周期。
- **默认数据库名 `stash-go.sqlite` → `case.db`**：若存在存量 `stash-go.sqlite`，当前会被忽略（不会自动迁移）。如需兼容，首次启动时检测旧文件并提示用户。偿还时机：有真实用户存量时。
- 命名清理分批计划：第一级用户可见字符串（立改）；第二级配置键/环境变量（兼容旧名）；第三级标识符（按包，`StashConfig→LibraryConfig` 与扫描模式 D1 同批、`stashignore→caseignore` 兼容旧文件）；第四级数据库表名（不改，见上）；第五级 blob 路径（无残留，不动）。
