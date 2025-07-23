package http

import (
	"github.com/koalatea/go-project-skeleton/ent"
)

// An Option to configure a Tavern HTTP Server.
type Option func(*Server)

// WithAuthentication enables http request authentication for the server.
func WithAuthentication(graph *ent.Client) Option {
	return Option(func(server *Server) {
		server.Authenticator = &requestAuthenticator{graph}
	})
}

// WithAuthenticationBypass enables requests to bypass authentication for the server.
func WithAuthenticationBypass(graph *ent.Client) Option {
	return Option(func(server *Server) {
		server.Authenticator = &bypassAuthenticator{graph}
	})
}

// WithRequestLogging configures specialized HTTP request logging for the server, overriding the default logger.
func WithRequestLogging(logger Logger) Option {
	return Option(func(server *Server) {
		server.Logger = logger
	})
}
