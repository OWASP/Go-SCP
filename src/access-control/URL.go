// Command url demonstrates access control on URLs and on a direct object
// reference (the page query parameter). It builds on the session example in
// ../session-management.
//
// Run it with the signing secret in the environment and a TLS certificate in
// cert/:
//
//	JWT_SECRET=... go run . -addr :8443
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	cookieName = "Auth"
	// sessionTimeout is short because OWASP recommends 15-30 minutes for
	// low-risk applications.
	sessionTimeout = 30 * time.Minute

	// Server timeouts stop slow clients from holding connections open
	// (see the HTTP Server Security section).
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 2 * time.Minute
	shutdownTimeout   = 10 * time.Second
)

// claimsKey is the context key under which requireSession stores the
// verified Claims.
type claimsKey struct{}

// Claims is the JWT payload stored in the session cookie.
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM,
	)
	defer stop()

	err := run(ctx, os.Args, os.Getenv, os.Stderr)
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		stop()
		os.Exit(1)
	}
}

// run serves the example over HTTPS until ctx is cancelled.
func run(ctx context.Context, args []string, getenv func(string) string, stderr io.Writer) error {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(stderr)
	addr := flags.String("addr", ":443", "address to listen on")
	certFile := flags.String("cert", "cert/cert.pem", "TLS certificate file")
	keyFile := flags.String("key", "cert/key.pem", "TLS private key file")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	// SECURITY: Never hardcode secrets in source code.
	secret := getenv("JWT_SECRET")
	if secret == "" {
		return errors.New("JWT_SECRET environment variable is required")
	}
	logger := slog.New(slog.NewTextHandler(stderr, nil))

	srv := &http.Server{
		Addr:              *addr,
		Handler:           newServer(logger, []byte(secret)),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
	return serveTLS(ctx, srv, *certFile, *keyFile)
}

// newServer returns the example's routes.
func newServer(logger *slog.Logger, secret []byte) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /{$}", handleHomepage())
	mux.Handle("GET /profile", requireSession(secret, handleProfile()))
	mux.Handle("GET /login", handleLogin(logger, secret))
	mux.Handle("POST /logout", requireSession(secret, handleLogout()))
	mux.Handle("GET /healthz", handleHealthz())
	mux.Handle("/", http.NotFoundHandler())
	return mux
}

// handleLogin creates a JWT and puts it in the client's cookie.
func handleLogin(logger *slog.Logger, secret []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expires := time.Now().Add(sessionTimeout)
		claims := Claims{
			Username: "TestUser",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expires),
				Issuer:    "localhost:9000",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(secret)
		if err != nil {
			// Never ignore errors from cryptographic operations.
			logger.ErrorContext(r.Context(), "signing token", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, sessionCookie(signedToken, int(sessionTimeout.Seconds())))
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	})
}

// handleHomepage is public and links to the protected pages.
func handleHomepage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeHTML(w, `Unauthorized - Please login <br>
<a href="/login">Login</a> <br>
<a href="/profile?page=page2">URL Parameters page 2 - Protected</a> <br>
<a href="/profile">Profile - Protected</a> <br>`)
	})
}

// handleProfile is only reachable with a valid session.
func handleProfile() http.Handler {
	// protectedPages are the only values the page parameter may reference.
	// Looking the parameter up in an allowlist, rather than using it
	// directly, keeps a user from reaching objects they were not meant to see.
	protectedPages := map[string]string{
		"page2": "PAGE 2",
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(claimsKey{}).(Claims)
		if !ok {
			writeLoginPage(w)
			return
		}

		page := r.URL.Query().Get("page")
		if page == "" {
			writeHTML(w, "Hello "+html.EscapeString(claims.Username)+" <br>"+logoutForm)
			return
		}
		title, ok := protectedPages[page]
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeHTML(w, title+" <br>"+logoutForm)
	})
}

// handleLogout deletes the session cookie.
func handleLogout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A cookie is only replaced if Name, Path and Domain match, so build
		// it the same way as at login. A negative MaxAge deletes it.
		http.SetCookie(w, sessionCookie("", -1))
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

func handleHealthz() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// requireSession lets the request through to next only if it carries a
// session cookie with a valid token signed with secret. It stores the
// token's Claims in the request context.
func requireSession(secret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			writeLoginPage(w)
			return
		}

		var claims Claims
		token, err := jwt.ParseWithClaims(
			cookie.Value,
			&claims,
			func(token *jwt.Token) (any, error) {
				// Reject tokens signed with any other algorithm, such as "none".
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return secret, nil
			},
		)
		if err != nil || !token.Valid {
			writeLoginPage(w)
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// sessionCookie returns the session cookie holding token, which the browser
// keeps for maxAge seconds. Max-Age is relative, so unlike Expires it does
// not depend on the client's clock.
func sessionCookie(token string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,                    // Prevents JavaScript access (mitigates XSS)
		Secure:   true,                    // Only sent over HTTPS
		SameSite: http.SameSiteStrictMode, // CSRF protection
	}
}

const logoutForm = `<form method="post" action="/logout"><button>Logout</button></form>`

func writeLoginPage(w http.ResponseWriter) {
	writeHTML(w, `Unauthorized - Please login <br><a href="/login">Login</a>`)
}

func writeHTML(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, body)
}

// serveTLS runs srv until ctx is cancelled, then shuts it down gracefully.
func serveTLS(ctx context.Context, srv *http.Server, certFile, keyFile string) error {
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServeTLS(certFile, keyFile) }()

	select {
	case err := <-errc:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
