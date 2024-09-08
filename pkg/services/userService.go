package services

import (
	"errors"
	models "go_ecommerce/pkg/models/user"
	"go_ecommerce/pkg/repositories"
	"go_ecommerce/pkg/utils"

	"github.com/google/uuid"
)

// UserService struct
type UserService struct {
	UserRepo *repositories.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		UserRepo: userRepo,
	}
}

// CreateUser calls the repository to insert a new user into the database
func (s *UserService) CreateUser(user *models.UserRegister) error {
	if user.Name == "" || user.Email == "" {
		return errors.New("name and email are required")
	}

	return s.UserRepo.CreateUser(user)
}

// GetUserByID fetches a user by ID from the repository
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	return s.UserRepo.GetUserByID(id)
}

// GetUserByName fetches a user by name from the repository
func (s *UserService) GetUserByName(name string) (*models.User, error) {
	return s.UserRepo.GetUserByName(name)
}

// GetUserByEmail fetches a user by email from the repository
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	return s.UserRepo.GetUserByEmail(email)
}

// UpdateUser updates a user’s information in the database
func (s *UserService) UpdateUser(user *models.User) error {
	if user.ID == uuid.Nil {
		return errors.New("user ID is required")
	}

	return s.UserRepo.UpdateUser(user)
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(id string) error {
	if id == "" {
		return errors.New("user ID is required")
	}

	return s.UserRepo.DeleteUser(id)
}

func (s *UserService) AuthenticateUser(email, password string) (*models.User, error) {
	// Example logic: fetch user by email and compare hashed passwords
	authedUser, err := s.UserRepo.GetAuthedUserByEmail(email)
	if err != nil || !utils.CheckPasswordHash(password, authedUser.Auth.Password) {
		return nil, errors.New("invalid email or password")
	}

	user := &models.User{
		ID:    authedUser.ID,
		Name:  authedUser.Name,
		Email: authedUser.Email,
		Phone: authedUser.Phone,
	}

	return user, nil
}
