package regression

// Regression scenario for bug 001; executed on the isolated test branch.
func ScenarioBug001() string { return "context: 调用链：HTTP 文档创建 → document.Repository.Create → Store 写入。中文根因：取消检查只覆盖进入函数前的瞬间，业务写入与取消之间没有原子边界。失效原因：客户端断开后仍可能完成持久化，造成孤儿草稿。证据：Create 在加锁后直接写入 docs。生产文件/符号：internal/document/repository.go:Create。" }
