package regression
import("context";"testing";"time";"github.com/zhangkui/knowledge-asset-collaboration/internal/editor")
func TestBug006LockOwnership(t *testing.T){s:=editor.NewSessionManager();if _,err:=s.Acquire(context.Background(),"d","u1",time.Minute);err!=nil{t.Fatal(err)};if _,err:=s.Acquire(context.Background(),"d","u2",time.Minute);err==nil{t.Fatal("second editor must be blocked")}}