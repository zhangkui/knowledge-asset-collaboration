# Bug 010

## user_query
Investigate why a failed notification read operation can be reported as successful after the request context is cancelled.

## bug_category
error

## mode
diagnosis

## verify_cmds
- go test ./...
- go test -race ./internal/...

## gold_root_cause
调用链：通知中心批量已读 API → Center.MarkAllRead。中文根因：MarkAllRead 在获取锁前未检查上下文，取消请求仍会遍历并写入通知。失效原因：客户端超时后服务端继续改变未读计数，造成审计与 UI 不一致。证据：生产符号需要在事务边界内检查 ctx 并返回取消错误。生产文件/符号：internal/notification/center.go:Center.MarkAllRead。

## success_criteria
目标行为：取消请求返回 context.Canceled 且不改变通知状态。边界：取消前正常请求全部标记已读。合法场景：只影响指定用户。验证标准：取消上下文回归测试检查错误、未读数量和其他用户通知均不变。

## go_version
go1.26.1 windows/amd64 (GOTOOLCHAIN=auto)

## repository_and_environment
branch_model: orphan-redgreen
green_branch: bug010_green
red_branch: bug010_red
base_commit(G1_sha): pending-local-commit
repo_url: https://github.com/zhangkui/knowledge-asset-collaboration/tree/bug010_green
