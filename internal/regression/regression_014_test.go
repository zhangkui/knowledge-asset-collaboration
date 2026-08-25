package regression

import (
    "context"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/attachment"
    "github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
)

func TestBug14_UploadChunkSnapshotsDoNotMutateInternalState(t *testing.T) {
    app := shared.NewApp()
    ctx := context.Background()
    upload, err := app.Uploads.Start(ctx, "doc-14", "release.zip", "application/zip", 10, 1)
    if err != nil {
        t.Fatalf("start upload: %v", err)
    }
    if _, err := app.Uploads.PutChunk(ctx, upload.ID, attachment.Chunk{Number: 0, Offset: 0, Size: 10, Checksum: "sha-14"}); err != nil {
        t.Fatalf("put upload chunk: %v", err)
    }
    snapshot, err := app.Uploads.Get(ctx, upload.ID)
    if err != nil {
        t.Fatalf("get upload snapshot: %v", err)
    }
    snapshot.Chunks[0] = attachment.Chunk{Number: 0, Size: 999, Checksum: "tampered"}
    again, err := app.Uploads.Get(ctx, upload.ID)
    if err != nil {
        t.Fatalf("get upload after caller mutation: %v", err)
    }
    if again.Chunks[0].Size != 10 || again.Chunks[0].Checksum != "sha-14" {
        t.Fatalf("caller mutation polluted upload state: %#v", again.Chunks)
    }
}
