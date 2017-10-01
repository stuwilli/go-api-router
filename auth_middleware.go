package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	wr "github.com/stuwilli/go-web-response"
)

//JWTSecretKey ...
const JWTSecretKey = "SECRETKEY"

type claimsKey string

//ClaimsKey ...
const ClaimsKey claimsKey = "claims"

//JWTAuth ...
func JWTAuth(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {

		tokenString, err := TokenFromHeader(r)

		if err != nil {
			rb := wr.NewBuilder().Status(http.StatusUnauthorized).
				Error(err).Build()
			rb.WriteJSON(w)
			return
		}

		token, err := ValidateToken(tokenString)

		if err != nil {
			rb := wr.NewBuilder().Status(http.StatusUnauthorized).
				Error(err).Build()
			rb.WriteJSON(w)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, ClaimsKey, token.Claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

//BuildJWTHandler ...
func BuildJWTHandler(config *ServiceConfig) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {

			tokenString, err := TokenFromHeader(r)

			if err != nil {
				rb := wr.NewBuilder().Status(http.StatusUnauthorized).
					Error(err).Build()
				rb.WriteJSON(w)
				return
			}

			token, err := ValidateToken(tokenString)

			if err != nil {
				rb := wr.NewBuilder().Status(http.StatusUnauthorized).
					Error(err).Build()
				rb.WriteJSON(w)
				return
			}

			if config.Auth.UseACM && !checkRole(config.Auth.RequiredRole, token) {

				rb := wr.NewBuilder().Status(http.StatusUnauthorized).
					Error(fmt.Errorf("Insufficient privileges")).Build()
				rb.WriteJSON(w)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)

	}
}

func checkRole(required []float64, token *jwt.Token) bool {

	if claims, ok := token.Claims.(jwt.MapClaims); ok {

		if _, ok := claims["role"]; ok {

			if contains(required, claims["role"].(float64)) {
				return true
			}
		}

	}
	return false
}

func contains(s []float64, e float64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//TokenFromHeader ...
func TokenFromHeader(r *http.Request) (string, error) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorization header is required")
	}

	authHeaderParts := strings.Split(authHeader, " ")

	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return "", fmt.Errorf("Authorization header format must be Bearer {token}")
	}

	return authHeaderParts[1], nil
}

//ValidateToken ...
func ValidateToken(tokenString string) (*jwt.Token, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(JWTSecretKey), nil
	})

	if !token.Valid {

		if ve, ok := err.(*jwt.ValidationError); ok {

			if ve.Errors&jwt.ValidationErrorMalformed != 0 {

				return nil, fmt.Errorf("Malformed token")

			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {

				return nil, fmt.Errorf("Expired or inactive token")

			} else {

				return nil, err
			}
		} else {

			return nil, err
		}
	}

	return token, nil

}
