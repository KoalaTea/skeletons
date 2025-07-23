package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/koalatea/go-project-skeleton/ent/enttest"
	internalHttp "github.com/koalatea/go-project-skeleton/internal/http"
	"github.com/koalatea/go-project-skeleton/oauthclient"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestRequestAuthenticator(t *testing.T) {
	// Setup Dependencies
	ctx := context.Background()
	var (
		driverName     = "sqlite3"
		dataSourceName = "file:ent?mode=memory&cache=shared&_fk=1"
	)
	graph := enttest.Open(t, driverName, dataSourceName, enttest.WithOptions())
	defer graph.Close()
	srv := internalHttp.NewServer(
		internalHttp.RouteMap{
			"/": http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		},
		internalHttp.WithAuthentication(graph),
	)

	// Test Data
	existingAdmin := graph.User.Create().
		SetName("test_admin").
		SetOAuthID("test_admin").
		SetIsActivated(true).
		SaveX(ctx)

	// Test Cases
	tests := []struct {
		name     string
		req      *http.Request
		cookies  []*http.Cookie
		headers  http.Header
		wantCode int
	}{
		{
			name: "SessionCookie",
			req:  httptest.NewRequest(http.MethodGet, "/test", nil),
			cookies: []*http.Cookie{{
				Name:  oauthclient.SessionCookieName,
				Value: existingAdmin.SessionToken,
			}},
			headers:  http.Header{},
			wantCode: http.StatusOK,
		},
		// {
		// 	name:    "AccessToken",
		// 	req:     httptest.NewRequest(http.MethodGet, "/test", nil),
		// 	cookies: []*http.Cookie{},
		// 	headers: http.Header{
		// 		auth.HeaderAPIAccessToken: []string{existingAdmin.AccessToken},
		// 	},
		// 	wantCode: http.StatusOK,
		// },
		{
			name:    "Unauthenticated",
			req:     httptest.NewRequest(http.MethodGet, "/test", nil),
			cookies: []*http.Cookie{},
			headers: http.Header{},

			// This is allowed, to enable unauthenticated endpoints (e.g. login endpoint).
			// Endpoints rely on auth.IsAuthenticatedContext() and related methods for authorization.
			wantCode: http.StatusOK,
		},
		{
			name: "InvalidSessionCookie",
			req:  httptest.NewRequest(http.MethodGet, "/test", nil),
			cookies: []*http.Cookie{{
				Name:  oauthclient.SessionCookieName,
				Value: "already_eaten",
			}},
			headers:  http.Header{},
			wantCode: http.StatusUnauthorized,
		},
		// 	{
		// 		name:    "InvalidAccessToken",
		// 		req:     httptest.NewRequest(http.MethodGet, "/test", nil),
		// 		cookies: []*http.Cookie{},
		// 		headers: http.Header{
		// 			auth.HeaderAPIAccessToken: []string{"1337"},
		// 		},
		// 		wantCode: http.StatusUnauthorized,
		// 	},
		// 	{
		// 		name: "InvalidAccessTokenWithValidSession",
		// 		req:  httptest.NewRequest(http.MethodGet, "/test", nil),
		// 		cookies: []*http.Cookie{{
		// 			Name:  oauthclient.SessionCookieName,
		// 			Value: existingAdmin.SessionToken,
		// 		}},
		// 		headers: http.Header{
		// 			auth.HeaderAPIAccessToken: []string{"1337"},
		// 		},
		// 		wantCode: http.StatusUnauthorized,
		// 	},
	}

	// Run Tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare Request
			tc.req.Header = tc.headers
			for _, cookie := range tc.cookies {
				tc.req.AddCookie(cookie)
			}

			// Serve HTTP
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, tc.req)
			resp := w.Result()

			// Assertions
			assert.Equal(t, tc.wantCode, resp.StatusCode)
		})
	}
}
