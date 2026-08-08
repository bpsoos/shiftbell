package static

import (
	"embed"
	"io/fs"
)

//go:embed assets/*
var content embed.FS

func Assets() (fs.FS, error) {
	return fs.Sub(content, "assets")
}
