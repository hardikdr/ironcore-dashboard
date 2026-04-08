// Package root embeds the built Vue SPA from dist/frontend.
package root

import (
	"embed"
	"io/fs"
)

//go:embed dist/frontend
var frontendFS embed.FS

// Frontend returns a sub-filesystem rooted at dist/frontend.
func Frontend() fs.FS {
	sub, err := fs.Sub(frontendFS, "dist/frontend")
	if err != nil {
		panic("failed to sub frontendFS: " + err.Error())
	}
	return sub
}
