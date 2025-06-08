package repositories

import (
	"context"
	"database/sql"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/utils"
	"time"
)

type IAuthRepository interface {
	GetAuthByUserID(ctx context.Context, userID string) (*models.Auth, error)
	UpdatePassword(ctx context.Context, userID string, newHashedPassword string) error
}

type AuthRepository struct {
	DB *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{
		DB: db,
	}
}

// GetAuthByUserID retrieves authentication details for a user by their ID.
func (r *AuthRepository) GetAuthByUserID(ctx context.Context, userID string) (*models.Auth, error) {
	query := "SELECT id, user_id, created_at, active, password, updated_at FROM auths WHERE user_id = $1 AND active = TRUE"
	auth := &models.Auth{}
	err := r.DB.QueryRowContext(ctx, query, userID).Scan(
		&auth.ID,
		&auth.UserID,
		&auth.CreatedAt,
		&auth.Active,
		&auth.Password,
		&auth.UpdatedAt,
	)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetAuthByUserID", userID)
	}
	return auth, nil
}

// UpdatePassword updates the password for a user's authentication record.
func (r *AuthRepository) UpdatePassword(ctx context.Context, userID string, newHashedPassword string) error {
	updatedAt := time.Now()
	query := "UPDATE auths SET password = $1, updated_at = $2 WHERE user_id = $3 AND active = TRUE"
	result, err := r.DB.ExecContext(ctx, query, newHashedPassword, updatedAt, userID)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdatePassword", userID)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdatePassword", userID)
	}

	if rowsAffected == 0 {
		return utils.HandleRepositoryErrors(ctx, sql.ErrNoRows, "repository/UpdatePassword", userID)
	}
	return nil
}
