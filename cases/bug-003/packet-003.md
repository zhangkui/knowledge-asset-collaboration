# Bug 003

## user_query
Search with an empty query must return no results instead of the entire index.

## bug_category
slice

## mode
bugfix

## verify_cmds
- go test ./...
- go test -race ./internal/...

## gold_root_cause
调用链：搜索 API → search.Index.Query。中文根因：Query 对空字符串直接执行 Contains，所有标题正文都匹配。失效原因：缺少空查询边界保护。证据：strings.Contains(value, "") 恒为真。生产文件/符号：internal/search/index.go:Index.Query。

## success_criteria
目标行为：空白查询返回空结果。边界：空串和全空白均为空；非空查询保持匹配。合法场景：标题或正文包含关键词时返回结果。验证标准：测试断言空查询长度为零且正常查询仍返回命中。

## go_version
go1.26.1 windows/amd64 (GOTOOLCHAIN=auto)

## repository_and_environment
branch_model: orphan-redgreen
green_branch: bug003_green
red_branch: bug003_red
base_commit(G1_sha): pending-local-commit
repo_url: https://github.com/zhangkui/knowledge-asset-collaboration/tree/bug003_green
