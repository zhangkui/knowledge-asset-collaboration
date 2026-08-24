package regression

// Regression scenario for bug 003; executed on the isolated test branch.
func ScenarioBug003() string { return "slice: 调用链：搜索 API → search.Index.Query。中文根因：Query 对空字符串直接执行 Contains，所有标题正文都匹配。失效原因：缺少空查询边界保护。证据：strings.Contains(value, "") 恒为真。生产文件/符号：internal/search/index.go:Index.Query。" }
