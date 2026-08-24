package regression

// Regression scenario for bug 006; executed on the isolated test branch.
func ScenarioBug006() string { return "context: 调用链：审核决定 API → review.Service.Decide。中文根因：Decide 在加锁前未检查 context，取消请求仍会写入审核记录。失效原因：超时请求可能在客户端已放弃后改变发布流程。证据：生产符号 Decide 缺少 ctx.Err 检查。生产文件/符号：internal/review/service.go:Service.Decide。" }
