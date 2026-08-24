# Bug 012

## user_query
Completing an attachment must reject unknown attachment IDs and preserve the original upload state.

## bug_category
error

## mode
bugfix

## production_symbol
internal/attachment.Service.Complete

## gold_root_cause
中文根因：附件完成接口需要区分不存在资源和正常完成。生产文件/符号：internal/attachment/service.go:Service.Complete。失效原因：调用方把重试请求当成成功会导致下载列表显示不完整附件。证据：Complete 读取 items 后更新 Uploaded 字段。 生产文件/符号：internal/attachment.Service.Complete 调用链：HTTP/业务服务 → internal/attachment.Service.Complete。失效原因：调用方把重试请求当成成功会导致下载列表显示不完整附件。证据：Complete 读取 items 后更新 Uploaded 字段。 证据：Complete 读取 items 后更新 Uploaded 字段。

## success_criteria
目标行为：未知 ID 返回明确错误；已存在附件完成后 Uploaded 等于 Size。边界：大小为零和重复完成都不能破坏状态。合法场景：已启动分片附件可以完成。验证标准：错误非 nil，完成结果 Uploaded==Size，重复完成保持幂等。

## go_version
go1.26.1 windows/amd64 (GOTOOLCHAIN=auto)

## verify_cmds
- go test ./cases/bug-012/regression -count=1
- go test -race ./cases/bug-012/regression -count=10

## branches
- green: bug012_green
- red: bug012_red
