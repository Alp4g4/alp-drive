package static

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"net/http"
)

//go:embed app.css app.js htmx.min.js assets
var files embed.FS

var versions = hashFiles()

func hashFiles() map[string]string {
	v := map[string]string{}
	_ = fs.WalkDir(files, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, _ := fs.ReadFile(files, name)
		sum := sha256.Sum256(data)
		v[name] = hex.EncodeToString(sum[:4])
		return nil
	})
	return v
}

func URL(name string) string {
	return "/static/" + name + "?v=" + versions[name]
}

func Handler() http.Handler {
	fileServer := http.FileServerFS(files)
	return http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("v") != "" {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		fileServer.ServeHTTP(w, r)
	}))
}
