package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/debug"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/XSAM/otelsql"
	"github.com/koalatea/go-project-skeleton/ent"
	"github.com/koalatea/go-project-skeleton/ent/migrate"
	"github.com/koalatea/go-project-skeleton/graphql"
	internalHttp "github.com/koalatea/go-project-skeleton/internal/http"
	"github.com/koalatea/go-project-skeleton/oauthclient"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Server struct {
	client *ent.Client
}

func newServer(options ...func(*Server)) *Server {
	s := &Server{}
	for _, opt := range options {
		opt(s)
	}
	return s
}

func dbConnect() (*ent.Client, error) {
	// in memory
	mysqlDSN := "file:ent?mode=memory&cache=shared&_fk=1"
	// file on disk
	// mysqlDSN := "file:server/nopush/db.sql?_fk=1"

	// Do not know if this actually does some tracing stuff or not. XSAM/otelsql though
	db, err := otelsql.Open(dialect.SQLite, mysqlDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	drv := entsql.OpenDB(dialect.SQLite, db)
	graph := ent.NewClient(ent.Driver(drv))

	// non XSAM/otelsql
	// graph, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1") // TODO real graph db setup
	// if err != nil {
	// 	return err
	// }

	// TODO real setup of DB. This might be non problem because it may fail if already initialized so log failure continue
	if err = graph.Schema.Create(
		context.Background(),
		migrate.WithGlobalUniqueID(true),
	); err != nil {
		graph.Close()
		return nil, fmt.Errorf("failed to initialize graph schema: %w", err)
	}
	return graph, nil
}

func newGraphqlHandler(graph *ent.Client) http.Handler {
	server := handler.NewDefaultServer(graphql.NewSchema(graph))
	server.Use(entgql.Transactioner{TxOpener: graph})
	server.Use(&debug.Tracer{})
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		server.ServeHTTP(w, req)
	})
}

func (srv *Server) Run(ctx context.Context) error {
	cfg := getConfig("server/nopush/config.json")

	oauth := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.SecretKey,
		RedirectURL:  "http://localhost:8080/oauth/authorize",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate keys for usage in oauth flow", "err", err)
	}

	// Create Ent Client and Initialize graph schema
	graph, err := dbConnect()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to connect to the db", "err", err)
	}

	server := handler.NewDefaultServer(graphql.NewSchema(graph))
	server.Use(entgql.Transactioner{TxOpener: graph})
	server.Use(&debug.Tracer{})

	// httpLogger := log.New(os.Stderr, "[HTTP] ", log.Flags())
	routes := internalHttp.RouteMap{
		"/graphql/playground": internalHttp.Endpoint{
			Handler:              playground.Handler("playground", "/graphql"),
			AllowUnauthenticated: true,
		},
		"/graphql": internalHttp.Endpoint{
			Handler: newGraphqlHandler(graph),
		},
		"/oauth/login": internalHttp.Endpoint{
			Handler:              oauthclient.NewOAuthLoginHandler(oauth, privKey),
			AllowUnauthenticated: true,
		},
		"/oauth/authorize": internalHttp.Endpoint{
			Handler:              oauthclient.NewOAuthAuthorizationHandler(oauth, pubKey, graph, "https://www.googleapis.com/oauth2/v3/userinfo"),
			AllowUnauthenticated: true,
		},
		// // trailing slash is required to work with react
		// "/www/": internalHttp.Endpoint{
		// 	Handler: www.NewHandler(httpLogger),
		// },
	}

	// If performance profiling has been enabled, register the profiling routes
	if cfg.PProfEnabled {
		// TODO I think recording this as a guage metric that says DEVELOPMENT ONLY FEATURE ENABLED type thing could be interesting
		// Along with guages for each one following the same preset prefix so people can alert on the metric showing up in production
		// and see which features are enabled
		// Currently these would be bypass auth and performance profiling
		slog.WarnContext(ctx, "performance profiling is enabled, do not use in production as this may leak sensitive information")
		registerProfiler(routes)
	}
	// TODO use cfg.Bypassauth in some way
	router := internalHttp.NewServer(routes, internalHttp.WithAuthenticationBypass(graph))
	// run the Metric server and the main server
	metricsHTTP := newMetricsServer()
	httpServer := &http.Server{Addr: "0.0.0.0:8080", Handler: router}
	defer metricsHTTP.Shutdown(context.Background())
	defer graph.Close()
	defer httpServer.Close()
	go func() {
		slog.InfoContext(ctx, "Metrics HTTP started", "metrics_addr", metricsHTTP.Addr)
		if err := metricsHTTP.ListenAndServe(); err != nil {
			slog.WarnContext(ctx, "stopped metrics http server", "err", err)
		}
	}()
	slog.InfoContext(ctx, "AutherServer HTTP started", "http_addr", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("stopped http server: %w", err)
	}
	return nil
}
