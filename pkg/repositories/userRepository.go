package repositories

import (
	"database/sql"
	"fmt"
	"go_ecommerce/pkg/database"
	models "go_ecommerce/pkg/models/user"
	"go_ecommerce/pkg/utils"
	"log"
	"time"

	"github.com/google/uuid"
)

// UserRepository struct
type UserRepository struct {
	DB *sql.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		DB: database.DB,
	}
}

// CreateUser inserts a new user into the database
func (r *UserRepository) CreateUser(user *models.UserRegister) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	defer utils.HandleTransaction(tx, &err)

	hashedPassword := utils.HashPassword(user.Password)

	userID := uuid.New().String()

	_, err = tx.Exec(`
        INSERT INTO users (id, name, email, phone)
        VALUES ($1, $2, $3, $4)`,
		userID, user.Name, user.Email, user.Phone)

	if err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}

	authID := uuid.New().String()
	creationDate := time.Now().Format("2006-01-02")
	active := true

	_, err = tx.Exec(`
	INSERT INTO auths (id, user_id, creation_date, active, password)
	VALUES ($1, $2, $3, $4, $5)`,
		authID, userID, creationDate, active, hashedPassword)

	if err != nil {
		return fmt.Errorf("error inserting auth: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by their ID
func (r *UserRepository) GetUserByID(id string) (*models.User, error) {
	query := `
		SELECT * FROM users WHERE id = $1
	`
	user, err := r.fetchUserByValue(query, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByName retrieves a user by their name
func (r *UserRepository) GetUserByName(name string) (*models.User, error) {
	query := `
		SELECT * FROM users WHERE name = $1
	`
	user, err := r.fetchUserByValue(query, name)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT * FROM users WHERE email = $1
	`

	user, err := r.fetchUserByValue(query, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email
func (r *UserRepository) GetAuthedUserByEmail(email string) (*models.AuthedUser, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	queryUsers := `
		SELECT * FROM users WHERE email = $1
	`

	var user models.User
	err = r.DB.QueryRow(queryUsers, email).Scan(&user.ID, &user.Name, &user.Email, &user.Phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error retrieving user: %v", err)
	}

	defer utils.HandleTransaction(tx, &err)

	queryAuths := `
		SELECT * FROM auths WHERE user_id = $1
	`

	var auth models.Auth
	err = r.DB.QueryRow(queryAuths, user.ID).Scan(&auth.ID, &auth.UserID, &auth.CreationDate, &auth.Active, &auth.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error retrieving user: %v", err)
	}

	authedUser := &models.AuthedUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
		Auth:  auth,
	}
	return authedUser, nil
}

// UpdateUser updates an existing user's information
func (r *UserRepository) UpdateUser(user *models.User) error {
	query := `
		UPDATE users SET name = $2, email = $3, phone = $4 WHERE id = $1
	`

	_, err := r.DB.Exec(query, user.ID, user.Name, user.Email, user.Phone)
	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}

	log.Printf("User %s updated successfully\n", user.Name)
	return nil
}

// DeleteUser removes a user from the database
func (r *UserRepository) DeleteUser(id string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	queryAuth := `
		DELETE FROM auths WHERE user_id = $1
	`

	_, err = r.DB.Exec(queryAuth, id)
	if err != nil {
		return fmt.Errorf("error deleting auth for user: %v", err)
	}

	defer utils.HandleTransaction(tx, &err)

	queryUser := `
	DELETE FROM users WHERE id = $1
`

	_, err = r.DB.Exec(queryUser, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	}

	log.Printf("User with ID %s deleted successfully\n", id)
	return nil
}

func (r *UserRepository) fetchUserByValue(query string, value string) (*models.User, error) {
	var user models.User
	err := r.DB.QueryRow(query, value).Scan(&user.ID, &user.Name, &user.Email, &user.Phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error retrieving user: %v", err)
	}

	return &user, nil
}
