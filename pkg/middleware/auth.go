package middleware

import (
	"context"
	"fmt"
	"go_ecommerce/pkg/models"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// var jwtSecret = []byte(config.LoadConfig().JWTKey)
type TokenGenerator interface {
	GenerateJWT(user *models.User) (string, error)
	AuthenticateJWT(next http.Handler) http.Handler
}

type contextKey string

const userCtxKey contextKey = "user"

type Claims struct {
	User *models.User `json:"user"`
	jwt.StandardClaims
}

type Authenticator struct {
	jwtSecret []byte
}

func NewAuthenticator(key string) (*Authenticator, error) {
	if key == "" {
		return nil, fmt.Errorf("JWT key cannot be empty")
	}
	return &Authenticator{jwtSecret: []byte(key)}, nil
}

// GenerateJWT generates a new JWT token for a given user ID
func (a *Authenticator) GenerateJWT(user *models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		User: user,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(a.jwtSecret)
}

// AuthenticateJWT is a middleware function that checks for a valid JWT token in the request headers
func (a *Authenticator) AuthenticateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Authorization header must start with 'Bearer '", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return a.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			if err != nil {
				if ve, ok := err.(*jwt.ValidationError); ok {
					if ve.Errors&jwt.ValidationErrorMalformed != 0 {
						log.Printf("{middleware/AuthenticateJWT - Malformed token}")
					} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
						log.Printf("{middleware/AuthenticateJWT - Token is expired or not yet valid}")
					} else {
						log.Printf("{middleware/AuthenticateJWT - Couldn't handle this token: %v}", err)
					}
				}
				log.Printf("{middleware/AuthenticateJWT - Couldn't handle this token: %v}", err)

			}
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userCtxKey, claims.User)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext is a helper function to extract the user ID from the request context
func GetUserFromContext(r *http.Request) *models.User {
	user, ok := r.Context().Value(userCtxKey).(*models.User)
	if !ok {
		log.Println("No valid user found in context")
		return nil
	}
	return user
}
