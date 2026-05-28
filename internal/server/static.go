package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var faviconDirs = []string{
	"img/favicons",
	"../../img/favicons",
	"/app/img/favicons",
}

var imageDirs = []string{
	"img",
	"../../img",
	"/app/img",
}

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
		for _, dir := range faviconDirs {
			path := filepath.Join(dir, name)
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				http.ServeFile(w, r, path)
				return
			}
		}

		http.NotFound(w, r)
	}
}

func serveImagePath(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/img/")
	if !safeImageName(name) {
		http.NotFound(w, r)
		return
	}

	for _, dir := range imageDirs {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}
	}

	http.NotFound(w, r)
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
