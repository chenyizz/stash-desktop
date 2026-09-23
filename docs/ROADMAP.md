# 迁移路线图

## 阶段 2：最小闭环（当前）

### 2.4 详情页

- [x] Go 重写 NFO 解析：`backend/pkg/metadata/nfo/`
- [x] 扫描时自动读 NFO
- [ ] `GetScene(id)` 返回完整 DTO
- [ ] 详情页 Svelte 组件
- [ ] hash 路由：列表 ↔ 详情

### 2.5 封面图

- [ ] AssetServer 挂载 `data/blobs/`
- [ ] 前端 `<img src="/covers/xxx.jpg">`
- [ ] 缩略图生成：FFmpeg

## 阶段 3：核心业务

- [ ] 演员页
- [ ] 标签页
- [ ] 工作室页
- [ ] 搜索
- [ ] 过滤
- [ ] 分页
- [ ] 虚拟滚动
- [ ] 扫描完成事件，替代 setTimeout

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