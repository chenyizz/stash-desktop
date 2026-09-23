1. 1. # AGENTS.md — Case 桌面应用（Stash 迁移）

      ## 0. Agent 启动协议

      每次新会话开始时，Agent 必须先读：

      1. `AGENTS.md`
      2. `docs/PROJECT_STATE.md`
      3. `docs/WORKFLOW.md`
   4. `docs/CODEMAP.md`   
   
   然后复述：
   
      - 当前阶段
      - 当前焦点
      - 本次任务
      - 冻结区
      - 可改区
   - 本次计划修改的文件
   
   未经确认，不要直接大改代码。
   
   ## 1. 项目是什么
   
   把 Stash（Web 应用）迁移为 Windows/Linux 桌面应用。
   
   技术栈：
   
      - Wails v3
      - Go
      - Svelte 5
      - TypeScript
   - SQLite
   
   参考原版：
   
      - https://github.com/stashapp/stash
   - 分支：develop
   
   ## 2. 关键路径
   
      - 项目根目录：`D:\cheny\Documents\代码\开发\Case`
      - Go module 名：`case`
      - 数据目录（开发时）：`bin\data\`
      - 日志：`bin\log\case.log`
      - 数据库：`bin\data\case.db`
      - 前端入口：`frontend\src\App.svelte`
      - Wails 服务：`internal\app\app.go`
      - 核心业务：`backend\manager\` 和 `backend\pkg\`
   - 文档目录：`docs\`
   
   ## 3. 冻结区与可改区
   
   ### 冻结区：迁移期不要顺手重构
   
      - `backend/manager` 核心结构
      - `backend/pkg` 核心结构
   - 从 Stash 原样迁移的 SQL / Repository / Model
   
   ### 可改区
   
      - `internal/app`
      - `frontend`
      - 新增 `backend/pkg/metadata/nfo/`
      - `docs/`
      - `AGENTS.md`
   - `CHANGELOG.md`
   
   如果要改冻结区，Agent 必须先输出计划，人工确认后再动。
   
      > “不要重构”不是永远不重构，而是迁移期冻结。
   > 将来阶段 7 可以做独立重构，但必须有测试、小步提交、行为不变。
   
   ## 4. 硬规则
   
      1. 迁移期不要重构 `backend/manager`、`backend/pkg` 核心结构。只在 `internal/app` 做适配。
      2. 不要删除带 `#issue` 编号注释的代码。每一条都是踩坑修复。
      3. 前端不直接依赖 `backend/pkg`。只通过 `internal/app` 暴露的 Wails 方法访问。
      4. 新增 Go 方法：小写开头为私有，大写开头才会暴露给前端。
      5. 所有 SQL 查询必须通过 `mgr.Repository` 的 `WithReadTxn` / `WithTxn` 包装。
      6. CGO 必须开启：`CGO_ENABLED=1`。
      7. 不要改 `build/windows/Taskfile.yml` 里的 `CGO_ENABLED=0`。
      8. 数据目录策略：exe 同目录可写则用 exe 目录；不可写则降级到 `%LOCALAPPDATA%\case`。
      9. Blob 存储默认 `FILESYSTEM`，不存数据库。
   10. 日志用 `logger.Infof` / `logger.Debugf` / `logger.Errorf`，不要用 `fmt.Printf` 或标准库 `log`。
   
   ## 5. 效率规则（减少 token 消耗）   <!-- 新增 -->
   
   1. **先查代码地图**：`docs/CODEMAP.md` 提供模块定位、关键类型摘要、常用搜索命令。优先看它，不要全目录遍历。
      2. **禁止目录遍历**：不要用 `Get-ChildItem -Recurse` 或 `find .` 遍历整个目录。
      3. **精确搜索**：找定义用 `rg "type X struct"`，找调用用 `rg "X\("`，确认行号后再 `read_file`。
      4. **探索预算**：一次任务最多读 5 个文件、跑 3 条探索命令，然后必须出计划。
      5. **命令输出截断**：构建/测试输出超过 30 行时只保留末尾 30 行；日志只抓 `ERROR` / `WARN`。
      6. **只读相关包**：只读 `backend/pkg` 中与当前任务相关的包，不要全读。
      7. **接续会话**：优先用 `opencode --continue` 或 `--session <id>`，不要每次开新会话。
      8. **主动压缩**：对话超过 50 轮时建议 `/compact`。
   9. **写代码不废话**：只给完整文件内容，不要逐行解释，除非用户要求。
   
   ## 6. 已移除的 Stash 模块
   
      不要重新引入：
   
      - GraphQL / gqlgen
      - Apollo Client
      - DLNA
      - StashDB 客户端
      - Scraper 系统（暂留但不再使用）
      - Desktop 模块
      - HTTP 服务器相关：host / port / jwt / session
   
      ## 7. 构建与运行
   
   ```bash
      # 编译
   go build ./...
      
   # 开发模式
      wails3 dev
      
      # 前端构建
      cd frontend && npm run build
      
      # 打包
      wails3 build
   ```
   
   ## 8. 最致命错误
   
      详细见 `docs/TROUBLESHOOTING.md`。
   
      - `CGO_ENABLED=0`：改成 1。
      - `package xxx is not in std`：检查 `go.mod` module 名。
      - `no matching files found`：先 `npm run build`。
      - `s.Title != nil`：类型是 string 不是 *string。
   - `result.Scenes` 私有：用 `result.Resolve(ctx)`。
      - 前端拿不到数据：Go 方法首字母没大写。
   - SQL 事务错误：没走 `WithReadTxn` / `WithTxn`。
   
   ## 9. 文档索引
   
      - 当前状态：`docs/PROJECT_STATE.md`
      - 工作流：`docs/WORKFLOW.md`
      - **代码地图：`docs/CODEMAP.md`**   <!-- 新增 -->
      - Agent 工具与 MCP：`docs/AGENT_TOOLS.md`
      - 错误排查：`docs/TROUBLESHOOTING.md`
      - 架构决策：`docs/ARCHITECTURE.md`
      - 代码约定：`docs/CONVENTIONS.md`
      - 路线图：`docs/ROADMAP.md`
   
      ## 10. Agent 收尾协议
   
      每完成一个子步骤，Agent 必须：
   
      1. 跑验证命令。
      2. 更新 `docs/PROJECT_STATE.md`。
      3. 追加 `CHANGELOG.md`。
      4. 如解决新错误，追加 `docs/TROUBLESHOOTING.md`。
      5. 如完成路线图项，勾选 `docs/ROADMAP.md`。
      6. **如新增、删除、重命名模块或关键文件，更新 `docs/CODEMAP.md`**，并在其“变更日志”表追加一行。   
      7. 只有硬规则变化时，才改 `AGENTS.md`。
   
   ## 11. 自动提交协议
   
   每完成一个子步骤，且满足全部条件时，自动提交：
   
   1. `go build ./...` 通过
   2. `go vet` 通过
   3. `go test ./...` 通过
   4. 涉及前端或绑定时，`wails3 dev` 冒烟通过
   5. 工作区只有本子步骤相关的改动
   
   提交规则：
   
   - 禁止 `git add .` 或 `git add -A`
   - 只 add 本子步骤明确修改/新增的文件
   - commit message 格式：`<type>(<scope>): <简述>`
   - 提交前先 `git status`，有无关文件则跳过并提醒我
   
   不自动提交的情况：
   
   - 任何验证命令失败
   - 工作区有未跟踪的敏感文件
   - 改动涉及冻结区
   - commit message 无法用一句话描述