package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestProfileRequiresValidSession(t *testing.T) {
	secret := []byte("test-secret")
	handler := newServer(slog.New(slog.DiscardHandler), secret)

	sign := func(method jwt.SigningMethod, key any, expires time.Time) string {
		t.Helper()
		claims := Claims{
			Username: "<b>TestUser</b>",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expires),
			},
		}
		token, err := jwt.NewWithClaims(method, claims).SignedString(key)
		if err != nil {
			t.Fatalf("signing: %s", err)
		}
		return token
	}
	future := time.Now().Add(time.Hour)
	const maxAge = 3600
	const unauthorized = "Unauthorized"

	cases := map[string]struct {
		cookie   *http.Cookie
		wantBody string
	}{
		"valid token": {
			cookie:   sessionCookie(sign(jwt.SigningMethodHS256, secret, future), maxAge),
			wantBody: "Hello &lt;b&gt;TestUser&lt;/b&gt;",
		},
		"no cookie": {
			wantBody: unauthorized,
		},
		"signed with another key": {
			cookie:   sessionCookie(sign(jwt.SigningMethodHS256, []byte("other"), future), maxAge),
			wantBody: unauthorized,
		},
		"unsigned token (alg none)": {
			cookie: sessionCookie(
				sign(jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, future),
				maxAge,
			),
			wantBody: unauthorized,
		},
		"expired token": {
			cookie: sessionCookie(
				sign(jwt.SigningMethodHS256, secret, time.Now().Add(-time.Minute)),
				maxAge,
			),
			wantBody: unauthorized,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			body, err := io.ReadAll(rec.Result().Body)
			if err != nil {
				t.Fatalf("reading body: %s", err)
			}
			if !strings.Contains(string(body), tc.wantBody) {
				t.Fatalf("expected body containing %q, got %q", tc.wantBody, body)
			}
		})
	}
}

func TestLoginThenLogout(t *testing.T) {
	handler := newServer(slog.New(slog.DiscardHandler), []byte("test-secret"))

	login := httptest.NewRecorder()
	handler.ServeHTTP(
		login,
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/login", nil),
	)
	cookies := login.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != cookieName || cookies[0].Value == "" {
		t.Fatalf("expected one %s cookie after login, got %v", cookieName, cookies)
	}
	session := cookies[0]
	if !session.Secure || !session.HttpOnly || session.SameSite != http.SameSiteStrictMode {
		t.Fatalf("session cookie missing security attributes: %+v", session)
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/logout", nil)
	req.AddCookie(session)
	logout := httptest.NewRecorder()
	handler.ServeHTTP(logout, req)
	cookies = logout.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge >= 0 || cookies[0].Path != session.Path {
		t.Fatalf("expected logout to expire the session cookie, got %v", cookies)
	}
}
