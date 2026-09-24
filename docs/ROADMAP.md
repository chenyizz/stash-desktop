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

## 阶段 3：核心业务

- [x] 演员页
- [x] 标签页
- [x] 工作室页
- [ ] 搜索
- [ ] 过滤
- [x] 分页
- [ ] 虚拟滚动
- [x] 扫描完成事件，替代 setTimeout
- [ ] 封面缩略图：FFmpeg（原 2.5 顺延）

## 阶段 4：插件系统

- [ ] Python 插件桥接：stderr JSON
- [ ] 插件配置界面
- [ ] 示例插件
- [ ] Scraper 作为插件类型设计

## 阶段 5：性能优化

- [ ] N+1 查询解决
- [ ] TanStack Query 缓存
- [ ] 虚拟列表
- [ ] Prepared Statement 缓存
- [ ] 前端懒加载

## 阶段 6：打包发布

- [ ] NSIS 安装包
- [ ] 便携版 zip
- [ ] 数据目录迁移策略
- [ ] 自动更新

## 阶段 7：重构

- [ ] 接口化 backend/manager
- [ ] 解耦 backend/pkg
- [ ] 清理 Stash 残留
- [ ] 补充回归测试