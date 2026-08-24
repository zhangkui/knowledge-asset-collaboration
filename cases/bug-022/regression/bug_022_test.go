package regression
import("context";"testing";"github.com/zhangkui/knowledge-asset-collaboration/internal/report")
func TestBug022EmptyReport(t *testing.T){m,err:=report.Build(context.Background());if err!=nil||m.GeneratedAt.IsZero()||m.Documents!=0{t.Fatalf("metrics=%+v err=%v",m,err)}}