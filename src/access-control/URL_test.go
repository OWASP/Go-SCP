package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProfileAccessControl(t *testing.T) {
	handler := newServer(slog.New(slog.DiscardHandler), []byte("test-secret"))

	// Log in once to get a valid session cookie.
	login := httptest.NewRecorder()
	handler.ServeHTTP(
		login,
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/login", nil),
	)
	cookies := login.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one cookie after login, got %v", cookies)
	}
	session := cookies[0]

	cases := map[string]struct {
		target     string
		withCookie bool
		wantStatus int
		wantBody   string
	}{
		"profile without session": {
			target:     "/profile",
			wantStatus: http.StatusOK,
			wantBody:   "Unauthorized",
		},
		"allowed page without session": {
			target:     "/profile?page=page2",
			wantStatus: http.StatusOK,
			wantBody:   "Unauthorized",
		},
		"profile with session": {
			target:     "/profile",
			withCookie: true,
			wantStatus: http.StatusOK,
			wantBody:   "Hello TestUser",
		},
		"allowed page with session": {
			target:     "/profile?page=page2",
			withCookie: true,
			wantStatus: http.StatusOK,
			wantBody:   "PAGE 2",
		},
		"page outside the allowlist": {
			target:     "/profile?page=../admin",
			withCookie: true,
			wantStatus: http.StatusNotFound,
			wantBody:   "not found",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.target, nil)
			if tc.withCookie {
				req.AddCookie(session)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			resp := rec.Result()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("reading body: %s", err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, resp.StatusCode)
			}
			if !strings.Contains(string(body), tc.wantBody) {
				t.Fatalf("expected body containing %q, got %q", tc.wantBody, body)
			}
		})
	}
}
