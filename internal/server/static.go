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

func registerFaviconRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/favicon.ico", serveFaviconFile("favicon.ico"))
	mux.HandleFunc("/apple-touch-icon.png", serveFaviconFile("apple-touch-icon.png"))
	mux.HandleFunc("/site.webmanifest", serveFaviconFile("site.webmanifest"))
	mux.HandleFunc("/favicons/", serveFaviconPath)
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
