package http

import (
	"net/http"
)

// A Server for Tavern HTTP traffic.
type Server struct {
	Logger
	http.Handler
	Authenticator
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Authenticate Request (if possible)
	ctx, err := srv.Authenticate(r)
	if err != nil {
		switch err {
		case ErrInvalidAccessToken:
			http.Error(w, "invalid access token", http.StatusUnauthorized)
			return
		case ErrInvalidAuthCookie:
			resetAuthCookie(w)
			http.Error(w, "invalid auth cookie", http.StatusUnauthorized)
			return
		case ErrReadingAuthCookie:
			resetAuthCookie(w)
			http.Error(w, "failed to read auth cookie", http.StatusBadRequest)
			return
		default:
			resetAuthCookie(w)
			http.Error(w, "unexpected error occurred", http.StatusInternalServerError)
			return
		}
	}
	r = r.WithContext(ctx)

	// Log Request
	if srv.Logger != nil {
		srv.Log(r)
	}

	// Handle Request
	srv.Handler.ServeHTTP(w, r)
}

// NewServer configures a new server for serving HTTP traffic.
func NewServer(routes RouteMap, options ...Option) *Server {
	// Register routes
	router := http.NewServeMux()
	for route, handler := range routes {
		router.Handle(route, addHttpTelemetry(route, handler))
	}

	// Apply Options
	server := &Server{
		Handler:       router,
		Logger:        defaultRequestLogger{},
		Authenticator: &requestAuthenticator{},
	}
	for _, opt := range options {
		opt(server)
	}

	return server
}
