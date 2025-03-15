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
        INSERT INTO users (id, first_name, last_name, middle_name, email, phone, is_producer)
        VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, user.FirstName, user.LastName, user.MiddleName, user.Email, user.Phone, user.IsProducer)

	if err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}

	authID := uuid.New().String()
	created_at := time.Now()
	active := true

	_, err = tx.Exec(`
	INSERT INTO auths (id, user_id, created_at, active, password)
	VALUES ($1, $2, $3, $4, $5)`,
		authID, userID, created_at, active, hashedPassword)

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
func (r *UserRepository) GetUserByFullName(firstName string, lastName string, middleName string) (*models.User, error) {
	query := `
		SELECT * FROM users WHERE first_name = $1 AND last_name = $2 AND middle_name = $3 
	`
	var user models.User
	err := r.DB.QueryRow(query, firstName, lastName, middleName).Scan(&user.ID, &user.FirstName, &user.Email, &user.Phone, &user.LastName, &user.MiddleName, &user.IsProducer)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error retrieving user: %v", err)
	}

	return &user, nil
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
	err = r.DB.QueryRow(queryUsers, email).Scan(&user.ID, &user.FirstName, &user.Email, &user.Phone, &user.LastName, &user.MiddleName, &user.IsProducer)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("user not found")
			return nil, fmt.Errorf("user not found")
		}
		log.Printf("error retrieving user: %v", err)
		return nil, fmt.Errorf("error retrieving user: %v", err)
	}

	defer utils.HandleTransaction(tx, &err)

	queryAuths := `
		SELECT * FROM auths WHERE user_id = $1 AND active = $2
	`

	var auth models.Auth
	err = r.DB.QueryRow(queryAuths, user.ID, true).Scan(&auth.ID, &auth.UserID, &auth.CreatedAt, &auth.Active, &auth.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("auth not found %v", err)
			return nil, fmt.Errorf("auth not found")
		}
		log.Printf("error retrieving auth: %v", err)
		return nil, fmt.Errorf("error retrieving auth: %v", err)
	}

	authedUser := &models.AuthedUser{
		ID:         user.ID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Email:      user.Email,
		Phone:      user.Phone,
		IsProducer: user.IsProducer,
		Auth:       auth,
	}
	return authedUser, nil
}

// UpdateUser updates an existing user's information
func (r *UserRepository) UpdateUser(user *models.User) error {
	query := `
		UPDATE users SET first_name = $2, last_name = $3, middle_name = $4, email = $5, phone = $6, is_producer = $7 WHERE id = $1
	`

	_, err := r.DB.Exec(query, user.ID, user.FirstName, user.LastName, user.MiddleName, user.Email, user.Phone, user.IsProducer)
	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}

	log.Printf("User %s %s updated successfully\n", user.FirstName, user.LastName)
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
	err := r.DB.QueryRow(query, value).Scan(&user.ID, &user.FirstName, &user.Email, &user.Phone, &user.LastName, &user.MiddleName, &user.IsProducer)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error retrieving user: %v", err)
	}

	return &user, nil
}
