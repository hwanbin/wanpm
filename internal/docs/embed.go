package docs

import (
	"embed"
	"io/fs"
)

//go:embed all:swagger-ui
var assets embed.FS

//go:embed openapi.yaml
var OpenAPISpec []byte

func Assets() (fs.FS, error) {
	return fs.Sub(assets, "swagger-ui")
}
