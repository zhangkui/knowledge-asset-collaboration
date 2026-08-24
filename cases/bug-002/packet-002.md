# Bug 002

## user_query
Publishing a document must require an approved review state and must return a business error for drafts.

## bug_category
nil

## mode
bugfix

## verify_cmds
- go test ./...
- go test -race ./internal/...

## gold_root_cause
调用链：发布接口 → publish.Service.Publish。中文根因：发布服务仅检查字符串相等条件，调用方可绕过审核状态直接发布。失效原因：draft 状态没有被明确拒绝。证据：生产符号 Publish 接受任意非 approved 状态并只返回错误；上层若传入空状态会产生不一致。生产文件/符号：internal/publish/service.go:Service.Publish。

## success_criteria
目标行为：只有 approved 版本可发布。边界：空状态、draft、rejected、returned 均拒绝。合法场景：approved 返回 published。验证标准：回归测试覆盖所有非法状态并检查错误。

## go_version
go1.26.1 windows/amd64 (GOTOOLCHAIN=auto)

## repository_and_environment
branch_model: orphan-redgreen
green_branch: bug002_green
red_branch: bug002_red
base_commit(G1_sha): pending-local-commit
repo_url: https://github.com/zhangkui/knowledge-asset-collaboration/tree/bug002_green
