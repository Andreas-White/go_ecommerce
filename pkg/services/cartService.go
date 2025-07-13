package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/repositories"
	"go_ecommerce/pkg/utils"
)

type ICartService interface {
	AddProductsToCart(ctx context.Context, userID *uuid.UUID, sessionID *string, cartItems []models.CartItemDTO) (*models.Cart, error)
	RemoveProductsFromCart(ctx context.Context, cartItem []models.CartItemDTO) error
	ClearCart(ctx context.Context, cartId *uuid.UUID) error
	GetAllCartItemsByCartID(ctx context.Context, cartID uuid.UUID) ([]models.CartItemProductDetails, error)
	UpdateCartItems(ctx context.Context, cartItems []models.CartItemDTO) error
	GetCartByUserID(ctx context.Context, userID uuid.UUID) (*models.Cart, error)
	GetCartBySessionID(ctx context.Context, sessionID string) (*models.Cart, error)
}

type CartService struct {
	cartRepo    repositories.ICartRepository
	productRepo repositories.IProductRepository
}

func NewCartService(cartRepo repositories.ICartRepository, productRepo repositories.IProductRepository) ICartService {
	return &CartService{cartRepo: cartRepo, productRepo: productRepo}
}

func (s *CartService) AddProductsToCart(ctx context.Context, userID *uuid.UUID, sessionID *string, cartItems []models.CartItemDTO) (*models.Cart, error) {
	var cart *models.Cart
	var err error

	if userID != nil {
		cart, err = s.cartRepo.GetCartByUserID(ctx, *userID)
		if err != nil {
			return nil, utils.HandleServiceErrors(ctx, err, "service/AddProductsToCart")
		}
	} else if sessionID != nil {
		cart, err = s.cartRepo.GetCartBySessionID(ctx, *sessionID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				cart, err = s.cartRepo.CreateGuestCart(ctx, *sessionID)
				if err != nil {
					return nil, utils.HandleServiceErrors(ctx, err, "service/AddProductsToCart")
				}
			} else {
				return nil, utils.HandleServiceErrors(ctx, err, "service/AddProductsToCart")
			}
		}
	} else {
		return nil, utils.HandleServiceErrors(ctx, errors.New("cart cannot be created without a user or session"), "service/AddProductsToCart")
	}

	return s.cartRepo.AddProductsToCart(ctx, cartItems, cart)
}

func (s *CartService) RemoveProductsFromCart(ctx context.Context, cartItem []models.CartItemDTO) error {
	return s.cartRepo.RemoveProductsFromCart(ctx, cartItem)
}

func (s *CartService) ClearCart(ctx context.Context, cartId *uuid.UUID) error {
	return s.cartRepo.ClearCart(ctx, cartId)
}

func (s *CartService) GetAllCartItemsByCartID(ctx context.Context, cartID uuid.UUID) ([]models.CartItemProductDetails, error) {
	return s.cartRepo.GetAllCartItemsByCartID(ctx, cartID)
}

func (s *CartService) UpdateCartItems(ctx context.Context, cartItems []models.CartItemDTO) error {
	return s.cartRepo.UpdateCartItems(ctx, cartItems)
}

func (s *CartService) GetCartByUserID(ctx context.Context, userID uuid.UUID) (*models.Cart, error) {
	return s.cartRepo.GetCartByUserID(ctx, userID)
}

func (s *CartService) GetCartBySessionID(ctx context.Context, sessionID string) (*models.Cart, error) {
	return s.cartRepo.GetCartBySessionID(ctx, sessionID)
}
