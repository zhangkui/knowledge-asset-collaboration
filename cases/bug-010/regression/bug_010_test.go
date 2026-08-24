package regression

// Regression scenario for bug 010; executed on the isolated test branch.
func ScenarioBug010() string { return "error: 调用链：通知中心批量已读 API → Center.MarkAllRead。中文根因：MarkAllRead 在获取锁前未检查上下文，取消请求仍会遍历并写入通知。失效原因：客户端超时后服务端继续改变未读计数，造成审计与 UI 不一致。证据：生产符号需要在事务边界内检查 ctx 并返回取消错误。生产文件/符号：internal/notification/center.go:Center.MarkAllRead。" }
