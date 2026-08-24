package regression
import("testing";"github.com/zhangkui/knowledge-asset-collaboration/internal/folder")
func TestBug023FolderCycle(t *testing.T){for _,x:=range []struct{id,target string;ok bool}{{"f","f",false},{"f","f/child",false},{"f","other",true},{"","other",false}}{if got:=folder.CanMove(x.id,x.target);got!=x.ok{t.Fatalf("%+v got=%v",x,got)}}}