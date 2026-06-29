package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// SPAHandler serves web/dist with fallback to index.html.
func SPAHandler(distDir string) http.Handler {
	fs := http.FileServer(http.Dir(distDir))
	indexPath := filepath.Join(distDir, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if strings.HasPrefix(p, "/api/") {
			http.NotFound(w, r)
			return
		}
		// 前端路由直链（/apis、/environments 等）统一回退 index.html
		if p == "/apis" || p == "/cases" || p == "/testdata" || p == "/impact" || p == "/mr" || p == "/scenarios" || p == "/reports" || p == "/environments" || p == "/login" {
			serveIndex(w, indexPath)
			return
		}
		clean := filepath.Clean(strings.TrimPrefix(p, "/"))
		if clean == "." {
			serveIndex(w, indexPath)
			return
		}
		target := filepath.Join(distDir, clean)
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			r.URL.Path = "/" + clean
			fs.ServeHTTP(w, r)
			return
		}
		serveIndex(w, indexPath)
	})
}

func serveIndex(w http.ResponseWriter, indexPath string) {
	b, err := os.ReadFile(indexPath)
	if err != nil {
		http.Error(w, "web UI not built: run cd web && npm run build", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(b)
}

func WebDistDir() string {
	return filepath.Join("web", "dist")
}
