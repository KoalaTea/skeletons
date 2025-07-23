package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/koalatea/go-project-skeleton/ent"
	"github.com/koalatea/go-project-skeleton/ent/user"
	"github.com/koalatea/go-project-skeleton/oauthclient"
)

// ErrInvalidViewer occurs when an invalid type of viewer is retrieved from the context.
// ErrNoViewer occurs when no viewer can be retrieved from the context.
var (
	ErrInvalidViewer     = fmt.Errorf("invalid viewer kind in context")
	ErrNoViewer          = fmt.Errorf("no authenticated viewer context")
	ErrReadingAuthCookie = fmt.Errorf("failed to read auth cookie")
	ErrInvalidAuthCookie = fmt.Errorf("invalid auth cookie")
)

// Viewer describes the query/mutation viewer-context.
type Viewer interface{}

type ctxKey struct{}

// FromContext returns the Viewer stored in a context.
func FromContext(ctx context.Context) Viewer {
	v, _ := ctx.Value(ctxKey{}).(Viewer)
	return v
}

// UserFromContext returns the user identity associated with the provided context, or nil if no user identity or a different identity type is associated.
func UserFromContext(ctx context.Context) *ent.User {
	val := ctx.Value(ctxKey{})
	u, ok := val.(*ent.User)
	if !ok || u == nil {
		return nil
	}
	return u
}

// NewContext returns a copy of parent context with the given Viewer attached with it.
func NewContext(parent context.Context, v Viewer) context.Context {
	return context.WithValue(parent, ctxKey{}, v)
}
func AuthenticationBypass(client *ent.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, err := client.User.Query().First(r.Context())
			if err != nil || u == nil {
				u = client.User.Create().
					SetName("auth-disabled").
					SetOAuthID("auth-disabled").
					SetIsActivated(true).
					SaveX(r.Context())
			}
			rWithUser := r.WithContext(NewContext(r.Context(), u))
			next.ServeHTTP(w, rWithUser)
		})
	}
}

// ContextFromAccessToken returns a copy of parent context with a user associated with it (if it exists).
func ContextFromSessionToken(ctx context.Context, graph *ent.Client, token string) (context.Context, error) {
	u, err := graph.User.Query().
		Where(user.SessionToken(token)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return NewContext(ctx, u), nil
}

func Authenticate(r *http.Request, graph *ent.Client) (context.Context, error) {
	authCookie, err := r.Cookie(oauthclient.SessionCookieName)
	if err != nil && err != http.ErrNoCookie {
		log.Printf("[ERROR] failed to read auth cookie: %v", err)
		return nil, ErrReadingAuthCookie
	}

	// If no auth cookie provided, do not authenticate the context
	if authCookie == nil {
		return r.Context(), nil
	}

	// Create an authenticated context (if provided cookie is valid)
	authCtx, err := ContextFromSessionToken(r.Context(), graph, authCookie.Value)
	if err != nil {
		log.Printf("failed to create session from auth cookie: %v", err)
		return nil, ErrInvalidAuthCookie
	}

	return authCtx, nil
}

func HandleUser(client *ent.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, err := Authenticate(r, client)
			if err != nil {
				switch err {
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
			rWithUser := r.WithContext(ctx)
			next.ServeHTTP(w, rWithUser)
		})
	}
}

func resetAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthclient.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	})
}

// IsAuthenticatedContext returns true if the context is associated with an authenticated identity, false otherwise.
func IsAuthenticated(ctx context.Context) bool {
	u := UserFromContext(ctx)
	return u != nil
}

// IsAuthenticatedContext returns true if the context is associated with an authenticated identity, false otherwise.
func IsActivated(ctx context.Context) bool {
	u := UserFromContext(ctx)
	if u == nil {
		return false
	}

	return u.IsActivated
}
