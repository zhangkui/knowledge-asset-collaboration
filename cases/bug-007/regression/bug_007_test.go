package regression

// Regression scenario for bug 007; executed on the isolated test branch.
func ScenarioBug007() string { return "nil: 调用链：受保护 HTTP 路由 → auth.Service.Authorize。中文根因：Authorize 仅检查 Bearer 前缀和点号数量，没有校验签名与 ExpiresAt。失效原因：任意伪造的三段字符串被当作有效身份。证据：Authorize 直接返回 Claims{Subject:p[0]}。生产文件/符号：internal/auth/service.go:Service.Authorize。" }
