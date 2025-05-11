package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"go_ecommerce/pkg/database"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/utils"
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
func (r *UserRepository) CreateUser(ctx context.Context, user *models.UserRegister) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("{repository/CreateUser - context error before starting transaction: %w}", ctx.Err())
		}
		return fmt.Errorf("{repository/CreateUser - error starting transaction: %w}", err)
	}

	defer utils.HandleTransaction(tx, &err)

	hashedPassword := utils.HashPassword(user.Password)

	userID := uuid.New().String()
	created_at := time.Now()
	authID := uuid.New().String()
	active := true

	userQuery := `
        INSERT INTO users (id, first_name, last_name, middle_name, email, phone, is_producer, address, city, country, zip_code, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	authQuery := `
		INSERT INTO auths (id, user_id, created_at, active, password)
		VALUES ($1, $2, $3, $4, $5)`

	_, err = tx.ExecContext(ctx, userQuery,
		userID, user.FirstName, user.LastName, user.MiddleName, user.Email, user.Phone, user.IsProducer, user.Address, user.City, user.Country, user.ZipCode, created_at)

	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("{repository/CreateUser - context error inserting user: %w}", ctx.Err())
		}
		return fmt.Errorf("{repository/CreateUser - error inserting user: %w}", err)
	}

	_, err = tx.ExecContext(ctx, authQuery, authID, userID, created_at, active, hashedPassword)

	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("{repository/CreateUser - context error inserting auth: %w}", ctx.Err())
		}
		return fmt.Errorf("{repository/CreateUser - error inserting auth: %w}", err)
	}

	return nil
}

// GetUserByID retrieves a user by their ID
func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT * FROM users WHERE id = $1
	`
	user, err := r.fetchUserByValue(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("{repository/GetUserByID - error retrieving user by ID: %w}", err)
	}

	return user, nil
}

// GetUserByName retrieves a user by their name
func (r *UserRepository) GetUserByFullName(ctx context.Context, firstName string, lastName string, middleName string) (*models.User, error) {
	query := `
		SELECT * FROM users WHERE first_name = $1 AND last_name = $2 AND middle_name = $3 
	`
	var user models.User
	err := r.DB.QueryRowContext(ctx, query, firstName, lastName, middleName).Scan(&user.ID, &user.FirstName, &user.Email, &user.Phone, &user.LastName, &user.MiddleName, &user.IsProducer, &user.Address, &user.City, &user.Country, &user.ZipCode, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("{repository/GetUserByFullName - user not found: %w}", err)
		}
		if ctx.Err() != nil {
			return nil, fmt.Errorf("{repository/GetUserByFullName - context error retrieving user by full name: %w}", ctx.Err())
		}
		return nil, fmt.Errorf("repository/GetUserByFullName - error retrieving user by full name: %w}", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by their email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT * FROM users WHERE email = $1
	`

	user, err := r.fetchUserByValue(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("repository/GetUserByEmail - error retrieving user by email: %w}", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email
func (r *UserRepository) GetAuthedUserByEmail(ctx context.Context, email string) (*models.AuthedUser, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("{repository/GetAuthedUserByEmail - context error before starting transaction for authed user: %w}", ctx.Err())
		}
		return nil, fmt.Errorf("{repository/GetAuthedUserByEmail - error starting transaction for authed user with email %s: %w}", email, err)
	}

	queryUsers := `
		SELECT * FROM users WHERE email = $1
	`
	queryAuths := `
		SELECT * FROM auths WHERE user_id = $1 AND active = $2
	`

	var user models.User
	err = r.DB.QueryRowContext(ctx, queryUsers, email).Scan(&user.ID, &user.FirstName, &user.Email, &user.Phone, &user.LastName, &user.MiddleName, &user.IsProducer, &user.Address, &user.City, &user.Country, &user.ZipCode, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("{repository/GetAuthedUserByEmail - user with id %s not found: %w}", user.ID, err)
		}
		if ctx.Err() != nil {
			return nil, fmt.Errorf("{repository/GetAuthedUserByEmail - context error retrieving authed user: %w}", ctx.Err())
		}
		return nil, fmt.Errorf("{repository/GetAuthedUserByEmail - error retrieving authed user with id %s: %w}", user.ID, err)
	}

	defer utils.HandleTransaction(tx, &err)

	var auth models.Auth
	err = r.DB.QueryRowContext(ctx, queryAuths, user.ID, true).Scan(&auth.ID, &auth.UserID, &auth.CreatedAt, &auth.Active, &auth.Password, &auth.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("{repository/GetAuthedUserByEmail - auth not found for user with id %s: %w}", user.ID, err)
		}
		if ctx.Err() != nil {
			return nil, fmt.Errorf("{repository/GetAuthedUserByEmail - context error retrieving auth for user: %w}", ctx.Err())
		}
		return nil, fmt.Errorf("{repository/GetAuthedUserByEmail - error retrieving auth for user with id %s: %w}", user.ID, err)
	}

	authedUser := &models.AuthedUser{
		ID:         user.ID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Email:      user.Email,
		Phone:      user.Phone,
		IsProducer: user.IsProducer,
		Address:    user.Address,
		City:       user.City,
		Country:    user.Country,
		ZipCode:    user.ZipCode,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
		Auth:       auth,
	}
	return authedUser, nil
}

// UpdateUser updates an existing user's information
func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	updated_at := time.Now()
	query := `
		UPDATE users SET first_name = $2, last_name = $3, middle_name = $4, email = $5, phone = $6, is_producer = $7, address = $8, city = $9, country = $10, zip_code = $11, created_at = $12, updated_at = $13 WHERE id = $1
	`

	_, err := r.DB.ExecContext(ctx, query, user.ID, user.FirstName, user.LastName, user.MiddleName, user.Email, user.Phone, user.IsProducer, user.Address, user.City, user.Country, user.ZipCode, user.CreatedAt, updated_at)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("{repository/UpdateUser - context error updating user: %w}", ctx.Err())
		}
		return fmt.Errorf("{repository/UpdateUser - error updating user: %w}", err)
	}

	return nil
}

// DeleteUser removes a user from the database
func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("{repository/DeleteUser - context error before starting transaction for delete: %w}", ctx.Err())
		}
		return fmt.Errorf("{repository/DeleteUser - error starting transaction for delete: %w}", err)
	}

	defer utils.HandleTransaction(tx, &err)

	queryAuth := `
		DELETE FROM auths WHERE user_id = $1
	`
	queryUser := `
		DELETE FROM users WHERE id = $1
	`

	_, err = r.DB.ExecContext(ctx, queryAuth, id)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("{repository/DeleteUser - context error deleting auth for user: %w}", ctx.Err())
		}
		return fmt.Errorf("{repository/DeleteUser - error deleting auth for user: %w}", err)
	}

	_, err = r.DB.ExecContext(ctx, queryUser, id)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("{repository/DeleteUser - context error deleting user: %w}", ctx.Err())
		}
		return fmt.Errorf("{repository/DeleteUser - error deleting user: %w}", err)
	}

	return nil
}

func (r *UserRepository) fetchUserByValue(ctx context.Context, query string, value string) (*models.User, error) {
	var user models.User
	err := r.DB.QueryRowContext(ctx, query, value).Scan(&user.ID, &user.FirstName, &user.Email, &user.Phone, &user.LastName, &user.MiddleName, &user.IsProducer, &user.Address, &user.City, &user.Country, &user.ZipCode, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("{repository/fetchUserByValue - user not found: %w}", err)
		}
		if ctx.Err() != nil {
			return nil, fmt.Errorf("{repository/fetchUserByValue - context error fetching user by value: %w}", ctx.Err())
		}
		return nil, fmt.Errorf("{repository/fetchUserByValue - error fetching user by value: %w}", err)
	}

	return &user, nil
}
