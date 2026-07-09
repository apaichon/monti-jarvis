package tenantweb

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Handler(root string) http.Handler {
	root = filepath.Clean(root)
	if _, err := os.Stat(root); err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`<!doctype html><html><body style="font-family:system-ui;padding:2rem">
<h1>Tenant portal not built</h1>
<p>Run <code>make tenant-web</code> then restart the server.</p>
</body></html>`))
		})
	}

	fileServer := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/tenant")
		if path == "" || path == "/" {
			path = "/index.html"
		}
		if !strings.Contains(filepath.Base(path), ".") {
			index := filepath.Join(root, "index.html")
			if _, err := os.Stat(index); err == nil {
				http.ServeFile(w, r, index)
				return
			}
		}
		r2 := r.Clone(r.Context())
		r2.URL.Path = path
		fileServer.ServeHTTP(w, r2)
	})
}