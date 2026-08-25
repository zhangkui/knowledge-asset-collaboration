package regression

import (
    "context"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/editor"
)

func TestBug17_DisconnectedSessionCannotBeRevivedByHeartbeat(t *testing.T) {
    ctx := context.Background()
    sessions := editor.NewSessionManager()
    hub := editor.NewHub()
    if _, err := sessions.Connect(ctx, "session-17", "doc-17", "editor-17"); err != nil {
        t.Fatalf("connect session: %v", err)
    }
    if err := hub.Join(ctx, "doc-17", editor.Presence{UserID: "editor-17", DocumentID: "doc-17"}); err != nil {
        t.Fatalf("join room: %v", err)
    }
    if err := sessions.DisconnectFromHub(ctx, hub, "session-17"); err != nil {
        t.Fatalf("disconnect session: %v", err)
    }
    if _, err := sessions.Heartbeat(ctx, "session-17"); err == nil {
        t.Fatal("heartbeat must reject a disconnected session")
    }
    presence, err := sessions.Presence(ctx, "doc-17")
    if err != nil {
        t.Fatalf("read session presence: %v", err)
    }
    if len(presence) != 0 || len(hub.Online("doc-17")) != 0 {
        t.Fatalf("disconnected session remained online: sessions=%#v room=%#v", presence, hub.Online("doc-17"))
    }
}
