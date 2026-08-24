package regression
import("context";"testing";"github.com/zhangkui/knowledge-asset-collaboration/internal/permission")
func TestBug007PermissionDeny(t *testing.T){s:=&permission.Service{};ctx:=context.Background();_=s.Grant(ctx,permission.Grant{SubjectID:"u",ResourceID:"doc",Permission:"edit",ExplicitDeny:true});if s.Allowed(ctx,"u","doc","edit"){t.Fatal("deny must reject")}}