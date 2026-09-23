# 工作流

## 1. 会话启动

如果 Agent 能读文件，让它先读：

1. `AGENTS.md`
2. `docs/PROJECT_STATE.md`
3. `docs/WORKFLOW.md`

然后要求它复述：

- 当前阶段
- 当前焦点
- 本次任务
- 冻结区
- 可改区
- 计划修改的文件

如果 Agent 不能自动读文件，在首条消息里要求它调用 `read_file` 读取上述文件。

不要每次把全部文档贴进上下文。

## 2. 任务粒度

一次只推进一个子步骤。

好例子：

- 只做 `GetScene(id)`。
- 只做 NFO 解析。
- 只做详情页组件。

坏例子：

- 同时做详情页、封面图、路由、分页。

## 3. Agent 输出要求

任务开始时，先让 Agent 列计划：

```text
请不要直接写代码。
先列出：
1. 需要改哪些文件
2. 每个文件改什么
3. 数据从哪来
4. 可能遇到的坑
5. 验证命令
我确认后再写代码。
```

写代码时要求：

```text
请给出完整文件内容，不要只给 diff 片段，我要直接覆盖。
```

## 4. 执行规则

- 冻结区：`backend/manager`、`backend/pkg` 核心结构。
- 可改区：`internal/app`、`frontend`、`docs`、新增 NFO 模块。
- 改冻结区前，必须人工确认。
- 一次只改一个子步骤。
- 每步必须能回滚。

## 5. 验证命令

```bash
go build ./...

cd frontend && npm run build

wails3 dev
```

如果失败，贴完整输出，不要截断。

## 6. Agent 收尾协议

每完成一个子步骤，Agent 必须自动更新：

1. `docs/PROJECT_STATE.md`
   - 当前焦点
   - 已完成
   - 下一步
   - 已知问题
   - 最近变更
2. `CHANGELOG.md`
   - 日期
   - 新增
   - 修改
   - 修复
   - 验证结果
3. `docs/TROUBLESHOOTING.md`
   - 如果遇到新错误并解决，追加一行。
4. `docs/ROADMAP.md`
   - 如果完成路线图项，勾选。
5. `AGENTS.md`
   - 只有硬规则变化才改。
   - 不要每次自动改。
6. `docs/CODEMAP.md`
   - 如新增、删除、重命名模块或关键文件，更新对应模块表和“变更日志”。
   - 如发现新的常用搜索命令，补进“常用搜索命令”段。

## 7. 上下文控制

- `AGENTS.md` 控制在 150 行以内。
- `PROJECT_STATE.md` 控制在 100 行以内。
- 其他文档按需读取，不要每次全读。
- 任务只涉及前端时，不读 `backend/pkg` 细节。
- 任务只涉及 Go 时，不读前端组件。
- 让 Agent 先搜索相关文件，再读具体文件。
- 探索代码前先读 `docs/CODEMAP.md`，不要在每轮任务里重复读同一批源文件。
- `docs/CODEMAP.md` 由 Agent 自动维护，控制在 200 行以内。

## 8. 自动更新怎么实现

Agent 必须有写文件工具，例如：

- `read_file`
- `write_file`
- `list_dir`
- `run_command`
- `fetch_url`

在 Agent 系统提示里写死：

```text
每完成一个子步骤，必须更新 docs/PROJECT_STATE.md 和 CHANGELOG.md。
如果解决新错误，追加 docs/TROUBLESHOOTING.md。
如果完成路线图项，勾选 docs/ROADMAP.md。
不要自动改 AGENTS.md，除非硬规则变化。
```