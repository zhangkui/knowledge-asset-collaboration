# Bug 022

## user_query
Report generation must return a usable zero metric set when no activity exists.

## bug_category
nil

## mode
diagnosis

## production_symbol
internal/report.Build

## gold_root_cause
中文根因：统计报表需要返回完整可序列化指标，而不是 nil 或未初始化结构。生产文件/符号：internal/report/report.go:Build。失效原因：看板读取空指针字段会导致页面渲染失败。证据：Build返回 Metrics并设置 GeneratedAt。 生产文件/符号：internal/report.Build 调用链：HTTP/业务服务 → internal/report.Build。失效原因：看板读取空指针字段会导致页面渲染失败。证据：Build返回 Metrics并设置 GeneratedAt。 证据：Build返回 Metrics并设置 GeneratedAt。

## success_criteria
目标行为：空数据生成时间有效且所有数值为零。边界：取消上下文返回取消错误。合法场景：新空间无文档时仍可展示看板。验证标准：GeneratedAt非零，所有计数为零且无 panic。

## go_version
go1.26.1 windows/amd64 (GOTOOLCHAIN=auto)

## verify_cmds
- go test ./cases/bug-022/regression -count=1
- go test -race ./cases/bug-022/regression -count=10

## branches
- green: bug022_green
- red: bug022_red
