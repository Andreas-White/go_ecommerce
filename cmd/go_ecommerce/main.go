package main

import (
	"fmt"
	"go_ecommerce/internal/config"
	"go_ecommerce/pkg/database"
	"go_ecommerce/pkg/handlers"
	"go_ecommerce/pkg/middleware"
	"go_ecommerce/pkg/repositories"
	"go_ecommerce/pkg/services"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	config := config.LoadConfig()

	database.Init()
	defer database.CloseDB()

	userRepo := repositories.NewUserRepository()
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	// Define routes
	http.HandleFunc("/users/register", userHandler.Register)
	http.HandleFunc("/users/login", userHandler.Login)
	http.Handle("/users/get_by_id", middleware.AuthenticateJWT(http.HandlerFunc(userHandler.GetUserByID)))
	http.Handle("/users/get_by_name", middleware.AuthenticateJWT(http.HandlerFunc(userHandler.GetUserByName)))
	http.Handle("/users/get_by_email", middleware.AuthenticateJWT(http.HandlerFunc(userHandler.GetUserByEmail)))
	http.Handle("/users/update", middleware.AuthenticateJWT(http.HandlerFunc(userHandler.UpdateUser)))
	http.Handle("/users/delete", middleware.AuthenticateJWT(http.HandlerFunc(userHandler.DeleteUser)))

	// server
	fmt.Println("Server is listening on port 8080")
	http.ListenAndServe(config.AppPort, nil)
}
