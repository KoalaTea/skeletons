# from fresh
go mod init github.com/koalatea/go-project-skeleton
go get -d entgo.io/ent/cmd/ent
printf '//go:build tools\npackage tools\nimport (_ "github.com/99designs/gqlgen"\n _ "github.com/99designs/gqlgen/graphql/introspection")' | gofmt > tools.go
go mod tidy
go generate