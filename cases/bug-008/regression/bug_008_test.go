package regression
import("context";"testing";"time";"github.com/zhangkui/knowledge-asset-collaboration/internal/share")
func TestBug008ShareExpiry(t *testing.T){l,err:=share.Create(context.Background(),"doc","read",time.Hour);if err!=nil||!share.Valid(l,time.Now()){t.Fatalf("link=%+v err=%v",l,err)};if share.Valid(l,time.Now().Add(2*time.Hour)){t.Fatal("expired link accepted")}}