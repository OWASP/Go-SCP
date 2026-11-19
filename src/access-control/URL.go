package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5" // Maintained JWT library (replaces deprecated dgrijalva/jwt-go)
)

type Key int

const MyKey Key = 0

// JWT schema of the data it will store
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// getJWTSecret retrieves the JWT secret from environment variables.
// SECURITY: Never hardcode secrets in source code.
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	return []byte(secret)
}

func Homepage(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html")
	fmt.Fprint(res, "Unauthorized - Please login <br>")
	fmt.Fprintf(res, "<a href=\"login\"> Login </a> <br>")
	fmt.Fprintf(res, "<a href=\"profile?page=page2\"> URL Parameters page 2 - Protected </a> <br>")
	fmt.Fprintf(res, "<a href=\"profile\"> Profile - Protected </a> <br>")
}

// create a JWT and put in the clients cookie
func setToken(res http.ResponseWriter, req *http.Request) {
	// 30m Expiration for non-sensitive applications - OWASP
	expireCookie := time.Now().Add(time.Minute * 30)

	// token Claims using jwt/v5 RegisteredClaims
	claims := Claims{
		Username: "TestUser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireCookie),
			Issuer:    "localhost:9000",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(getJWTSecret())
	if err != nil {
		http.Error(res, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error signing token: %v", err)
		return
	}

	// Set Cookie parameters
	// SameSite attribute helps prevent CSRF attacks
	cookie := http.Cookie{
		Name:     "Auth",
		Value:    signedToken,
		Expires:  expireCookie, // 30 min
		HttpOnly: true,         // Prevents JavaScript access
		Path:     "/",
		Domain:   "127.0.0.1",
		Secure:   true,     // Only sent over HTTPS
		SameSite: http.SameSiteStrictMode, // CSRF protection
	}
	http.SetCookie(res, &cookie)
	http.Redirect(res, req, "/profile", http.StatusTemporaryRedirect)
}

// middleware to protect private pages
func validate(page http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie("Auth")
		if err != nil {
			res.Header().Set("Content-Type", "text/html")
			fmt.Fprint(res, "Unauthorized - Please login <br>")
			fmt.Fprintf(res, "<a href=\"login\"> Login </a>")
			return
		}

		token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return getJWTSecret(), nil
		})
		if err != nil {
			res.Header().Set("Content-Type", "text/html")
			fmt.Fprint(res, "Unauthorized - Please login <br>")
			fmt.Fprintf(res, "<a href=\"login\"> Login </a>")
			return
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			ctx := context.WithValue(req.Context(), MyKey, *claims)
			page(res, req.WithContext(ctx))
		} else {
			res.Header().Set("Content-Type", "text/html")
			fmt.Fprint(res, "Unauthorized - Please login <br>")
			fmt.Fprintf(res, "<a href=\"login\"> Login </a>")
			return
		}
	}
}

// only viewable if the client has a valid token
func protectedProfile(res http.ResponseWriter, req *http.Request) {
	claims, ok := req.Context().Value(MyKey).(Claims)
	if !ok {
		res.Header().Set("Content-Type", "text/html")
		fmt.Fprint(res, "Unauthorized - Please login <br>")
		fmt.Fprintf(res, "<a href=\"login\"> Login </a>")
		return
	}
	url := req.URL.Query().Get("page")

	if url == "page2" {
		res.Header().Set("Content-Type", "text/html")
		fmt.Fprint(res, "PAGE 2 <br>")
		fmt.Fprintf(res, "<a href=\"logout\"> Logout </a>")
	} else {
		res.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(res, "Hello %s <br>", claims.Username)
		fmt.Fprintf(res, "<a href=\"logout\"> Logout </a>")
	}
}

// deletes the cookie
func logout(res http.ResponseWriter, req *http.Request) {
	deleteCookie := http.Cookie{Name: "Auth", Value: "none", Expires: time.Now()}
	http.SetCookie(res, &deleteCookie)
}

func main() {
	http.HandleFunc("/", Homepage)
	http.HandleFunc("/login", setToken)
	http.HandleFunc("/profile", validate(protectedProfile))
	http.HandleFunc("/logout", validate(logout))
	err := http.ListenAndServeTLS(":443", "cert/cert.pem", "cert/key.pem", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	//http.ListenAndServe(":9000", nil)
}
