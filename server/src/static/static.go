package static

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed all:dist
var staticFS embed.FS

var fileSystem http.FileSystem

func init() {
	subFS, err := fs.Sub(staticFS, "dist")
	if err != nil {
		panic(err)
	}
	fileSystem = http.FS(subFS)
}

func GetFileSystem() http.FileSystem {
	return fileSystem
}

func ServeFile(w http.ResponseWriter, r *http.Request, fileSystem http.FileSystem, name string) {
	f, err := fileSystem.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if stat.IsDir() {
		indexName := strings.TrimSuffix(name, "/") + "/index.html"
		indexFile, err := fileSystem.Open(indexName)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer indexFile.Close()
		stat, err = indexFile.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		f = indexFile
	}

	ext := filepath.Ext(name)
	var contentType string
	switch ext {
	case ".html":
		contentType = "text/html; charset=utf-8"
	case ".css":
		contentType = "text/css; charset=utf-8"
	case ".js":
		contentType = "application/javascript; charset=utf-8"
	case ".json":
		contentType = "application/json; charset=utf-8"
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".svg":
		contentType = "image/svg+xml"
	case ".ico":
		contentType = "image/x-icon"
	case ".woff":
		contentType = "font/woff"
	case ".woff2":
		contentType = "font/woff2"
	case ".ttf":
		contentType = "font/ttf"
	case ".eot":
		contentType = "application/vnd.ms-fontobject"
	default:
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	http.ServeContent(w, r, name, stat.ModTime(), f)
}
