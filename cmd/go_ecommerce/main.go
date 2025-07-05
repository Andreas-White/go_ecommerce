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

	// server
	log.Printf("Server is listening on port %v", cfg.AppPort)
	err = http.ListenAndServe(cfg.AppPort, nil)
	if err != nil {
		log.Fatal("Server failed: ", err)
	}
}
