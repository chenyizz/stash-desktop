# 常见错误与解法

| 错误                        | 原因                       | 解法                                              |
| --------------------------- | -------------------------- | ------------------------------------------------- |
| `package xxx is not in std` | import 路径写错            | 检查 `go.mod` module 名                           |
| `CGO_ENABLED=0`             | Taskfile 写死              | 改 `build/windows/Taskfile.yml` 为 1              |
| `no matching files found`   | `frontend/dist` 为空       | 先 `cd frontend && npm run build`                 |
| `undefined: xxx`            | 缺包或字段名错             | 看编译器报错，贴完整输出                          |
| `s.Title != nil`            | 类型是 string 不是 *string | 直接 `s.Title`                                    |
| `result.Scenes`             | 字段私有                   | 用 `result.Resolve(ctx)`                          |
| SQL 查询报事务错误          | 没走 Repository 包装       | 用 `WithReadTxn` / `WithTxn`                      |
| 前端拿不到数据              | 方法未导出                 | Go 方法首字母大写                                 |
| 数据目录不可写              | 装在 Program Files         | 降级到 `%LOCALAPPDATA%\case`                      |
| 扫描完成等不到              | `setTimeout` 不可靠        | 后续改成事件通知                                  |
| `wails3 dev` 白屏           | 前端未构建或绑定未生成     | 先 `npm run build`，再 `wails3 generate bindings` |
| SQLite 锁                   | 并发写                     | 写操作用 `WithTxn`，不要裸 `db.Query`             |

## 排查顺序

1. 贴完整报错，不要截断。
2. 确认当前阶段和冻结区。
3. 确认改的是可改区。
4. 先 `go build ./...`。
5. 再 `cd frontend && npm run build`。
6. 再 `wails3 dev`。
7. 最后才怀疑数据库和系统环境。