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

	userRepo := repositories.NewUserRepository(DB)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService, authMiddleware)

	// Define routes
	http.HandleFunc("/users/register", userHandler.Register)
	http.HandleFunc("/users/login", userHandler.Login)
	http.Handle("/users/get_by_id", authMiddleware.AuthenticateJWT(http.HandlerFunc(userHandler.GetUserByID)))
	http.Handle("/users/get_by_name", authMiddleware.AuthenticateJWT(http.HandlerFunc(userHandler.GetUserByName)))
	http.Handle("/users/get_by_email", authMiddleware.AuthenticateJWT(http.HandlerFunc(userHandler.GetUserByEmail)))
	http.Handle("/users/update", authMiddleware.AuthenticateJWT(http.HandlerFunc(userHandler.UpdateUser)))
	http.Handle("/users/delete", authMiddleware.AuthenticateJWT(http.HandlerFunc(userHandler.DeleteUser)))

	// server
	log.Printf("Server is listening on port %v", cfg.AppPort)
	err = http.ListenAndServe(cfg.AppPort, nil)
	if err != nil {
		log.Fatal("Server failed: ", err)
	}
}
