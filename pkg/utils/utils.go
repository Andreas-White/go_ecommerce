package utils

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
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
	if *err != nil {
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
