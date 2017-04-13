package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go" //JWT is not in the native Go packages
)

type Key int

const MyKey Key = 0

// JWT schema of the data it will store
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
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
	//30m Expiration for non-sensitive applications - OWSAP
	expireToken := time.Now().Add(time.Minute * 30).Unix()
	expireCookie := time.Now().Add(time.Minute * 30)

	//token Claims
	claims := Claims{
		"TestUser",
		jwt.StandardClaims{
			ExpiresAt: expireToken,
			Issuer:    "localhost:9000",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString([]byte("secret"))

	//Set Cookie parameters
	//Expires - 30m
	//HTTP only
	//Path
	//Domain
	cookie := http.Cookie{Name: "Auth", Value: signedToken, Expires: expireCookie, HttpOnly: true, Path: "/", Domain: "127.0.0.1", Secure: true}
	http.SetCookie(res, &cookie)
	http.Redirect(res, req, "/profile", 307)
}

// middleware to protect private pages
func validate(page http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		cookie, err := req.Cookie("Auth")
		if err != nil {
			res.Header().Set("Content-Type", "text/html")
			fmt.Fprint(res, "Unauthorized - Please login <br>")
			fmt.Fprintf(res, "<a href=\"login\"> Login </a>")
			return
		}

		token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method")
			}
			return []byte("secret"), nil
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
	})
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
	return
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
