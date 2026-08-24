# Bug 020

## user_query
Concurrent version creation must produce unique version identifiers for a document.

## bug_category
concurrency

## mode
bugfix

## production_symbol
internal/document_version.Repository.Create

## gold_root_cause
中文根因：版本创建的标识生成和写入必须在同一并发安全策略下执行。生产文件/符号：internal/document_version/repository.go:Repository.Create。失效原因：并发编辑生成重复 ID 会覆盖历史版本。证据：Create写入items并分配ID。 生产文件/符号：internal/document_version.Repository.Create 调用链：HTTP/业务服务 → internal/document_version.Repository.Create。失效原因：并发编辑生成重复 ID 会覆盖历史版本。证据：Create写入items并分配ID。 证据：Create写入items并分配ID。

## success_criteria
目标行为：并发创建的版本 ID 全部唯一且记录数量完整。边界：同一文档和不同文档均适用。合法场景：多人协作自动保存产生多个版本。验证标准：race测试通过，ID集合大小等于创建数量。

## go_version
go1.26.1 windows/amd64 (GOTOOLCHAIN=auto)

## verify_cmds
- go test ./cases/bug-020/regression -count=1
- go test -race ./cases/bug-020/regression -count=10

## branches
- green: bug020_green
- red: bug020_red
