package productweb

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Handler serves the product marketing SPA under /product/.
// When root is missing or disabled is true, returns a short HTML notice.
//
// Note: never pass "/index.html" to http.FileServer — net/http redirects that
// path to "./", which loops for SPA roots mounted at /product/.
func Handler(root string, enabled bool) http.Handler {
	if !enabled {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`<!doctype html><html><body style="font-family:system-ui;padding:2rem">
<h1>Product site disabled</h1>
<p>Set <code>PRODUCT_WEB_ENABLED=true</code> to enable marketing pages.</p>
</body></html>`))
		})
	}

	root = filepath.Clean(root)
	if _, err := os.Stat(root); err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`<!doctype html><html><body style="font-family:system-ui;padding:2rem">
<h1>Product web not built</h1>
<p>Run <code>make product-web</code> then restart the server.</p>
</body></html>`))
		})
	}

	index := filepath.Join(root, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(r.URL.Path, "/product")
		rel = strings.TrimPrefix(rel, "/")
		// Home and SPA client routes (no file extension) → index.html.
		if rel == "" || rel == "index.html" || !strings.Contains(filepath.Base(rel), ".") {
			http.ServeFile(w, r, index)
			return
		}
		// Static assets (_app/*, images/*, *.css, *.js, pre-rendered *.html, …).
		full := filepath.Join(root, filepath.Clean("/"+rel))
		if !strings.HasPrefix(full, root+string(os.PathSeparator)) && full != root {
			http.NotFound(w, r)
			return
		}
		if st, err := os.Stat(full); err != nil || st.IsDir() {
			http.ServeFile(w, r, index)
			return
		}
		http.ServeFile(w, r, full)
	})
}
