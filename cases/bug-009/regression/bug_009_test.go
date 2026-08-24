package regression

// Regression scenario for bug 009; executed on the isolated test branch.
func ScenarioBug009() string { return "slice: 调用链：文件夹移动 API → folder.CanMove。中文根因：循环检测使用字符串前缀判断路径，节点 ID 的边界没有结构化解析。失效原因：folder-1 与 folder-10 这类前缀关系被误判为父子关系。证据：CanMove 使用 strings.HasPrefix(target, id+"/")，调用方传入扁平 ID 时语义不一致。生产文件/符号：internal/folder/service.go:CanMove。" }
