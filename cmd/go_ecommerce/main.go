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

	// company
	companyRepo := repositories.NewCompanyRepository(DB)
	companyService := services.NewCompanyService(companyRepo)
	companyHandler := handlers.NewCompanyHandler(companyService)

	// product
	productRepo := repositories.NewProductRepository(DB)

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
	reviewService := services.NewReviewService(reviewRepo, orderRepo, companyRepo, productRepo)
	reviewHandler := handlers.NewReviewHandler(reviewService)

	productService := services.NewProductService(productRepo, companyRepo, reviewRepo)
	productHandler := handlers.NewProductHandler(productService)

	// Create a new mux to apply CORS middleware
	mux := http.NewServeMux()

	// Define routes

	//user basic routes
	mux.Handle("/users/register", middleware.CSRFMiddleware(http.HandlerFunc(userHandler.Register)))
	mux.Handle("/users/login", middleware.CSRFMiddleware(http.HandlerFunc(userHandler.Login)))
	mux.Handle("/users/get-by-id", authMiddleware.AuthenticateJWT(http.HandlerFunc(userHandler.GetUserByID)))
	mux.Handle("/users/get-by-name", authMiddleware.AuthenticateJWT(http.HandlerFunc(userHandler.GetUserByName)))
	mux.Handle("/users/get-by-email", authMiddleware.AuthenticateJWT(http.HandlerFunc(userHandler.GetUserByEmail)))
	mux.Handle("/users/update", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(userHandler.UpdateUser))))
	mux.Handle("/users/delete", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(userHandler.DeleteUser))))

	//auth routes
	mux.Handle("/auth/change-password", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(authHandler.ChangePassword))))
	mux.Handle("/auth/logout", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(authHandler.Logout))))

	//product routes
	mux.Handle("/products/create", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(productHandler.CreateProduct))))
	mux.HandleFunc("/products", productHandler.GetProducts)
	mux.HandleFunc("/product", productHandler.GetProduct)
	mux.HandleFunc("/products/category", productHandler.GetProductsByCategory)
	mux.Handle("/products/my-products", authMiddleware.AuthenticateJWT(http.HandlerFunc(productHandler.GetProductsByUserID)))
	mux.Handle("/products/update", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(productHandler.UpdateProduct))))
	mux.Handle("/products/delete", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(productHandler.DeleteProduct))))

	// cart routes
	mux.Handle("/cart/add", authMiddleware.OptionalAuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(cartHandler.AddToCart))))
	mux.Handle("/cart/remove", authMiddleware.OptionalAuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(cartHandler.RemoveFromCart))))
	mux.Handle("/cart/clear", authMiddleware.OptionalAuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(cartHandler.ClearCart))))
	mux.Handle("/cart/get", authMiddleware.OptionalAuthenticateJWT(http.HandlerFunc(cartHandler.GetCartItems)))
	mux.Handle("/cart/update", authMiddleware.OptionalAuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(cartHandler.UpdateCartItems))))

	// order routes
	mux.Handle("/orders/checkout", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(orderHandler.ProcessCheckout))))
	mux.Handle("/orders/confirm", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(orderHandler.ConfirmOrder))))
	mux.Handle("/orders/summary", authMiddleware.AuthenticateJWT(http.HandlerFunc(orderHandler.GetOrderGroupSummary)))
	mux.Handle("/orders/details", authMiddleware.AuthenticateJWT(http.HandlerFunc(orderHandler.GetOrderDetails)))
	mux.Handle("/orders/group-details", authMiddleware.AuthenticateJWT(http.HandlerFunc(orderHandler.GetOrderGroupDetails)))
	mux.Handle("/orders/user", authMiddleware.AuthenticateJWT(http.HandlerFunc(orderHandler.GetUserOrders)))
	mux.Handle("/orders/producer", authMiddleware.AuthenticateJWT(http.HandlerFunc(orderHandler.GetProducerOrders)))
	mux.Handle("/orders/fulfill", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(orderHandler.FulfillOrder))))
	mux.Handle("/orders/sales-report", authMiddleware.AuthenticateJWT(http.HandlerFunc(orderHandler.GetSalesReport)))
	mux.Handle("/orders/cancel", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(orderHandler.CancelOrder))))
	mux.Handle("/orders/delete", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(orderHandler.CustomerDeleteOrder))))

	// review routes
	mux.Handle("/reviews/add", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(reviewHandler.AddReview))))
	mux.HandleFunc("/reviews/get", reviewHandler.GetReviewsByProductID)
	mux.Handle("/reviews/update", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(reviewHandler.UpdateReview))))
	mux.Handle("/reviews/delete", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(reviewHandler.DeleteReview))))

	// company routes
	mux.Handle("/companies/create", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(companyHandler.CreateCompany))))
	mux.Handle("/companies/get-by-user", authMiddleware.AuthenticateJWT(http.HandlerFunc(companyHandler.GetCompanyByUserID)))
	mux.HandleFunc("/companies/get-by-id", companyHandler.GetCompanyByCompanyID)
	mux.Handle("/companies/update", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(companyHandler.UpdateCompany))))
	mux.Handle("/companies/delete", authMiddleware.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(companyHandler.DeleteCompany))))

	// server
	log.Printf("Server is listening on port %v", cfg.AppPort)
	err = http.ListenAndServe(cfg.AppPort, middleware.CORS(mux))
	if err != nil {
		log.Fatal("Server failed: ", err)
	}
}
