package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"coriolis-ovm-exporter/apiserver/params"
	"coriolis-ovm-exporter/config"

	jwt "github.com/dgrijalva/jwt-go"
)

// JWTClaims holds JWT claims
type JWTClaims struct {
	User string `json:"user"`
	jwt.StandardClaims
}

type jwtMiddleware struct {
	cfg *config.JWTAuth
}

// NewJWTMiddleware returns a populated jwtMiddleware
func NewJWTMiddleware(cfg *config.JWTAuth) (Middleware, error) {
	return &jwtMiddleware{
		cfg: cfg,
	}, nil
}

func invalidAuthResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		params.APIErrorResponse{
			Error:   "Authentication failed",
			Details: "Invalid authentication token",
		})
}

// Middleware implements the middleware interface
func (amw *jwtMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("authorization")
		if authorizationHeader == "" {
			invalidAuthResponse(w)
			return
		}

		bearerToken := strings.Split(authorizationHeader, " ")
		if len(bearerToken) != 2 {
			invalidAuthResponse(w)
			return
		}

		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(bearerToken[1], claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Invalid signing method")
			}
			return []byte(amw.cfg.Secret), nil
		})

		if err != nil {
			invalidAuthResponse(w)
			return
		}

		if token.Valid != true {
			invalidAuthResponse(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}
