package regression

// Regression scenario for bug 008; executed on the isolated test branch.
func ScenarioBug008() string { return "concurrency: 调用链：WebSocket 加入/断开 → editor.Hub.Join/Leave → rooms map。中文根因：协作状态依赖内存房间，连接生命周期事件的顺序没有版本化，旧连接的 Leave 可能删除新连接的同一用户状态。失效原因：重连竞态会让在线用户列表短暂丢失。证据：Leave 只按 documentID/userID 删除，不验证连接代数。生产文件/符号：internal/editor/hub.go:Hub.Leave。" }
