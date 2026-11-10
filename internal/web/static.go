package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/index.html
var files embed.FS

func Handler() http.Handler {
	staticFiles, err := fs.Sub(files, "static")
	if err != nil {
		panic(err)
	}

	return http.FileServer(http.FS(staticFiles))
}
