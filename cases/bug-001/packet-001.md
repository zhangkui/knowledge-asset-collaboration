# Bug 001

## user_query
Document creation must stop when the request context is cancelled. A client that disconnects while creating a document must not leave a document record behind.

## bug_category
context

## mode
bugfix

## verify_cmds
- go test ./...
- go test -race ./internal/...

## gold_root_cause
中文根因：Repository.Create 只在加锁前检查 context，获取写锁后到写入 docs 映射之间没有再次检查取消状态，导致客户端断开后仍可能持久化文档。生产文件/符号：internal/document/repository.go:Repository.Create。调用链：HTTP 文档创建 → document.Repository.Create → 加锁 → 生成文档 ID → 写入 Repository.docs。失效原因：取消发生在第一次检查之后时，旧实现直接继续写入。证据：bug1_red 的 Create 在 r.mu.Lock() 后没有 ctx.Err() 检查，而 bug1_green 在写入前增加了第二次检查。

## success_criteria
目标行为：请求上下文在校验后取消时，Create 必须返回 context.Canceled 或 context.DeadlineExceeded，且不得新增文档。边界：取消发生在入口检查后、获取锁后、写入前仍必须生效。合法场景：未取消且标题有效的请求仍创建 draft 文档并返回非空 ID。验证标准：公开回归在 bug1_red 上失败、在 bug1_green 上通过；go test ./... 与 go test -race ./internal/... 均通过绿分支。

## go_version
go1.26.1 windows/amd64 (GOTOOLCHAIN=auto)

## repository_and_environment
branch_model: orphan-redgreen
green_branch: bug1_green
red_branch: bug1_red
base_commit(G1_sha): 2b188c3
repo_url: https://github.com/zhangkui/knowledge-asset-collaboration/tree/bug1_green