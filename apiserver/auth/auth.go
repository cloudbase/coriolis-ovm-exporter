// Coriolis OVM exporter
// Copyright (C) 2021 Cloudbase Solutions SRL
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
