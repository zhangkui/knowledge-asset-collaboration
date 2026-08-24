# Bug 004

## user_query
Concurrent team member updates must not lose members when two requests add users at the same time.

## bug_category
concurrency

## mode
bugfix

## verify_cmds
- go test ./...
- go test -race ./internal/...

## gold_root_cause
调用链：团队成员 API → organization.AddMember。中文根因：成员切片按值返回，调用方在并发更新时可能基于旧快照覆盖另一请求。失效原因：缺少集中化的并发更新或版本校验。证据：AddMember 只操作调用者持有的 Team 副本。生产文件/符号：internal/organization/service.go:AddMember。

## success_criteria
目标行为：并发添加不同用户最终保留全部成员。边界：重复用户仍只出现一次。合法场景：串行添加行为不变。验证标准：race 测试并发添加 100 个用户，最终集合完整且无重复。

## go_version
go1.26.1 windows/amd64 (GOTOOLCHAIN=auto)

## repository_and_environment
branch_model: orphan-redgreen
green_branch: bug004_green
red_branch: bug004_red
base_commit(G1_sha): pending-local-commit
repo_url: https://github.com/zhangkui/knowledge-asset-collaboration/tree/bug004_green
