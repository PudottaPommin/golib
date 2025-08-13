package http

import (
	"strings"
)

var wellKnownMimeTypesLower = map[string]string{
	".avif": "image/avif",
	".css":  "text/css; charset=utf-8",
	".gif":  "image/gif",
	".htm":  "text/html; charset=utf-8",
	".html": "text/html; charset=utf-8",
	".ico":  "image/x-icon",
	".jpeg": "image/jpeg",
	".jpg":  "image/jpeg",
	".js":   "text/javascript; charset=utf-8",
	".json": "application/json",
	".mjs":  "text/javascript; charset=utf-8",
	".pdf":  "application/pdf",
	".png":  "image/png",
	".svg":  "image/svg+xml",
	".wasm": "application/wasm",
	".webp": "image/webp",
	".xml":  "text/xml; charset=utf-8",

	// well, there are some types missing from the builtin list
	".txt": "text/plain; charset=utf-8",
}

func ResolveWellKnownMimeType(ext string) string {
	ext = strings.ToLower(ext)
	return wellKnownMimeTypesLower[ext]
}
