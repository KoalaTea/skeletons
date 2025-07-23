package http

import "net/http"

// ServeMux used for endpoint registration.
type ServeMux interface {
	Handle(pattern string, handler http.Handler)
}

// A RouteMap contains a mapping of route patterns to http handlers.
type RouteMap map[string]http.Handler

// HandleFunc registers the handler function for the given pattern.
func (routes RouteMap) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	routes[pattern] = http.HandlerFunc(handler)
}

// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle panics.
func (routes RouteMap) Handle(pattern string, handler http.Handler) {
	routes[pattern] = handler
}

// TODO might move this to internal/http make handleFunc just pattern to handler so it is a very basic gathering of routes as they 'should'
// exist in another project. Though they have some prerequisites like auth still...
type RouteMapOptions func(*RouteMap)

func (routes RouteMap) Extend(rm RouteMap, options ...RouteMapOptions) {
	for pattern, handler := range rm {
		routes.Handle(pattern, handler)
	}
}
