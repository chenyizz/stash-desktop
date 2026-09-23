```text
# Agent 工具与运行规范

## 1. 目的

本文件定义 Case 项目开发 Agent 的工具权限、自动更新协议、MCP 与 Skills 使用原则。Agent 每次执行任务前必须读取并遵守。

## 2. 运行环境

- Agent 通过 DeepSeek API 调用。
- Agent 支持工具调用。
- 工作目录限制在项目根目录：
  `D:\cheny\Documents\代码\开发\Case`
- 所有文件读写、命令执行都必须限制在该目录内。

## 3. 必须工具

Agent 必须具备以下工具：

- `read_file`：读取项目文件。
- `write_file`：更新状态、写代码、追加日志。
- `list_dir`：列目录。
- `run_command`：执行构建、测试、开发命令。
- `fetch_url`：读取 GitHub、Wails、Svelte 等在线文档。

## 4. 可选工具

- `search_web`：搜索最新文档。
- `git_status`
- `git_diff`
- `apply_patch`

## 5. 权限边界

允许：

- 读写项目根目录内文件。
- 执行项目内构建和开发命令。
- 读取公开在线文档。

禁止：

- 全盘写权限。
- 无限制 shell。
- 直连生产数据库。
- 删除项目外文件。
- 未经确认修改冻结区。

冻结区：

- `backend/manager` 核心结构
- `backend/pkg` 核心结构
- 从 Stash 原样迁移的 SQL / Repository / Model

可改区：

- `internal/app`
- `frontend`
- 新增 `backend/pkg/metadata/nfo/`
- `docs/`
- `AGENTS.md`
- `CHANGELOG.md`

## 6. MCP 接入原则

如果 Agent 框架支持 MCP，优先接入：

- `filesystem`：权限限定在项目根目录。
- `shell`：权限限定在项目根目录。
- `fetch`：用于读取在线文档。

如果框架自带上述工具，则不必额外接入 MCP。

## 7. Skills 使用原则

当前阶段不需要 Skills。

优先级：

1. 文件读写
2. 终端执行
3. 联网 fetch
4. 自动更新状态
5. Skills

将来可选 Skills：

- 代码审查
- 迁移检查
- NFO 解析
- Svelte 组件生成

## 8. 自动更新协议

每完成一个子步骤，Agent 必须：

1. 运行验证命令。
2. 更新 `docs/PROJECT_STATE.md`。
3. 追加 `CHANGELOG.md`。
4. 如解决新错误，追加 `docs/TROUBLESHOOTING.md`。
5. 如完成路线图项，勾选 `docs/ROADMAP.md`。
6. 不自动修改 `AGENTS.md`，除非硬规则变化。

更新文件前必须先读原文件，保留历史记录，只追加或替换对应段落。

## 9. 系统提示词片段

将以下内容加入 Agent 系统提示：

​```text
你是 Case 项目的开发 Agent。
每次任务开始先读 AGENTS.md、docs/PROJECT_STATE.md、docs/WORKFLOW.md。
复述当前阶段、焦点、冻结区、本次任务。
一次只做一个子步骤。
未经确认不改冻结区。
完成后自动更新 docs/PROJECT_STATE.md 和 CHANGELOG.md。
解决新错误后追加 docs/TROUBLESHOOTING.md。
完成路线图项后勾选 docs/ROADMAP.md。
不要自动改 AGENTS.md，除非硬规则变化。
​```
```