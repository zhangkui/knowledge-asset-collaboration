# Bug 013

## user_query
Audit log listing must return an isolated snapshot so callers cannot mutate historical events.

## bug_category
slice

## mode
diagnosis

## production_symbol
internal/audit.Logger.List

## gold_root_cause
中文根因：审计查询返回值必须与内部事件存储隔离。生产文件/符号：internal/audit/logger.go:Logger.List。失效原因：外部修改返回切片会篡改后续审计查询结果。证据：List 复制切片但事件字段仍需保持不可变语义。 生产文件/符号：internal/audit.Logger.List 调用链：HTTP/业务服务 → internal/audit.Logger.List。失效原因：外部修改返回切片会篡改后续审计查询结果。证据：List 复制切片但事件字段仍需保持不可变语义。 证据：List 复制切片但事件字段仍需保持不可变语义。

## success_criteria
目标行为：查询结果修改不影响日志存储。边界：空日志返回空切片而非异常。合法场景：多条日志按写入顺序可查询。验证标准：修改第一次查询结果后，第二次查询仍保留原始 ActorID 和 Action。

## go_version
go1.26.1 windows/amd64 (GOTOOLCHAIN=auto)

## verify_cmds
- go test ./cases/bug-013/regression -count=1
- go test -race ./cases/bug-013/regression -count=10

## branches
- green: bug013_green
- red: bug013_red
