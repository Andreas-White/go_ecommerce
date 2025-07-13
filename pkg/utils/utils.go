package utils

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"
)

// RespondWithError sends an error response in JSON format
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}

// RespondWithJSON sends a response in JSON format
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// HandleTransaction is a reusable function to handle transaction commit or rollback
func HandleTransaction(tx *sql.Tx, err *error) {
	if p := recover(); p != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			*err = fmt.Errorf("transaction rollback failed: %v, original error: %w", rollbackErr, *err)
		}
	} else if *err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			*err = fmt.Errorf("transaction rollback failed: %v, original error: %w", rollbackErr, *err)
		}
	} else {
		commitErr := tx.Commit()
		if commitErr != nil {
			*err = fmt.Errorf("transaction commit failed: %w", commitErr)
		}
	}
}

// hashPassword hashes the password using SHA-256
func HashPassword(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

// CheckPasswordHash compares a plain-text password with its hashed version stored in the database
func CheckPasswordHash(password, hashedPassword string) bool {
	hashedInputPassword := HashPassword(password)
	return hashedInputPassword == hashedPassword
}

// IsValidEmail checks if the email format is valid
func IsValidEmail(email string) bool {
	trimmedEmail := strings.TrimSpace(email)
	if trimmedEmail == "" {
		return false
	}

	addr, err := mail.ParseAddress(trimmedEmail)
	if err != nil {
		return false
	}

	emailToValidate := addr.Address
	if emailToValidate == "" {
		return false
	}

	parts := strings.SplitN(emailToValidate, "@", 2)
	if len(parts) != 2 {
		return false
	}
	localPart := parts[0]
	domainPart := parts[1]

	if localPart == "" || domainPart == "" {
		return false
	}
	if !strings.Contains(domainPart, ".") {
		return false
	}
	if strings.HasPrefix(domainPart, "-") || strings.HasSuffix(domainPart, "-") ||
		strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") {
		return false
	}
	if strings.Contains(domainPart, "..") || strings.Contains(domainPart, ".-") || strings.Contains(domainPart, "-.") {
		return false
	}
	if strings.Contains(domainPart, "..") || strings.Contains(domainPart, ".-") || strings.Contains(domainPart, "-.") {
		return false
	}

	if len(localPart) > 64 {
		return false
	}
	if len(domainPart) > 255 {
		return false
	}
	if len(emailToValidate) > 254 {
		return false
	}

	return true
}

// GenerateCSRFToken generates a secure random string for CSRF protection
func GenerateCSRFToken(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func HandleAPIErrors(err error, w http.ResponseWriter, sourceFunc string, httpCode int, genericErrorMessage string) {
	if errors.Is(err, context.Canceled) {
		log.Printf("\n{%s - Request cancelled: %v}", sourceFunc, err)
		RespondWithError(w, http.StatusRequestTimeout, "Request cancelled")
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		log.Printf("\n{%s - Request timed out: %v}", sourceFunc, err)
		RespondWithError(w, http.StatusGatewayTimeout, "Operation timed out")
		return
	}
	log.Printf("\n{%s - error: %v}", sourceFunc, err)
	RespondWithError(w, httpCode, genericErrorMessage)
}

func HandleServiceErrors(ctx context.Context, err error, sourceFunc string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("\n{%s - context error: %w}", sourceFunc, ctx.Err())
	}
	return fmt.Errorf("\n{%s - error: %w}", sourceFunc, err)
}

func HandleRepositoryErrors(ctx context.Context, err error, sourceFunc string, userIdentifier string) error {
	if err == sql.ErrNoRows {
		return fmt.Errorf("\n{%s - not found: %w, user_identifier: %s}", sourceFunc, err, userIdentifier)
	}
	if ctx.Err() != nil {
		return fmt.Errorf("\n{%s - context error : %w, user_identifier: %s}", sourceFunc, ctx.Err(), userIdentifier)
	}
	return fmt.Errorf("\n{%s - error : %w, user_identifier: %s}", sourceFunc, err, userIdentifier)
}
