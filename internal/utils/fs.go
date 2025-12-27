package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

type FS struct {
	root   string
	Images string
	ISOs   string
}

func NewFS(root string) *FS {
	root, err := filepath.Abs(root)
	if err != nil {
		panic(fmt.Sprintf("failed to create fs: %s", err))
	}

	fs := &FS{
		root:   root,
		Images: filepath.Join(root, "images"),
		ISOs:   filepath.Join(root, "templates", "iso"),
	}

	for _, p := range []string{
		fs.Images,
		fs.ISOs,
	} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			panic(fmt.Sprintf("failed to create fs: %s", err))
		}
	}

	return fs
}
