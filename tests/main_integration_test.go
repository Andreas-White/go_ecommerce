package tests

import (
	"fmt"
	"go_ecommerce/internal/config"
	"go_ecommerce/pkg/database"
	"go_ecommerce/pkg/handlers"
	"go_ecommerce/pkg/middleware"
	"go_ecommerce/pkg/repositories"
	"go_ecommerce/pkg/services"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	tCfg := config.LoadTestConfig()
	cfg := &config.Config{
		DBHost:    tCfg.DBHost,
		DBPort:    tCfg.DBPort,
		DBUser:    tCfg.DBUser,
		DBPass:    tCfg.DBPass,
		DBName:    tCfg.DBName,
		DBSslMode: tCfg.DBSslMode,
		JWTKey:    tCfg.JWTKey,
	}

	// Run migrations for the test database
	migrationDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSslMode)

	// Migrations are in the parent directory's 'migrations' folder
	migrationsPath := "file://../migrations"

	log.Printf("Attempting to run migrations for test DB from: %s on database: %s", migrationsPath, cfg.DBName)
	mig, err := migrate.New(migrationsPath, migrationDSN)
	if err != nil {
		log.Fatalf("Failed to create new migrate instance for test DB: %v", err)
	}

	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply migrations for test DB: %v", err)
	}
	log.Println("Test database migrations applied successfully (or no changes needed).")

	testDB, err = database.Init(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize test database: %v", err)
	}
	defer testDB.Close()

	err = testDB.Ping()
	if err != nil {
		log.Fatalf("Failed to ping test database: %v", err)
	}

	auth, err := middleware.NewAuthenticator(cfg.JWTKey)
	if err != nil {
		log.Fatalf("Failed to create authenticator for tests: %v", err)
	}
	testAuthenticator = auth

	userRepo := repositories.NewUserRepository(testDB)
	userService := services.NewUserService(userRepo, testAuthenticator)
	testUserHandler = handlers.NewUserHandler(userService)

	// Initialize Auth components for auth routes
	authRepo := repositories.NewAuthRepository(testDB)
	authService := services.NewAuthService(authRepo)
	testAuthHandler = handlers.NewAuthHandler(authService) // Initialize package-level testAuthHandler

	// Initialize Product components
	productRepo := repositories.NewProductRepository(testDB)
	productService := services.NewProductService(productRepo)
	testProductHandler = handlers.NewProductHandler(productService)

	// Initialize Cart components
	cartRepo := repositories.NewCartRepository(testDB)
	cartService := services.NewCartService(cartRepo, productRepo)
	testCartHandler = handlers.NewCartHandler(cartService)

	// Initialize Order components
	orderRepo := repositories.NewOrderRepository(testDB)
	orderService := services.NewOrderService(orderRepo, cartRepo, productRepo)
	testOrderHandler := handlers.NewOrderHandler(orderService)

	// Initialize Review components
	reviewRepo := repositories.NewReviewRepository(testDB)
	reviewService := services.NewReviewService(reviewRepo, orderRepo)
	testReviewHandler := handlers.NewReviewHandler(reviewService)

	// Initialize Router
	testRouter = http.NewServeMux()

	// Register routes on testRouter
	// User public routes
	testRouter.HandleFunc("/users/register", testUserHandler.Register)
	testRouter.HandleFunc("/users/login", testUserHandler.Login)

	// User authenticated routes
	testRouter.Handle("/users/get-by-id", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.GetUserByID)))
	testRouter.Handle("/users/get-by-name", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.GetUserByName)))
	testRouter.Handle("/users/get-by-email", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.GetUserByEmail)))
	testRouter.Handle("/users/update", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.UpdateUser)))
	testRouter.Handle("/users/delete", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.DeleteUser)))

	// Auth routes (e.g., change password)
	testRouter.Handle("/auth/change-password", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testAuthHandler.ChangePassword)))

	// Product routes
	testRouter.HandleFunc("/products", testProductHandler.GetProducts)
	testRouter.HandleFunc("/product", testProductHandler.GetProduct)
	testRouter.HandleFunc("/products/category", testProductHandler.GetProductsByCategory)
	testRouter.HandleFunc("/products/user-id", testProductHandler.GetProductsByUserID)
	testRouter.Handle("/products/create", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testProductHandler.CreateProduct)))
	testRouter.Handle("/products/update", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testProductHandler.UpdateProduct)))
	testRouter.Handle("/products/delete", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testProductHandler.DeleteProduct)))

	// Cart routes
	testRouter.Handle("/cart/add", testAuthenticator.OptionalAuthenticateJWT(http.HandlerFunc(testCartHandler.AddToCart)))
	testRouter.Handle("/cart/remove", testAuthenticator.OptionalAuthenticateJWT(http.HandlerFunc(testCartHandler.RemoveFromCart)))
	testRouter.Handle("/cart/clear", testAuthenticator.OptionalAuthenticateJWT(http.HandlerFunc(testCartHandler.ClearCart)))
	testRouter.Handle("/cart/get", testAuthenticator.OptionalAuthenticateJWT(http.HandlerFunc(testCartHandler.GetCartItems)))
	testRouter.Handle("/cart/update", testAuthenticator.OptionalAuthenticateJWT(http.HandlerFunc(testCartHandler.UpdateCartItems)))

	// Order routes
	testRouter.Handle("/orders/checkout", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testOrderHandler.ProcessCheckout)))
	testRouter.Handle("/orders/confirm", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testOrderHandler.ConfirmOrder)))
	testRouter.Handle("/orders/summary", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testOrderHandler.GetOrderSummary)))
	testRouter.Handle("/orders/details", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testOrderHandler.GetOrderDetails)))
	testRouter.Handle("/orders/user", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testOrderHandler.GetUserOrders)))

	// Review routes
	testRouter.Handle("/reviews/add", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testReviewHandler.AddReview)))
	testRouter.HandleFunc("/reviews/get", testReviewHandler.GetReviewsByProductID)
	testRouter.Handle("/reviews/update", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testReviewHandler.UpdateReview)))
	testRouter.Handle("/reviews/delete", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testReviewHandler.DeleteReview)))

	code := m.Run()
	os.Exit(code)
}
