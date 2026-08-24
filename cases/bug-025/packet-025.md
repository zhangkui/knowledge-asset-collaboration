# Bug 025

## user_query
Concurrent notification delivery must preserve every notification and mark only the target user read.

## bug_category
concurrency

## mode
bugfix

## production_symbol
internal/notification.Center.Push

## gold_root_cause
中文根因：通知中心需要在并发写入和按用户读取之间保持数据一致性。生产文件/符号：internal/notification/center.go:Center.Push、MarkAllRead。失效原因：无锁或错误过滤会丢通知或误读其他用户消息。证据：Center维护共享items切片并提供按用户筛选。 生产文件/符号：internal/notification.Center.Push 调用链：HTTP/业务服务 → internal/notification.Center.Push。失效原因：无锁或错误过滤会丢通知或误读其他用户消息。证据：Center维护共享items切片并提供按用户筛选。 证据：Center维护共享items切片并提供按用户筛选。

## success_criteria
目标行为：并发推送全部保留，标记用户A不影响用户B。边界：重复消息也应各自保留，空用户查询返回空。合法场景：评论、审核和发布事件同时推送。验证标准：race测试无竞争，用户计数准确，其他用户仍有未读通知。

## go_version
go1.26.1 windows/amd64 (GOTOOLCHAIN=auto)

## verify_cmds
- go test ./cases/bug-025/regression -count=1
- go test -race ./cases/bug-025/regression -count=10

## branches
- green: bug025_green
- red: bug025_red
