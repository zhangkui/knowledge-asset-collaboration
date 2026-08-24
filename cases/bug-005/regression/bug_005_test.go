package regression

// Regression scenario for bug 005; executed on the isolated test branch.
func ScenarioBug005() string { return "error: 调用链：分片上传完成 API → attachment.Service.Complete。中文根因：完成操作无状态转换校验，重复请求继续覆盖 Uploaded。失效原因：重试和重复回调无法区分幂等成功与非法重复。证据：Complete 无 completed 状态字段或重复调用检查。生产文件/符号：internal/attachment/service.go:Service.Complete。" }
