# 媒体扫描模式设计

> 状态：草案（2026-09-23），实施前可能调整。实施后转为 ADR 引用（见 `docs/ARCHITECTURE.md` ADR-008）。

## 1. 需求

支持三类独立扫描模式：

- **模式 A 视频库**：只扫视频（`.mp4/.mkv` 等）；图片不作为独立 `Image` 实体入库；特定目录（`fanart/`、`poster/`、`extra/`）下的图片作为「附件」，在影片详情页展示，不参与列表页与搜索。
- **模式 B 图片库**：只扫图片（`.jpg/.png` 等）；**一个最小文件夹 = 一个图片集**（概念类似 scene）；图片集用 NFO 管理元数据，目录结构类似影片。
- **模式 C 混合库**：视频与图片各自独立扫描。

附加要求：附件展示；图片集列表页/详情页。

## 2. 原版对照

原版 Stash（本地 `backend/` 即 develop 迁移代码）：

- **扫描模式**：无全局「只视频/只图片」开关。仅有按库路径的 `StashConfig.ExcludeVideo/ExcludeImage`（`backend/manager/config/stash_config.go:12-13`），在 `backend/manager/task_scan.go:632-637` 生效；以及全局 `image_excludes`/video `excludes` 正则（`config.go:830`）。`ScanMetadataOptions` 只有生成开关。→ 只能用 exclude 组合近似模式，无显式语义。
- **scene 附件**：无 scene↔image 直接关联，无 fanart/extra 概念。唯一桥梁是 **scene↔gallery**（`scenes_galleries`，`Scene.GalleryIDs`/`Gallery.SceneIDs`），由同名 `.zip` 触发（`backend/pkg/scene/scan.go:191`）。场景封面是独立 blob 机制。
- **图片集**：有 **Gallery**（folder-based / zip-based），字段与 Scene 高度相似（`backend/pkg/models/model_gallery.go:10-39`），有封面正则 `(poster|cover|folder|board)\.[^\.]+$`（`config.go:213`），有导出/导入 JSON，但**没有 NFO**。

**缺口**：模式 A 的「附件」原版没有；模式 B 的「NFO 管理图片集」原版没有（只有 Gallery + JSON）。

## 3. 决策记录

### D1 库路径用双布尔 flags，不用三选一枚举

```go
type LibraryMode struct {
    Videos      bool // 是否扫描视频文件
    Images      bool // 是否扫描图片文件
    Attachments bool // 识别 fanart/extra/poster 作为附件
    // 将来可加 Audio bool / Subs bool 等
}
```

- 语义：只视频 `{Videos:true, Images:false}`；只图片 `{false,true}`；混合 `{true,true}`；都 false → **配置校验拒绝**（返回错误「至少启用一种媒体类型」）。
- 默认：`Videos=true, Images=true, Attachments=true`（等价现状，兼容现有库）。
- 理由：媒体类型当前只有视频/图片两种，双布尔覆盖全部组合；YAML 配置对普通用户直白；未来加音频/字幕是同一模式扩展。若将来媒体类型超过 5 种或需要互斥组合，再升级为枚举 + 集合（当前不做）。
- 影响：改 `backend/manager/config`（**冻结区**，需人工确认），并加一行校验。

**D1-a（已解决）**：模式 C（同时扫视频与图片）= `Videos && Images`，无需「同目录两条配置」。

### D2 附件机制：待决策项，本轮不定稿

附件不采用任何「现在就要欠债」的 hack。三种候选：

| 方案 | 成本 | 代价 | 收益 |
|---|---|---|---|
| A 文件系统直读 | 零冻结区改动 | 每次打开详情页扫目录 + `DecodeConfig`（50–200ms）；不可搜索/过滤/统计 | 无 |
| B custom fields 存 JSON 字符串 | 零冻结区改动 | custom fields 用户可编辑，需保留键命名空间；每次 `json.Unmarshal`；属 hack（`custom_fields.go:116-120` 明确不支持数组/对象） | 无 |
| C 新建 `scene_attachments` 表 | 新表 + 迁移 + Repository + DTO（**碰冻结区**） | 迁移与维护成本 | 可搜索、可过滤、可统计、可加元数据（描述/标签/顺序） |

**结论：已决策（阶段 3）**
> 附件机制采用**方案 A 文件系统直读**：扫描时/详情时识别 `<视频目录>/{fanart,poster,extra}/**`，由 `/attachments/<sceneID>/<index>` 按需提供，**不入库、不索引、不搜索**。
> 候选 B（custom fields JSON）与 C（`scene_attachments` 表）保留为将来扩展：当附件需要元数据/搜索/统计时升级到 C。
> `Attachments` 开关由 `LibraryMode`（D1）控制（阶段 3 落地）。

### D3 图片集复用 Gallery（例 1 落定）

- 确认按**例 1**：一个最小文件夹 = 一个图片集。
- 理由：需求原话「最小文件夹下的系列套图」，「最小文件夹」即例 1 的字面含义。
- 技术依据：`Gallery.FolderID` 强绑定一个文件夹（`sqlite/gallery.go:657 FindByFolderID`），例 1 完全契合。
- 例 2（多文件夹 = 一个图片集）：**记为未来扩展「Gallery 分组」**（给 Gallery 加 `GroupID`/parent 或新建 `ImageSet`），写入「不做」章节，不阻塞当前设计。
- 影响：新增 Gallery NFO 解析（新包），不新增表。

### D4 附件目录识别：约定 + 可配置正则

- 约定 `fanart/`、`poster/`、`extra/`；另加可配置正则兜底。
- 理由：开箱即用，兼顾自定义。

## 4. 数据模型

- **库路径**：`StashConfigInput/StashConfig` 增 `LibraryMode`（见 D1），纯扩字段，默认不改既有行为。
- **附件**：不入库。机制待 D2 决策；当前仅约定目录识别，不做入库/索引。
- **图片集**：复用 `Gallery`（folder-based，一文件夹一组），沿用 `FolderID`、封面、tags/performers/studio 关系；元数据来源新增 NFO。

## 5. 扫描流程

1. **按 mode 注册 handler**（`backend/manager/task_scan.go getScanHandlers`）：
   - 视频模式：不注册 `image.ScanHandler`（图片不入库），仅 video handler。
   - 图片模式：不注册 video handler，仅 image handler。
   - 混合（`Videos && Images`）：两者都注册。
2. **`scanFilter` 按 mode 过滤**：视频模式跳过 image/zip；图片模式跳过 video；尽量放在 manager 侧，减少改 `backend/pkg/file`。
3. **图片模式**：最小文件夹 → folder-based Gallery（该模式下等价强制 `createGalleriesFromFolders`，`.nogallery`/`.forcegallery` 仍生效）。
4. **Gallery NFO**：读取图片集 NFO 回填 `Gallery`（只填不覆盖，复用 2.4.2 Applier 思路）。

## 6. 前端展示

- **附件区块**（阶段 3 预留位，具体机制待 D2）：详情页展示；不参与搜索/列表。
- **图片集列表/详情页**（独立子步骤）：路由 `#/galleries`、`#/galleries/:id`，绑定 `FindGalleries` / `GetGallery`；放阶段 4+ 或阶段 5。

## 7. 当前状态

阶段 2 的「`FDD-2002.jpg` 独立入库」是 **Stash 原版默认行为**（image handler 无差别入库），叠加 `cover.Resolve` 直读同一文件，才出现同一文件两种身份。属设计缺口而非实现错误；在 `scanMode` 实现前保留，实现后由 mode 控制。

## 8. 与阶段 3-7 的关系

- **阶段 3**：列表页只显示 scene，不受影响；详情页预留附件区块位；附件机制在此阶段决策。
- **图片集页**：独立子步骤，阶段 4+ 或阶段 5。
- **阶段 4 插件/scraper**：Gallery NFO 可作为一种元数据来源。
- **阶段 5 性能**：附件按需读取 + 缩略图 + 缓存。
- **阶段 7 重构**：扫描模式抽象可推动 `file/scene/gallery` 解耦。

## 9. 实施顺序

1. 配置层加 `LibraryMode`（**冻结区**，人工确认）
2. 扫描流程按 mode 分派 handler/过滤（**冻结区**，人工确认）
3. 附件机制决策 + 识别 + 详情页展示（决策点在阶段 3；实现视 A/B/C 而定）
4. Gallery NFO 解析（**新增包** `backend/pkg/metadata/gallerynfo/`）
5. 图片集列表页（可改区）
6. 图片集详情页（可改区）

## 10. 不做（Non-goals）

- 例 2「多文件夹 = 一个图片集」：未来扩展「Gallery 分组」。
- 附件入库/索引：当前不做（见 D2）。
- 音频库、混合模式下的音频等：留待 flags 扩展。

## 11. 开放点

- Gallery NFO 命名约定（`folder.nfo` / 同名 `.nfo`）待定。
- D2 附件机制已决策为方案 A（见 §3 D2）；B/C 保留为未来扩展。
