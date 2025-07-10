package services

import (
	"context"
	"errors"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/repositories"
	"go_ecommerce/pkg/utils"

	"github.com/google/uuid"
)

type IReviewService interface {
	AddReview(ctx context.Context, userID uuid.UUID, dto *models.ReviewDTO) error
	GetReviewsByProductID(ctx context.Context, productID uuid.UUID) ([]models.Review, error)
	GetReviewByID(ctx context.Context, reviewID uuid.UUID) (*models.Review, error)
	UpdateReview(ctx context.Context, userID, reviewID uuid.UUID, rating int, comment *string) error
	DeleteReview(ctx context.Context, userID, reviewID uuid.UUID) error
}

type ReviewService struct {
	ReviewRepo  repositories.IReviewRepository
	OrderRepo   repositories.IOrderRepository
	CompanyRepo repositories.ICompanyRepository
	ProductRepo repositories.IProductRepository
}

func NewReviewService(reviewRepo repositories.IReviewRepository, orderRepo repositories.IOrderRepository, companyRepo repositories.ICompanyRepository, productRepo repositories.IProductRepository) IReviewService {
	return &ReviewService{
		ReviewRepo:  reviewRepo,
		OrderRepo:   orderRepo,
		CompanyRepo: companyRepo,
		ProductRepo: productRepo,
	}
}

// AddReview allows a user to add a review for a product they have purchased
func (s *ReviewService) AddReview(ctx context.Context, userID uuid.UUID, dto *models.ReviewDTO) error {
	// Check if user has already reviewed this product
	existing, err := s.ReviewRepo.GetReviewByUserAndProductID(ctx, userID, dto.ProductID)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/AddReview")
	}
	if existing != nil {
		return utils.HandleServiceErrors(ctx, errors.New("user has already reviewed this product"), "service/AddReview")
	}

	// Check if user has purchased this product (order with this product and userID exists)
	orders, err := s.OrderRepo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/AddReview")
	}
	purchased := false
	for _, order := range orders {
		orderDetails, err := s.OrderRepo.GetOrderWithDetails(ctx, order.ID)
		if err != nil {
			continue
		}
		for _, item := range orderDetails.Items {
			if item.ProductID == dto.ProductID {
				purchased = true
				break
			}
		}
		if purchased {
			break
		}
	}
	if !purchased {
		return utils.HandleServiceErrors(ctx, errors.New("user has not purchased this product"), "service/AddReview")
	}

	review := &models.Review{
		ProductID: dto.ProductID,
		UserID:    userID,
		Rating:    dto.Rating,
		Comment:   dto.Comment,
	}

	err = s.ReviewRepo.CreateReview(ctx, review)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/AddReview")
	}

	// Get the product to find the producer (user who owns the product)
	product, err := s.ProductRepo.GetProductByID(ctx, dto.ProductID.String())
	if err != nil {
		return nil
	}

	// Update the company's review statistics for the product owner
	err = s.CompanyRepo.UpdateCompanyReviewStats(ctx, product.UserID)
	if err != nil {
		return nil
	}

	return nil
}

func (s *ReviewService) GetReviewsByProductID(ctx context.Context, productID uuid.UUID) ([]models.Review, error) {
	return s.ReviewRepo.GetReviewsByProductID(ctx, productID)
}

func (s *ReviewService) GetReviewByID(ctx context.Context, reviewID uuid.UUID) (*models.Review, error) {
	return s.ReviewRepo.GetReviewByID(ctx, reviewID)
}

func (s *ReviewService) UpdateReview(ctx context.Context, userID, reviewID uuid.UUID, rating int, comment *string) error {
	if rating < 1 || rating > 5 {
		return utils.HandleServiceErrors(ctx, errors.New("rating must be between 1 and 5"), "service/UpdateReview")
	}

	// Get the review to find the product ID
	review, err := s.ReviewRepo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/UpdateReview")
	}

	if review == nil || review.UserID != userID {
		return utils.HandleServiceErrors(ctx, errors.New("review not found or not owned by user"), "service/UpdateReview")
	}

	err = s.ReviewRepo.UpdateReview(ctx, reviewID, userID, rating, comment)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/UpdateReview")
	}

	// Get the product to find the producer (user who owns the product)
	product, err := s.ProductRepo.GetProductByID(ctx, review.ProductID.String())
	if err != nil {
		return nil
	}

	// Update the company's review statistics for the product owner
	err = s.CompanyRepo.UpdateCompanyReviewStats(ctx, product.UserID)
	if err != nil {
		return nil
	}

	return nil
}

func (s *ReviewService) DeleteReview(ctx context.Context, userID, reviewID uuid.UUID) error {
	// Get the review to find the product ID before deleting
	review, err := s.ReviewRepo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/DeleteReview")
	}

	if review == nil || review.UserID != userID {
		return utils.HandleServiceErrors(ctx, errors.New("review not found or not owned by user"), "service/DeleteReview")
	}

	err = s.ReviewRepo.DeleteReview(ctx, reviewID, userID)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/DeleteReview")
	}

	// Get the product to find the producer (user who owns the product)
	product, err := s.ProductRepo.GetProductByID(ctx, review.ProductID.String())
	if err != nil {
		return nil
	}

	// Update the company's review statistics for the product owner
	err = s.CompanyRepo.UpdateCompanyReviewStats(ctx, product.UserID)
	if err != nil {
		return nil
	}

	return nil
}
