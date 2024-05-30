package graphql

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/koalatea/go-project-skeleton/ent"
	"github.com/koalatea/go-project-skeleton/graphql/generated"
	"go.opentelemetry.io/otel"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

var tracer = otel.Tracer("authserver/graphql")

type Resolver struct {
	client *ent.Client
}

func NewSchema(client *ent.Client) graphql.ExecutableSchema {
	return generated.NewExecutableSchema(generated.Config{
		Resolvers: &Resolver{client},
	})
}
