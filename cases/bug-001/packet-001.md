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
调用链：HTTP 文档创建 → document.Repository.Create → Store 写入。中文根因：取消检查只覆盖进入函数前的瞬间，业务写入与取消之间没有原子边界。失效原因：客户端断开后仍可能完成持久化，造成孤儿草稿。证据：Create 在加锁后直接写入 docs。生产文件/符号：internal/document/repository.go:Create。

## success_criteria
目标行为：取消请求不得创建文档。边界：取消发生在校验后、写入前也必须生效。合法场景：正常上下文仍可创建。验证标准：回归测试在取消上下文下观察不到新增文档。

## go_version
go1.26.1 windows/amd64 (GOTOOLCHAIN=auto)

## repository_and_environment
branch_model: orphan-redgreen
green_branch: bug001_green
red_branch: bug001_red
base_commit(G1_sha): pending-local-commit
repo_url: https://github.com/zhangkui/knowledge-asset-collaboration/tree/bug001_green
