# Bug 021

## user_query
Expired recycle-bin cleanup must stop promptly when its context is cancelled.

## bug_category
context

## mode
diagnosis

## production_symbol
internal/recycle.Bin.PurgeExpired

## gold_root_cause
中文根因：回收站批量清理属于后台任务，必须响应取消信号。生产文件/符号：internal/recycle/bin.go:PurgeExpired。失效原因：服务关闭时清理继续持锁会延迟退出。证据：PurgeExpired遍历并重建items切片。 生产文件/符号：internal/recycle.Bin.PurgeExpired 调用链：HTTP/业务服务 → internal/recycle.Bin.PurgeExpired。失效原因：服务关闭时清理继续持锁会延迟退出。证据：PurgeExpired遍历并重建items切片。 证据：PurgeExpired遍历并重建items切片。

## success_criteria
目标行为：取消上下文返回0且不执行清理。边界：有效上下文清理过期项目，未过期项目保留。合法场景：定时任务在正常上下文运行。验证标准：取消调用不删除项目，有效调用只删除ExpiresAt早于now的项目。

## go_version
go1.26.1 windows/amd64 (GOTOOLCHAIN=auto)

## verify_cmds
- go test ./cases/bug-021/regression -count=1
- go test -race ./cases/bug-021/regression -count=10

## branches
- green: bug021_green
- red: bug021_red
