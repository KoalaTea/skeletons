package www

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Handler is a custom handler for the single page react app - if the path doesn't exist the react app is returned.
type Handler struct {
	logger *log.Logger
}

// Content embedded from the application's build directory, includes the latest build of the UI.
//

//go:embed dist/*.html dist/*.svg
//go:embed dist/assets/*
var Content embed.FS

// ServeHTTP provides the Authserver UI, if the requested file does not exist it will serve index.html
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Serve the requested file
	path := fmt.Sprintf("dist/%s", strings.TrimPrefix(r.URL.Path, "/"))
	content, err := Content.ReadFile(path)
	if err == nil {
		if strings.HasSuffix(path, ".css") {
			w.Header().Add("Content-Type", "text/css")
		}
		if strings.HasSuffix(path, ".js") {
			w.Header().Add("Content-Type", "text/javascript")
		}
		w.Write(content)
		return
	}

	// Otherwise serve index.html
	index, err := Content.ReadFile("dist/index.html")
	if err != nil {
		h.logger.Printf("fatal error: failed to read index.html: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(index)
}

// NewHandler creates and returns a handler for the Authserver UI.
func NewHandler(logger *log.Logger) *Handler {
	return &Handler{logger}
}
