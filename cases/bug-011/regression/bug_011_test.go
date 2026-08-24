package regression
import("context";"testing";"github.com/zhangkui/knowledge-asset-collaboration/internal/annotation")
func TestBug011CancelledAnnotation(t *testing.T){ctx,cancel:=context.WithCancel(context.Background());cancel();_,err:=annotation.Create(ctx,annotation.Annotation{DocumentID:"d",AuthorID:"u",StartOffset:0,EndOffset:1,Note:"x"});if err==nil{t.Fatal("expected cancellation")}}