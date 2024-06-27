package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"log"
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
	"github.com/koalatea/go-project-skeleton/oauthclient"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Server struct {
	client *ent.Client
}

func newServer(ctx context.Context, options ...func(*Server)) *Server {
	s := &Server{}
	for _, opt := range options {
		opt(s)
	}
	return s
}

func (srv *Server) Run(ctx context.Context) error {
	cfg := getConfig("server/nopush/config.json")
	router := http.NewServeMux()

	// Do not know if this actually does some tracing stuff or not. XSAM/otelsql though
	db, err := otelsql.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	// db, err := otelsql.Open("sqlite3", "file:server/nopush/db.sql?_fk=1")

	if err != nil {
		panic(err)
	}

	drv := entsql.OpenDB(dialect.SQLite, db)
	graph := ent.NewClient(ent.Driver(drv))

	// graph, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1") // TODO real graph db setup
	// if err != nil {
	// 	return err
	// }
	if err = graph.Schema.Create(
		context.Background(),
		migrate.WithGlobalUniqueID(true),
	); err != nil {
		// TODO proper logging
		fmt.Printf("failed to initialize graph schema: %w", err)
	}
	server := handler.NewDefaultServer(graphql.NewSchema(graph))
	server.Use(entgql.Transactioner{TxOpener: graph})
	server.Use(&debug.Tracer{})

	router.Handle("/graphql/playground", otelhttp.NewHandler(playground.Handler("playground", "/graphql"), "/graphql/playground"))
	router.Handle("/graphql",
		otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			server.ServeHTTP(w, req)
		}), "/graphql"))

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
		fmt.Printf("Failed to generate keys for usage in oauth flow: %s", err)
	}
	router.Handle("/oauth/login", oauthclient.NewOAuthLoginHandler(oauth, privKey))
	router.Handle("/oauth/authorize", oauthclient.NewOAuthAuthorizationHandler(oauth, pubKey, graph, "https://www.googleapis.com/oauth2/v3/userinfo"))

	// If performance profiling has been enabled, register the profiling routes
	if cfg.PProfEnabled {
		log.Printf("[WARN] Performance profiling is enabled, do not use in production as this may leak sensitive information")
		registerProfiler(router)
	}
	// run the Metric server and the authserver
	metricsHTTP := newMetricsServer()
	go func() {
		log.Printf("Metrics HTTP Server started on %s", metricsHTTP.Addr)
		if err := metricsHTTP.ListenAndServe(); err != nil {
			log.Printf("[WARN] stopped metrics http server: %v", err)
		}
	}()
	if err := http.ListenAndServe("0.0.0.0:8080", router); err != nil {
		return err
	}
	return nil
}
