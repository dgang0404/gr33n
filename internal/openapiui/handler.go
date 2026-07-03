// Package openapiui serves an offline-first OpenAPI browser at /openapi (Phase 116 WS5).
package openapiui

import (
	_ "embed"
	"net/http"
	"os"
	"strings"
)

//go:embed openapi.yaml
var specYAML []byte

//go:embed assets/redoc.standalone.js
var redocJS []byte

//go:embed index.html
var indexHTML []byte

var devBuild bool

// SetDevBuild marks whether the binary was compiled with -tags dev.
func SetDevBuild(dev bool) {
	devBuild = dev
}

// Enabled reports whether OpenAPI UI should be mounted (dev builds or OPENAPI_UI=true).
func Enabled() bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("OPENAPI_UI")))
	switch raw {
	case "true", "1", "on", "yes":
		return true
	case "false", "0", "off", "no":
		return false
	default:
		return devBuild
	}
}

// Register mounts GET /openapi, /openapi/spec.yaml, and /openapi/redoc.standalone.js on mux.
func Register(mux *http.ServeMux) {
	mux.Handle("GET /openapi", http.HandlerFunc(serveIndex))
	mux.Handle("GET /openapi/", http.HandlerFunc(serveIndex))
	mux.Handle("GET /openapi/spec.yaml", http.HandlerFunc(serveSpec))
	mux.Handle("GET /openapi/redoc.standalone.js", http.HandlerFunc(serveRedoc))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/openapi" && r.URL.Path != "/openapi/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

func serveSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.Write(specYAML)
}

func serveRedoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	if len(redocJS) == 0 {
		http.Error(w, "redoc bundle missing", http.StatusInternalServerError)
		return
	}
	w.Write(redocJS)
}
