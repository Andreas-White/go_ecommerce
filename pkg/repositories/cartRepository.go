package repositories

import (
	"context"
	"database/sql"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/utils"
	"time"

	"github.com/google/uuid"
)

type ICartRepository interface {
	CreateGuestCart(ctx context.Context, sessionID string) (*models.Cart, error)
	GetCartByUserID(ctx context.Context, userID uuid.UUID) (*models.Cart, error)
	GetCartBySessionID(ctx context.Context, sessionID string) (*models.Cart, error)
	ClearCart(ctx context.Context, cartId *uuid.UUID) error
	AddProductsToCart(ctx context.Context, cartItems []models.CartItemDTO, cart *models.Cart) (*models.Cart, error)
	RemoveProductsFromCart(ctx context.Context, cartItems []models.CartItemDTO) error
	GetAllCartItemsByCartID(ctx context.Context, cartID uuid.UUID) ([]models.CartItemProductDetails, error)
	UpdateCartItems(ctx context.Context, cartItems []models.CartItemDTO) error
}

type CartRepository struct {
	DB *sql.DB
}

func NewCartRepository(db *sql.DB) ICartRepository {
	return &CartRepository{DB: db}
}

func (r *CartRepository) CreateGuestCart(ctx context.Context, sessionID string) (*models.Cart, error) {
	cartID := uuid.New()
	createdAt := time.Now()
	cart := &models.Cart{
		ID:        cartID,
		SessionID: &sessionID,
		CreatedAt: createdAt,
		UpdatedAt: &createdAt,
	}
	query := "INSERT INTO carts (id, user_id, created_at, updated_at, session_id) VALUES ($1, $2, $3, $4, $5)"
	_, err := r.DB.ExecContext(ctx, query, cart.ID, nil, cart.CreatedAt, cart.UpdatedAt, cart.SessionID)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/CreateGuestCart", sessionID)
	}
	return cart, nil
}

func (r *CartRepository) GetCartByUserID(ctx context.Context, userID uuid.UUID) (*models.Cart, error) {
	query := "SELECT id, user_id, created_at, updated_at, session_id FROM carts WHERE user_id = $1"
	cart := &models.Cart{}
	err := r.DB.QueryRowContext(ctx, query, userID).Scan(&cart.ID, &cart.UserID, &cart.CreatedAt, &cart.UpdatedAt, &cart.SessionID)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetCartByUserID", userID.String())
	}
	return cart, nil
}

func (r *CartRepository) GetCartBySessionID(ctx context.Context, sessionID string) (*models.Cart, error) {
	query := "SELECT id, user_id, created_at, updated_at, session_id FROM carts WHERE session_id = $1"
	cart := &models.Cart{}
	err := r.DB.QueryRowContext(ctx, query, sessionID).Scan(&cart.ID, &cart.UserID, &cart.CreatedAt, &cart.UpdatedAt, &cart.SessionID)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetCartBySessionID", sessionID)
	}
	return cart, nil
}

func (r *CartRepository) ClearCart(ctx context.Context, cartId *uuid.UUID) error {
	query := "DELETE FROM cart_items WHERE cart_id = $1"
	_, err := r.DB.ExecContext(ctx, query, cartId)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/ClearCart", cartId.String())
	}
	return nil
}

func (r *CartRepository) AddProductsToCart(ctx context.Context, cartItems []models.CartItemDTO, cart *models.Cart) (*models.Cart, error) {
	for _, item := range cartItems {
		cartItemId := uuid.New()
		createdAt := time.Now()

		cartItem := models.CartItem{
			ID:        cartItemId,
			CartID:    cart.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			CreatedAt: createdAt,
		}

		query := "INSERT INTO cart_items (id, cart_id, product_id, quantity, price, created_at) VALUES ($1, $2, $3, $4, $5, $6)"
		_, err := r.DB.ExecContext(ctx, query, cartItem.ID, cartItem.CartID, cartItem.ProductID, cartItem.Quantity, cartItem.Price, cartItem.CreatedAt)
		if err != nil {
			return nil, utils.HandleRepositoryErrors(ctx, err, "repository/AddProductToCart", item.ProductID.String())
		}
	}
	return cart, nil
}

func (r *CartRepository) RemoveProductsFromCart(ctx context.Context, cartItems []models.CartItemDTO) error {
	for _, item := range cartItems {
		query := "DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2"
		_, err := r.DB.ExecContext(ctx, query, item.CartID, item.ProductID)
		if err != nil {
			return utils.HandleRepositoryErrors(ctx, err, "repository/RemoveProductFromCart", item.ProductID.String())
		}
	}
	return nil
}

func (r *CartRepository) GetAllCartItemsByCartID(ctx context.Context, cartID uuid.UUID) ([]models.CartItemProductDetails, error) {
	query := `SELECT 
        ci.id, ci.cart_id, ci.product_id, ci.quantity, ci.price,
        p.name as product_name, p.description as product_description, p.stock as product_stock,
        p.category as product_category, p.image_url as product_image_url, p.user_id as product_user_id
        FROM cart_items ci
        JOIN products p ON ci.product_id = p.id
        WHERE ci.cart_id = $1`

	rows, err := r.DB.QueryContext(ctx, query, cartID)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetAllCartItemsByCartID", cartID.String())
	}
	defer rows.Close()

	cartItems := []models.CartItemProductDetails{}
	for rows.Next() {
		var item models.CartItemProductDetails
		err := rows.Scan(
			&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.Price,
			&item.ProductName, &item.ProductDescription, &item.ProductStock,
			&item.ProductCategory, &item.ProductImageUrl, &item.ProductUserID,
		)
		if err != nil {
			return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetAllCartItemsByCartID", cartID.String())
		}
		cartItems = append(cartItems, item)
	}
	return cartItems, nil
}

func (r *CartRepository) UpdateCartItems(ctx context.Context, cartItems []models.CartItemDTO) error {
	for _, item := range cartItems {
		var result sql.Result
		var err error
		var query string
		if item.Price == 0 {
			query = "UPDATE cart_items SET quantity = $1 WHERE cart_id = $2 AND product_id = $3"
			result, err = r.DB.ExecContext(ctx, query, item.Quantity, item.CartID, item.ProductID)
		} else {
			query = "UPDATE cart_items SET quantity = $1, price = $2 WHERE cart_id = $3 AND product_id = $4"
			result, err = r.DB.ExecContext(ctx, query, item.Quantity, item.Price, item.CartID, item.ProductID)
		}

		if err != nil {
			return utils.HandleRepositoryErrors(ctx, err, "repository/UpdateCartItems", item.ProductID.String())
		}

		affectedRows, err := result.RowsAffected()
		if err != nil {
			return utils.HandleRepositoryErrors(ctx, err, "repository/UpdateCartItems", item.ProductID.String())
		}
		if affectedRows == 0 {
			return utils.HandleRepositoryErrors(ctx, sql.ErrNoRows, "repository/UpdateCartItems", item.ProductID.String())
		}
	}
	return nil
}
