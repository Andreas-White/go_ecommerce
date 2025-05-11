package services

import (
	"context"
	"fmt"
	"go_ecommerce/pkg/models"
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
func (s *UserService) CreateUser(ctx context.Context, user *models.UserRegister) error {
	if user.FirstName == "" || user.LastName == "" || user.Email == "" {
		return fmt.Errorf("{service/CreateUser - full name and email are required}")
	}

	err := s.UserRepo.CreateUser(ctx, user)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("{service/CreateUser - context error during CreateUser: %w}", ctx.Err())
		}
		return fmt.Errorf("{service/CreateUser - failed to create user in repository: %w}", err)
	}

	return nil
}

// GetUserByID fetches a user by ID from the repository
func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, err := s.UserRepo.GetUserByID(ctx, id)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("{service/GetUserByID - context error during GetUserByID: %w}", ctx.Err())
		}
		return nil, fmt.Errorf("{service/GetUserByID - failed to get user by ID from repository: %w}", err)
	}
	return user, nil
}

// GetUserByName fetches a user by name from the repository
func (s *UserService) GetUserByName(ctx context.Context, firstName string, lastName string, middleName string) (*models.User, error) {
	user, err := s.UserRepo.GetUserByFullName(ctx, firstName, lastName, middleName)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("{service/GetUserByName - context error during GetUserByName: %w}", ctx.Err())
		}
		return nil, fmt.Errorf("{service/GetUserByName - failed to get user by name from repository: %w}", err)
	}
	return user, nil
}

// GetUserByEmail fetches a user by email from the repository
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.UserRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("{service/GetUserByEmail - context error during GetUserByEmail: %w}", ctx.Err())
		}
		return nil, fmt.Errorf("{service/GetUserByEmail - failed to get user by email from repository: %w}", err)
	}
	return user, nil
}

// UpdateUser updates a user’s information in the database
func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil { // Use uuid.Nil for comparison with uuid.UUID
		return fmt.Errorf("{service/UpdateUser - user ID is required}")
	}

	err := s.UserRepo.UpdateUser(ctx, user)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("{service/UpdateUser - context error during UpdateUser: %w}", ctx.Err())
		}
		return fmt.Errorf("{service/UpdateUser - failed to update user in repository: %w}", err)
	}
	return nil
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("{service/DeleteUser - user ID is required}")
	}

	// Pass the context down to the repository
	err := s.UserRepo.DeleteUser(ctx, id)
	if err != nil {
		// Check if the error from the repository was a context error
		if ctx.Err() != nil {
			return fmt.Errorf("{service/DeleteUser - context error during DeleteUser: %w}", ctx.Err())
		}
		// Wrap other errors from the repository
		return fmt.Errorf("{service/DeleteUser - failed to delete user in repository: %w}", err)
	}
	return nil
}

func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
	authedUser, err := s.UserRepo.GetAuthedUserByEmail(ctx, email)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("{service/AuthenticateUser - context error during AuthenticateUser: %w}", ctx.Err())
		}
		return nil, fmt.Errorf("{service/AuthenticateUser - error getting authed user: %w}", err)
	}

	// Check password hash - this is application logic, not a database call
	if !utils.CheckPasswordHash(password, authedUser.Auth.Password) {
		return nil, fmt.Errorf("{service/AuthenticateUser - incorrect password}")
	}

	user := &models.User{
		ID:         authedUser.ID,
		FirstName:  authedUser.FirstName,
		LastName:   authedUser.LastName,
		MiddleName: authedUser.MiddleName,
		Email:      authedUser.Email,
		Phone:      authedUser.Phone,
		IsProducer: authedUser.IsProducer,
		Address:    authedUser.Address,
		City:       authedUser.City,
		Country:    authedUser.Country,
		ZipCode:    authedUser.ZipCode,
		CreatedAt:  authedUser.CreatedAt,
		UpdatedAt:  authedUser.UpdatedAt,
	}

	return user, nil
}
