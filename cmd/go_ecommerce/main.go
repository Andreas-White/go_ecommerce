package main

import (
	"go_ecommerce/internal/config"
	"go_ecommerce/pkg/database"
	"go_ecommerce/pkg/handlers"
	"go_ecommerce/pkg/middleware"
	"go_ecommerce/pkg/repositories"
	"go_ecommerce/pkg/services"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()

	DB, err := database.Init(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}
	defer DB.Close()

	err = database.RunMigrations(cfg)
	if err != nil {
		log.Fatal("Failed to run migrations: ", err)
	}

	authMiddleware, err := middleware.NewAuthenticator(cfg.JWTKey)
	if err != nil {
		log.Fatal("Failed to initialize authenticator: ", err)
	}

	// auth
	authRepo := repositories.NewAuthRepository(DB)
	authService := services.NewAuthService(authRepo)
	authHandler := handlers.NewAuthHandler(authService)

	// user
	userRepo := repositories.NewUserRepository(DB)
	userService := services.NewUserService(userRepo, authMiddleware)
	userHandler := handlers.NewUserHandler(userService)

	// product
	productRepo := repositories.NewProductRepository(DB)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	// cart
	cartRepo := repositories.NewCartRepository(DB)
	cartService := services.NewCartService(cartRepo, productRepo)
	cartHandler := handlers.NewCartHandler(cartService)

	// order
	orderRepo := repositories.NewOrderRepository(DB)
	orderService := services.NewOrderService(orderRepo, cartRepo, productRepo)
	orderHandler := handlers.NewOrderHandler(orderService)

	// review
	reviewRepo := repositories.NewReviewRepository(DB)
	reviewService := services.NewReviewService(reviewRepo, orderRepo)
	reviewHandler := handlers.NewReviewHandler(reviewService)

	// Define routes
	//user basic routes
	http.HandleFunc("/users/register", userHandler.Register)
	http.HandleFunc("/users/login", userHandler.Login)
	http.Handle("/users/get-by-id", authMiddleware.AuthenticateJWT(http.HandlerFunc(userHandler.GetUserByID)))
	http.Handle("/users/get-by-name", authMiddleware.AuthenticateJWT(http.HandlerFunc(userHandler.GetUserByName)))
	http.Handle("/users/get-by-email", authMiddleware.AuthenticateJWT(http.HandlerFunc(userHandler.GetUserByEmail)))
	http.Handle("/users/update", authMiddleware.AuthenticateJWT(http.HandlerFunc(userHandler.UpdateUser)))
	http.Handle("/users/delete", authMiddleware.AuthenticateJWT(http.HandlerFunc(userHandler.DeleteUser)))

	//auth routes
	http.Handle("/auth/change-password", authMiddleware.AuthenticateJWT(http.HandlerFunc(authHandler.ChangePassword)))

	//product routes
	http.Handle("/products/create", authMiddleware.AuthenticateJWT(http.HandlerFunc(productHandler.CreateProduct)))
	http.HandleFunc("/products", productHandler.GetProducts)
	http.HandleFunc("/product", productHandler.GetProduct)
	http.HandleFunc("/products/{id}", productHandler.GetProduct)
	http.HandleFunc("/products/category", productHandler.GetProductsByCategory)
	http.HandleFunc("/products/user-id", productHandler.GetProductsByUserID)
	http.Handle("/products/update", authMiddleware.AuthenticateJWT(http.HandlerFunc(productHandler.UpdateProduct)))
	http.Handle("/products/delete", authMiddleware.AuthenticateJWT(http.HandlerFunc(productHandler.DeleteProduct)))

	// cart routes
	http.Handle("/cart/add", authMiddleware.OptionalAuthenticateJWT(http.HandlerFunc(cartHandler.AddToCart)))
	http.Handle("/cart/remove", authMiddleware.OptionalAuthenticateJWT(http.HandlerFunc(cartHandler.RemoveFromCart)))
	http.Handle("/cart/clear", authMiddleware.OptionalAuthenticateJWT(http.HandlerFunc(cartHandler.ClearCart)))
	http.Handle("/cart/get", authMiddleware.OptionalAuthenticateJWT(http.HandlerFunc(cartHandler.GetCartItems)))
	http.Handle("/cart/update", authMiddleware.OptionalAuthenticateJWT(http.HandlerFunc(cartHandler.UpdateCartItems)))

	// order routes
	http.Handle("/orders/checkout", authMiddleware.AuthenticateJWT(http.HandlerFunc(orderHandler.ProcessCheckout)))
	http.Handle("/orders/confirm", authMiddleware.AuthenticateJWT(http.HandlerFunc(orderHandler.ConfirmOrder)))
	http.Handle("/orders/summary", authMiddleware.AuthenticateJWT(http.HandlerFunc(orderHandler.GetOrderSummary)))
	http.Handle("/orders/details", authMiddleware.AuthenticateJWT(http.HandlerFunc(orderHandler.GetOrderDetails)))
	http.Handle("/orders/user", authMiddleware.AuthenticateJWT(http.HandlerFunc(orderHandler.GetUserOrders)))

	// review routes
	http.Handle("/reviews/add", authMiddleware.AuthenticateJWT(http.HandlerFunc(reviewHandler.AddReview)))
	http.HandleFunc("/reviews/get", reviewHandler.GetReviewsByProductID)
	http.Handle("/reviews/update", authMiddleware.AuthenticateJWT(http.HandlerFunc(reviewHandler.UpdateReview)))
	http.Handle("/reviews/delete", authMiddleware.AuthenticateJWT(http.HandlerFunc(reviewHandler.DeleteReview)))

	// server
	log.Printf("Server is listening on port %v", cfg.AppPort)
	err = http.ListenAndServe(cfg.AppPort, nil)
	if err != nil {
		log.Fatal("Server failed: ", err)
	}
}
