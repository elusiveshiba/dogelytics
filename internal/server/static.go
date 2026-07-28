package server

import (
	"bytes"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	assets "github.com/dogeorg/dogelytics/img"
)

func registerFaviconRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/favicon.ico", serveFaviconFile("favicon.ico"))
	mux.HandleFunc("/apple-touch-icon.png", serveFaviconFile("apple-touch-icon.png"))
	mux.HandleFunc("/site.webmanifest", serveFaviconFile("site.webmanifest"))
	mux.HandleFunc("/favicons/", serveFaviconPath)
}

func registerImageRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/img/", serveImagePath)
}

func serveFaviconPath(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/favicons/")
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, `\`) {
		http.NotFound(w, r)
		return
	}

	serveFaviconFile(name)(w, r)
}

func serveFaviconFile(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serveEmbeddedAsset(w, r, path.Join("favicons", name))
	}
}

func serveImagePath(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/img/")
	if !safeImageName(name) {
		http.NotFound(w, r)
		return
	}

	serveEmbeddedAsset(w, r, name)
}

func serveEmbeddedAsset(w http.ResponseWriter, r *http.Request, name string) {
	data, err := fs.ReadFile(assets.Files, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if contentType := mime.TypeByExtension(filepath.Ext(name)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeContent(w, r, path.Base(name), time.Time{}, bytes.NewReader(data))
}

func safeImageName(name string) bool {
	if name == "" || strings.Contains(name, `\`) {
		return false
	}

	clean := filepath.Clean(name)
	if clean == "." || clean != name || filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return false
	}

	switch strings.ToLower(filepath.Ext(clean)) {
	case ".ico", ".png", ".svg", ".webmanifest":
		return true
	default:
		return false
	}
}
