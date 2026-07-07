package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed public/*
var files embed.FS

func Handler() http.Handler {
	sub, err := fs.Sub(files, "public")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}