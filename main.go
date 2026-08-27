package main

import (
	"context"
	"github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	app := shared.NewApp()
	srv := &http.Server{Addr: getenv("HTTP_ADDR", ":8080"), Handler: frontendHandler(app.Router(), os.Getenv("WEB_DIR"))}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
func getenv(k, f string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return f
}

func frontendHandler(api http.Handler, webDir string) http.Handler {
	if webDir == "" {
		return api
	}
	index := filepath.Join(webDir, "index.html")
	if _, err := os.Stat(index); err != nil {
		log.Printf("frontend assets unavailable: %v", err)
		return api
	}
	assets := http.FileServer(http.Dir(webDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/healthz" {
			api.ServeHTTP(w, r)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}
		cleanPath := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if info, err := os.Stat(filepath.Join(webDir, filepath.FromSlash(cleanPath))); err == nil && !info.IsDir() {
			assets.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, index)
	})
}
