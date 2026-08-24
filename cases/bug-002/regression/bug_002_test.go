package regression
import("context";"testing";"github.com/zhangkui/knowledge-asset-collaboration/internal/publish")
func TestBug002PublishApproval(t *testing.T){s:=publish.Service{};if _,err:=s.Publish(context.Background(),"draft");err==nil{t.Fatal("draft cannot publish")};if got,err:=s.Publish(context.Background(),"approved");err!=nil||got!="published"{t.Fatalf("got=%s err=%v",got,err)}}