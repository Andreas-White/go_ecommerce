package repositories

import (
	"context"
	"database/sql"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/utils"
	"time"

	"github.com/google/uuid"
)

type IReviewRepository interface {
	CreateReview(ctx context.Context, review *models.Review) error
	GetReviewsByProductID(ctx context.Context, productID uuid.UUID) ([]models.Review, error)
	GetReviewByUserAndProductID(ctx context.Context, userID, productID uuid.UUID) (*models.Review, error)
	UpdateReview(ctx context.Context, reviewID, userID uuid.UUID, rating int, comment *string) error
	DeleteReview(ctx context.Context, reviewID, userID uuid.UUID) error
}

type ReviewRepository struct {
	DB *sql.DB
}

func NewReviewRepository(db *sql.DB) IReviewRepository {
	return &ReviewRepository{DB: db}
}

func (r *ReviewRepository) CreateReview(ctx context.Context, review *models.Review) error {
	now := time.Now()
	review.ID = uuid.New()
	review.CreatedAt = now
	query := `INSERT INTO reviews (id, product_id, user_id, rating, comment, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.DB.ExecContext(ctx, query, review.ID, review.ProductID, review.UserID, review.Rating, review.Comment, review.CreatedAt)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/CreateReview", review.ProductID.String())
	}
	return nil
}

func (r *ReviewRepository) GetReviewsByProductID(ctx context.Context, productID uuid.UUID) ([]models.Review, error) {
	query := `SELECT id, product_id, user_id, rating, comment, created_at FROM reviews WHERE product_id = $1 ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetReviewsByProductID", productID.String())
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var review models.Review
		err := rows.Scan(&review.ID, &review.ProductID, &review.UserID, &review.Rating, &review.Comment, &review.CreatedAt)
		if err != nil {
			return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetReviewsByProductID", productID.String())
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func (r *ReviewRepository) GetReviewByUserAndProductID(ctx context.Context, userID, productID uuid.UUID) (*models.Review, error) {
	query := `SELECT id, product_id, user_id, rating, comment, created_at FROM reviews WHERE user_id = $1 AND product_id = $2`
	row := r.DB.QueryRowContext(ctx, query, userID, productID)
	var review models.Review
	err := row.Scan(&review.ID, &review.ProductID, &review.UserID, &review.Rating, &review.Comment, &review.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetReviewByUserAndProductID", userID.String())
	}
	return &review, nil
}

func (r *ReviewRepository) UpdateReview(ctx context.Context, reviewID, userID uuid.UUID, rating int, comment *string) error {
	query := `UPDATE reviews SET rating = $1, comment = $2 WHERE id = $3 AND user_id = $4`
	result, err := r.DB.ExecContext(ctx, query, rating, comment, reviewID, userID)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdateReview", reviewID.String())
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdateReview", reviewID.String())
	}
	if rows == 0 {
		return utils.HandleRepositoryErrors(ctx, sql.ErrNoRows, "repository/UpdateReview", reviewID.String())
	}
	return nil
}

func (r *ReviewRepository) DeleteReview(ctx context.Context, reviewID, userID uuid.UUID) error {
	query := `DELETE FROM reviews WHERE id = $1 AND user_id = $2`
	result, err := r.DB.ExecContext(ctx, query, reviewID, userID)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/DeleteReview", reviewID.String())
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/DeleteReview", reviewID.String())
	}
	if rows == 0 {
		return utils.HandleRepositoryErrors(ctx, sql.ErrNoRows, "repository/DeleteReview", reviewID.String())
	}
	return nil
}
