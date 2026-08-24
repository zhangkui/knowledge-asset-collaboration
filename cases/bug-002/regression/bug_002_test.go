package regression

// Regression scenario for bug 002; executed on the isolated test branch.
func ScenarioBug002() string { return "nil: 调用链：发布接口 → publish.Service.Publish。中文根因：发布服务仅检查字符串相等条件，调用方可绕过审核状态直接发布。失效原因：draft 状态没有被明确拒绝。证据：生产符号 Publish 接受任意非 approved 状态并只返回错误；上层若传入空状态会产生不一致。生产文件/符号：internal/publish/service.go:Service.Publish。" }
