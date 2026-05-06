// Package webfs holds the embedded SvelteKit build output.
//
// The Makefile target `make web` populates webfs/dist with the contents of
// web/build before `go build`. Until then, dist/ contains a placeholder
// index.html that explains the situation.
package webfs

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var rawFS embed.FS

// FS returns a sub-filesystem rooted at the build output (no "dist/" prefix).
func FS() fs.FS {
	sub, err := fs.Sub(rawFS, "dist")
	if err != nil {
		// Fail closed — return an empty FS so the SPA handler shows its fallback.
		return embed.FS{}
	}
	return sub
}
