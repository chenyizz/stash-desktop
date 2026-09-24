# 迁移路线图

## 阶段 2：最小闭环（已完成 2026-09-23）

### 2.4 详情页

- [x] Go 重写 NFO 解析：`backend/pkg/metadata/nfo/`
- [x] 扫描时自动读 NFO
- [x] `GetScene(id)` 返回完整 DTO
- [x] 详情页 Svelte 组件
- [x] hash 路由：列表 ↔ 详情

### 2.5 封面图

- [x] AssetServer 中间件挂载 `/covers/<id>`（blob store）
- [x] 前端 `<img src={coverUrl}>`（详情页）

## 阶段 3：核心业务（收尾中）

- [x] 演员页
- [x] 演员详情页
- [x] 标签页
- [x] 工作室页
- [x] 搜索
- [x] 过滤
- [x] 分页
- [x] 扫描完成事件，替代 setTimeout
- [x] 封面缩略图：imaging 缩放（原 2.5 顺延）
- [ ] 虚拟滚动
- [ ] 标签详情页 / 工作室详情页
- [ ] 阶段收尾：`docs/MIGRATIONS.md`

## 阶段 4：配置写（Settings）

- [ ] 设置页骨架（库 / 扫描 / 界面 / 任务 / 维护）
- [ ] 库路径增删改 + 每路径 `LibraryMode` + 校验/建目录
- [ ] 附件目录配置（替换硬编码 `fanart/poster/extra`）
- [ ] 扫描默认项 / UI 偏好
- [ ] 配置持久化 + 热生效矩阵（ADR-011）
- [ ] 首次使用引导（库路径为空）

## 阶段 5：内容写

- [ ] 演员 CRUD / 别名管理 / 合并（含别名）/ 头像上传
- [ ] 标签 CRUD / 别名 / 父子 / 合并
- [ ] 工作室 CRUD（合并暂不做）
- [ ] 场景编辑（元数据 / tags / performers / studio）/ 封面替换 / 删除 / 合并
- [ ] 重复检测入口

## 阶段 6：NFO 回写

- [ ] NFO writer（序列化）
- [ ] 显式导出（单场景 / 批量）+ 冲突策略（ADR-010）

## 阶段 7：任务与维护

- [ ] 任务队列 UI（进度 / 取消）
- [ ] 备份恢复 / 数据库优化 / 清缓存

## 阶段 8：插件系统（原阶段 4）

- [ ] Python 插件桥接：stderr JSON
- [ ] 插件配置界面
- [ ] 示例插件
- [ ] Scraper 作为插件类型设计

## 阶段 9：性能优化（原阶段 5）

- [ ] N+1 查询解决
- [ ] TanStack Query 缓存
- [ ] FTS / 索引优化
- [ ] Prepared Statement 缓存
- [ ] 前端懒加载

## 阶段 10：打包发布（原阶段 6）

- [ ] NSIS 安装包
- [ ] 便携版 zip
- [ ] 数据目录迁移策略
- [ ] 自动更新

## 阶段 11：重构（原阶段 7）

- [ ] 接口化 backend/manager
- [ ] 解耦 backend/pkg
- [ ] 清理 Stash 残留
- [ ] 补充回归测试
