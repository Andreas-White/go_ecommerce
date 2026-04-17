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

	"go_ecommerce/pkg/utils"

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

	// Initialize Company components
	testCompanyRepo := repositories.NewCompanyRepository(testDB)
	testCompanyService := services.NewCompanyService(testCompanyRepo)
	testCompanyHandler := handlers.NewCompanyHandler(testCompanyService)

	// Initialize Review components (needed for ProductService)
	testReviewRepo := repositories.NewReviewRepository(testDB)

	// Initialize Product components
	productRepo := repositories.NewProductRepository(testDB)
	productService := services.NewProductService(productRepo, testCompanyRepo, testReviewRepo)
	testProductHandler = handlers.NewProductHandler(productService)

	// Initialize Cart components
	cartRepo := repositories.NewCartRepository(testDB)
	cartService := services.NewCartService(cartRepo, productRepo)
	testCartHandler = handlers.NewCartHandler(cartService)

	// Initialize Order components
	orderRepo := repositories.NewOrderRepository(testDB)
	orderService := services.NewOrderService(orderRepo, cartRepo, productRepo)
	testOrderHandler := handlers.NewOrderHandler(orderService)

	// Initialize Review service
	testReviewService := services.NewReviewService(testReviewRepo, orderRepo, testCompanyRepo, productRepo)
	testReviewHandler := handlers.NewReviewHandler(testReviewService)

	// Initialize Router
	testRouter = http.NewServeMux()

	// Register routes on testRouter
	// User public routes
	testRouter.Handle("/users/register", middleware.CSRFMiddleware(http.HandlerFunc(testUserHandler.Register)))
	testRouter.Handle("/users/login", middleware.CSRFMiddleware(http.HandlerFunc(testUserHandler.Login)))

	// CSRF endpoint for guest users
	testRouter.HandleFunc("/csrf", func(w http.ResponseWriter, r *http.Request) {
		csrfToken := utils.GenerateCSRFToken(32)
		csrfCookie := &http.Cookie{
			Name:     "csrf_token",
			Value:    csrfToken,
			Path:     "/",
			HttpOnly: false,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   60 * 60 * 24,
		}
		http.SetCookie(w, csrfCookie)
		w.WriteHeader(http.StatusNoContent)
	})

	// User authenticated routes
	testRouter.Handle("/users/get-by-id", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.GetUserByID)))
	testRouter.Handle("/users/get-by-name", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.GetUserByName)))
	testRouter.Handle("/users/get-by-email", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.GetUserByEmail)))
	testRouter.Handle("/users/update", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testUserHandler.UpdateUser))))
	testRouter.Handle("/users/delete", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testUserHandler.DeleteUser))))

	// Auth routes (e.g., change password)
	testRouter.Handle("/auth/change-password", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testAuthHandler.ChangePassword))))

	// Product routes
	testRouter.HandleFunc("/products", testProductHandler.GetProducts)
	testRouter.HandleFunc("/product", testProductHandler.GetProduct)
	testRouter.HandleFunc("/products/category", testProductHandler.GetProductsByCategory)
	testRouter.Handle("/products/my-products", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testProductHandler.GetProductsByUserID)))
	testRouter.Handle("/products/create", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testProductHandler.CreateProduct))))
	testRouter.Handle("/products/update", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testProductHandler.UpdateProduct))))
	testRouter.Handle("/products/delete", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testProductHandler.DeleteProduct))))

	// Cart routes
	testRouter.Handle("/cart/add", testAuthenticator.OptionalAuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testCartHandler.AddToCart))))
	testRouter.Handle("/cart/remove", testAuthenticator.OptionalAuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testCartHandler.RemoveFromCart))))
	testRouter.Handle("/cart/clear", testAuthenticator.OptionalAuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testCartHandler.ClearCart))))
	testRouter.Handle("/cart/get", testAuthenticator.OptionalAuthenticateJWT(http.HandlerFunc(testCartHandler.GetCartItems)))
	testRouter.Handle("/cart/update", testAuthenticator.OptionalAuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testCartHandler.UpdateCartItems))))

	// Order routes
	testRouter.Handle("/orders/checkout", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testOrderHandler.ProcessCheckout))))
	testRouter.Handle("/orders/confirm", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testOrderHandler.ConfirmOrder))))
	testRouter.Handle("/orders/summary", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testOrderHandler.GetOrderGroupSummary)))
	testRouter.Handle("/orders/details", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testOrderHandler.GetOrderDetails)))
	testRouter.Handle("/orders/user", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testOrderHandler.GetUserOrders)))
	testRouter.Handle("/orders/producer", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testOrderHandler.GetProducerOrders)))
	testRouter.Handle("/orders/fulfill", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testOrderHandler.FulfillOrder))))
	testRouter.Handle("/orders/sales-report", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testOrderHandler.GetSalesReport)))
	testRouter.Handle("/orders/cancel", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testOrderHandler.CancelOrder))))
	testRouter.Handle("/orders/delete", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testOrderHandler.CustomerDeleteOrder))))

	// Review routes
	testRouter.Handle("/reviews/add", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testReviewHandler.AddReview))))
	testRouter.HandleFunc("/reviews/get", testReviewHandler.GetReviewsByProductID)
	testRouter.Handle("/reviews/update", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testReviewHandler.UpdateReview))))
	testRouter.Handle("/reviews/delete", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testReviewHandler.DeleteReview))))

	// Company routes
	testRouter.Handle("/companies/create", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testCompanyHandler.CreateCompany))))
	testRouter.Handle("/companies/get-by-user", testAuthenticator.AuthenticateJWT(http.HandlerFunc(testCompanyHandler.GetCompanyByUserID)))
	testRouter.HandleFunc("/companies/get-by-id", testCompanyHandler.GetCompanyByCompanyID)
	testRouter.Handle("/companies/update", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testCompanyHandler.UpdateCompany))))
	testRouter.Handle("/companies/delete", testAuthenticator.AuthenticateJWT(middleware.CSRFMiddleware(http.HandlerFunc(testCompanyHandler.DeleteCompany))))

	code := m.Run()
	os.Exit(code)
}
