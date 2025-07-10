package repositories

import (
	"context"
	"database/sql"
	"errors"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/utils"
	"time"

	"github.com/google/uuid"
)

type ICompanyRepository interface {
	CreateCompany(ctx context.Context, company *models.CompanyDTO, userId uuid.UUID) error
	GetCompanyByCompanyID(ctx context.Context, companyID uuid.UUID) (*models.Company, error)
	GetCompanyByUserID(ctx context.Context, userID uuid.UUID) (*models.Company, error)
	UpdateCompany(ctx context.Context, company *models.CompanyDTO) error
	DeleteCompany(ctx context.Context, companyID uuid.UUID) error
	UpdateCompanyReviewStats(ctx context.Context, userID uuid.UUID) error
}

type CompanyRepository struct {
	DB *sql.DB
}

func NewCompanyRepository(db *sql.DB) ICompanyRepository {
	return &CompanyRepository{DB: db}
}

func (r *CompanyRepository) CreateCompany(ctx context.Context, company *models.CompanyDTO, userId uuid.UUID) error {
	company.ID = uuid.New()
	now := time.Now()

	query := `
		INSERT INTO companies (id, user_id, name, address, city, country, zip_code, review_average, review_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.DB.ExecContext(ctx, query, company.ID, userId, company.Name, company.Address, company.City, company.Country, company.ZipCode, company.ReviewAverage, company.ReviewCount, now, nil)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/CreateCompany", company.Name)
	}

	return nil
}

func (r *CompanyRepository) GetCompanyByCompanyID(ctx context.Context, companyID uuid.UUID) (*models.Company, error) {
	query := `
		SELECT id, user_id, name, address, city, country, zip_code, review_average, review_count, created_at, updated_at
		FROM companies
		WHERE id = $1
	`

	row := r.DB.QueryRowContext(ctx, query, companyID)

	var company models.Company
	err := row.Scan(&company.ID, &company.UserID, &company.Name, &company.Address, &company.City, &company.Country, &company.ZipCode, &company.ReviewAverage, &company.ReviewCount, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetCompanyByCompanyID", companyID.String())
	}

	return &company, nil
}

func (r *CompanyRepository) GetCompanyByUserID(ctx context.Context, userID uuid.UUID) (*models.Company, error) {
	query := `
		SELECT id, user_id, name, address, city, country, zip_code, review_average, review_count, created_at, updated_at
		FROM companies
		WHERE user_id = $1
	`

	row := r.DB.QueryRowContext(ctx, query, userID)

	var company models.Company
	err := row.Scan(&company.ID, &company.UserID, &company.Name, &company.Address, &company.City, &company.Country, &company.ZipCode, &company.ReviewAverage, &company.ReviewCount, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetCompanyByUserID", userID.String())
	}

	return &company, nil
}

func (r *CompanyRepository) UpdateCompany(ctx context.Context, company *models.CompanyDTO) error {
	query := `
		UPDATE companies
		SET name = $1, address = $2, city = $3, country = $4, zip_code = $5, review_average = $6, review_count = $7, updated_at = $8
		WHERE id = $9
	`

	rows, err := r.DB.ExecContext(ctx, query, company.Name, company.Address, company.City, company.Country, company.ZipCode, company.ReviewAverage, company.ReviewCount, time.Now(), company.ID)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdateCompany", company.ID.String())
	}

	rowsAffected, err := rows.RowsAffected()
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdateCompany", company.ID.String())
	}
	if rowsAffected == 0 {
		return utils.HandleRepositoryErrors(ctx, errors.New("company not found"), "repository/UpdateCompany", company.ID.String())
	}

	return nil
}

func (r *CompanyRepository) DeleteCompany(ctx context.Context, companyID uuid.UUID) error {
	query := `
		DELETE FROM companies
		WHERE id = $1
	`

	rows, err := r.DB.ExecContext(ctx, query, companyID)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/DeleteCompany", companyID.String())
	}

	rowsAffected, err := rows.RowsAffected()
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/DeleteCompany", companyID.String())
	}
	if rowsAffected == 0 {
		return utils.HandleRepositoryErrors(ctx, errors.New("company not found"), "repository/DeleteCompany", companyID.String())
	}

	return nil
}

// UpdateCompanyReviewStats recalculates and updates the review average and count for a company
// based on all reviews for products owned by the company's user
func (r *CompanyRepository) UpdateCompanyReviewStats(ctx context.Context, userID uuid.UUID) error {
	// Query to calculate average rating and count for all products owned by this user
	query := `
		SELECT 
			COALESCE(AVG(r.rating), 0) as average_rating,
			COALESCE(COUNT(r.id), 0) as review_count
		FROM reviews r
		JOIN products p ON r.product_id = p.id
		WHERE p.user_id = $1
	`

	var averageRating float64
	var reviewCount int

	err := r.DB.QueryRowContext(ctx, query, userID).Scan(&averageRating, &reviewCount)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdateCompanyReviewStats", userID.String())
	}

	// Update the company's review statistics
	updateQuery := `
		UPDATE companies 
		SET review_average = $1, review_count = $2, updated_at = $3
		WHERE user_id = $4
	`

	now := time.Now()
	result, err := r.DB.ExecContext(ctx, updateQuery, averageRating, reviewCount, now, userID)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdateCompanyReviewStats", userID.String())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdateCompanyReviewStats", userID.String())
	}

	// If no rows were affected, it means the user doesn't have a company record
	// This is acceptable - not all producers have company records
	if rowsAffected == 0 {
		return nil
	}

	return nil
}
