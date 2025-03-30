package middleware

import (
	"context"
	"go_ecommerce/internal/config"
	"go_ecommerce/pkg/models"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Your secret key, used to sign JWT tokens (this should be stored securely)
var jwtSecret = []byte(config.LoadConfig().JWTKey)

// Claims defines the structure of the JWT claims
type Claims struct {
	User *models.User `json:"user"`
	jwt.StandardClaims
}

// GenerateJWT generates a new JWT token for a given user ID
func GenerateJWT(user *models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		User: user,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Create the JWT token with the specified signing method and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token using the secret key
	return token.SignedString(jwtSecret)
}

// AuthenticateJWT is a middleware function that checks for a valid JWT token in the request headers
func AuthenticateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Bearer token should be in the format: "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		// If token is invalid, return unauthorized error
		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Attach the user ID to the request context for use in the handler
		ctx := context.WithValue(r.Context(), "user", claims.User)

		// Call the next handler, passing the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext is a helper function to extract the user ID from the request context
func GetUserFromContext(r *http.Request) *models.User {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		log.Println("No valid user found in context")
	}
	return user
}
