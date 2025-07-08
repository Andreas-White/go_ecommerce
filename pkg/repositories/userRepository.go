package repositories

import (
	"context"
	"database/sql"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/utils"
	"time"

	"github.com/google/uuid"
)

type IUserRepository interface {
	CreateUser(ctx context.Context, user *models.UserDTO) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByFullName(ctx context.Context, firstName string, lastName string, middleName string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetAuthedUserByEmail(ctx context.Context, email string) (*models.AuthedUser, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
}

// UserRepository struct
type UserRepository struct {
	DB *sql.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

// CreateUser inserts a new user into the database
func (r *UserRepository) CreateUser(ctx context.Context, user *models.UserDTO) error {
	tx, err := r.startTransaction(ctx)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/CreateUser", user.Email)
	}

	defer utils.HandleTransaction(tx, &err)

	hashedPassword := utils.HashPassword(user.Password)

	userID := uuid.New().String()
	authID := uuid.New().String()
	cartID := uuid.New().String()
	created_at := time.Now()
	active := true

	userQuery := `
        INSERT INTO users (id, first_name, last_name, middle_name, email, phone, is_producer, address,
		city, country, zip_code, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	authQuery := `
		INSERT INTO auths (id, user_id, created_at, active, password)
		VALUES ($1, $2, $3, $4, $5)`

	cartQuery := `
		INSERT INTO carts (id, user_id, session_id, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err = tx.ExecContext(ctx, userQuery,
		userID, user.FirstName, user.LastName, user.MiddleName, user.Email, user.Phone, user.IsProducer,
		user.Address, user.City, user.Country, user.ZipCode, created_at)

	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/CreateUser", user.Email)
	}

	_, err = tx.ExecContext(ctx, authQuery, authID, userID, created_at, active, hashedPassword)

	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/CreateUser", user.Email)
	}

	_, err = tx.ExecContext(ctx, cartQuery, cartID, userID, nil, created_at)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/CreateUser", user.Email)
	}

	return nil
}

// GetUserByID retrieves a user by their ID
func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, first_name, last_name, middle_name, email, phone, is_producer, address,
		city, country, zip_code, created_at, updated_at FROM users WHERE id = $1
	`
	user, err := r.fetchUserByValue(ctx, query, id)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetUserByID", id)
	}

	return user, nil
}

// GetUserByName retrieves a user by their name
func (r *UserRepository) GetUserByFullName(ctx context.Context, firstName string, lastName string, middleName string) (*models.User, error) {
	query := `
		SELECT id, first_name, last_name, middle_name, email, phone, is_producer, address,
		city, country, zip_code, created_at, updated_at FROM users WHERE first_name = $1 AND last_name = $2 AND middle_name = $3 
	`
	row := r.DB.QueryRowContext(ctx, query, firstName, lastName, middleName)
	user, err := scanUser(ctx, row)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetUserByFullName", firstName)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, first_name, last_name, middle_name, email, phone, is_producer, address,
		city, country, zip_code, created_at, updated_at FROM users WHERE email = $1
	`

	user, err := r.fetchUserByValue(ctx, query, email)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetUserByEmail", email)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email
func (r *UserRepository) GetAuthedUserByEmail(ctx context.Context, email string) (*models.AuthedUser, error) {
	tx, err := r.startTransaction(ctx)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetAuthedUserByEmail", email)
	}

	queryUsers := `
		SELECT id, first_name, last_name, middle_name, email, phone, is_producer, address,
		city, country, zip_code, created_at, updated_at FROM users WHERE email = $1
	`
	queryAuths := `
		SELECT id, user_id, created_at, active, password, updated_at FROM auths WHERE user_id = $1 AND active = $2
	`

	user, err := r.fetchUserByValue(ctx, queryUsers, email)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetAuthedUserByEmail", email)
	}

	defer utils.HandleTransaction(tx, &err)

	var auth models.Auth
	err = r.DB.QueryRowContext(ctx, queryAuths, user.ID, true).Scan(&auth.ID, &auth.UserID,
		&auth.CreatedAt, &auth.Active, &auth.Password, &auth.UpdatedAt)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetAuthedUserByEmail", email)
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
		UPDATE users SET first_name = $2, last_name = $3, middle_name = $4, email = $5, phone = $6, 
		is_producer = $7, address = $8, city = $9, country = $10, zip_code = $11, created_at = $12, 
		updated_at = $13 WHERE id = $1
	`

	_, err := r.DB.ExecContext(ctx, query, user.ID, user.FirstName, user.LastName, user.MiddleName,
		user.Email, user.Phone, user.IsProducer, user.Address, user.City, user.Country, user.ZipCode,
		user.CreatedAt, updated_at)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdateUser", user.ID.String())
	}

	return nil
}

// DeleteUser removes a user from the database
func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	queryUser := `
		DELETE FROM users WHERE id = $1
	`

	result, err := r.DB.ExecContext(ctx, queryUser, id)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/DeleteUser", id)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/DeleteUser", id)
	}
	if rowsAffected == 0 {
		return utils.HandleRepositoryErrors(ctx, sql.ErrNoRows,
			"repository/DeleteUser", id)
	}

	return nil
}

func (r *UserRepository) fetchUserByValue(ctx context.Context, query string, value string) (*models.User, error) {
	user, err := scanUser(ctx, r.DB.QueryRowContext(ctx, query, value))
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/fetchUserByValue", value)
	}
	return user, nil
}

func scanUser(ctx context.Context, row *sql.Row) (*models.User, error) {
	var user models.User
	err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.MiddleName, &user.Email, &user.Phone, &user.IsProducer,
		&user.Address, &user.City, &user.Country, &user.ZipCode, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		err = utils.HandleRepositoryErrors(ctx, err, "repository/ScanUser", "")
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) startTransaction(ctx context.Context) (*sql.Tx, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/startTransaction", "")
	}

	return tx, nil
}
