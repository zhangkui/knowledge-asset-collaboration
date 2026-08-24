package regression

// Regression scenario for bug 004; executed on the isolated test branch.
func ScenarioBug004() string { return "concurrency: 调用链：团队成员 API → organization.AddMember。中文根因：成员切片按值返回，调用方在并发更新时可能基于旧快照覆盖另一请求。失效原因：缺少集中化的并发更新或版本校验。证据：AddMember 只操作调用者持有的 Team 副本。生产文件/符号：internal/organization/service.go:AddMember。" }
