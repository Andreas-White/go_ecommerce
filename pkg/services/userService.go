package services

import (
	"context"
	"errors"
	"go_ecommerce/pkg/middleware"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/repositories"
	"go_ecommerce/pkg/utils"
	"log"

	"github.com/google/uuid"
)

type IUserService interface {
	CreateUser(ctx context.Context, user *models.UserDTO) error
	GetUserByID(ctx context.Context, id string) (*models.UserDTO, error)
	GetUserByName(ctx context.Context, firstName string, lastName string, middleName string) (*models.UserDTO, error)
	GetUserByEmail(ctx context.Context, email string) (*models.UserDTO, error)
	UpdateUser(ctx context.Context, user *models.User, userPayload *models.UserDTO) (*models.User, string, error)
	DeleteUser(ctx context.Context, id string) error
	AuthenticateUser(ctx context.Context, email, password string) (string, error)
}

// UserService struct
type UserService struct {
	UserRepo       repositories.IUserRepository
	TokenGenerator middleware.TokenGenerator
}

// NewUserService creates a new UserService
func NewUserService(userRepo repositories.IUserRepository, tokenGenerator middleware.TokenGenerator) IUserService {
	return &UserService{
		UserRepo:       userRepo,
		TokenGenerator: tokenGenerator,
	}
}

// CreateUser calls the repository to insert a new user into the database
func (s *UserService) CreateUser(ctx context.Context, user *models.UserDTO) error {
	if user.FirstName == "" || user.LastName == "" || user.Email == "" {
		return utils.HandleServiceErrors(ctx, errors.New("full name and email are required"), "service/CreateUser")
	}

	if len(user.Password) < 8 {
		return utils.HandleServiceErrors(ctx, errors.New("password must be at least 8 characters long"), "service/CreateUser")
	}

	if !utils.IsValidEmail(user.Email) {
		return utils.HandleServiceErrors(ctx, errors.New("invalid email format"), "service/CreateUser")
	}

	err := s.UserRepo.CreateUser(ctx, user)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/CreateUser")
	}

	return nil
}

// GetUserByID fetches a user by ID from the repository
func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.UserDTO, error) {
	user, err := s.UserRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/GetUserByID")
	}
	return toUserDTO(user), nil
}

// GetUserByName fetches a user by name from the repository
func (s *UserService) GetUserByName(ctx context.Context, firstName string, lastName string, middleName string) (*models.UserDTO, error) {
	user, err := s.UserRepo.GetUserByFullName(ctx, firstName, lastName, middleName)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/GetUserByName")
	}
	return toUserDTO(user), nil
}

// GetUserByEmail fetches a user by email from the repository
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.UserDTO, error) {
	user, err := s.UserRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/GetUserByEmail")
	}
	return toUserDTO(user), nil
}

// UpdateUser updates a user’s information in the database
func (s *UserService) UpdateUser(ctx context.Context, authUser *models.User, userPayload *models.UserDTO) (*models.User, string, error) {
	if authUser.ID == uuid.Nil {
		return nil, "", utils.HandleServiceErrors(ctx, errors.New("user ID is required"), "service/UpdateUser")
	}

	if !utils.IsValidEmail(authUser.Email) {
		return nil, "", utils.HandleServiceErrors(ctx, errors.New("invalid email format"), "service/UpdateUser")
	}

	updatedUser := &models.User{
		ID:         authUser.ID,
		FirstName:  stringOrDefault(userPayload.FirstName, authUser.FirstName),
		LastName:   stringOrDefault(userPayload.LastName, authUser.LastName),
		MiddleName: stringOrDefault(userPayload.MiddleName, authUser.MiddleName),
		Email:      stringOrDefault(userPayload.Email, authUser.Email),
		Phone:      int64OrDefault(userPayload.Phone, authUser.Phone),
		IsProducer: boolOrDefault(userPayload.IsProducer, authUser.IsProducer),
		Address:    stringOrDefault(userPayload.Address, authUser.Address),
		City:       stringOrDefault(userPayload.City, authUser.City),
		Country:    stringOrDefault(userPayload.Country, authUser.Country),
		ZipCode:    int32OrDefault(userPayload.ZipCode, authUser.ZipCode),
	}

	err := s.UserRepo.UpdateUser(ctx, updatedUser)
	if err != nil {
		return nil, "", utils.HandleServiceErrors(ctx, err, "service/UpdateUser")
	}

	newToken, err := s.TokenGenerator.GenerateJWT(updatedUser)
	if err != nil {
		return nil, "", utils.HandleServiceErrors(ctx, err, "service/UpdateUser")
	}

	return updatedUser, newToken, nil
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return utils.HandleServiceErrors(ctx, errors.New("user ID is required"), "service/DeleteUser")
	}

	err := s.UserRepo.DeleteUser(ctx, id)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/DeleteUser")
	}
	return nil
}

func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (string, error) {
	authedUser, err := s.UserRepo.GetAuthedUserByEmail(ctx, email)
	if err != nil {
		return "", utils.HandleServiceErrors(ctx, err, "service/AuthenticateUser")
	}

	if !utils.CheckPasswordHash(password, authedUser.Auth.Password) {
		return "", utils.HandleServiceErrors(ctx, errors.New("incorrect password"), "service/AuthenticateUser")
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

	// Generate JWT token for authenticated user
	token, err := s.TokenGenerator.GenerateJWT(user)
	if err != nil {
		log.Printf("{handler/Login - Error generating JWT token: %v}", err)
		return "", utils.HandleServiceErrors(ctx, err, "service/AuthenticateUser")
	}

	return token, nil
}

func toUserDTO(user *models.User) *models.UserDTO {
	return &models.UserDTO{
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Email:      user.Email,
		Phone:      user.Phone,
		IsProducer: user.IsProducer,
		Address:    user.Address,
		City:       user.City,
		Country:    user.Country,
		ZipCode:    user.ZipCode,
	}
}

func stringOrDefault(newValue string, oldValue string) string {
	if newValue != "" {
		return newValue
	}

	return oldValue
}

func int64OrDefault(newValue int64, oldValue int64) int64 {
	if newValue != 0 {
		return newValue
	}

	return oldValue
}

func int32OrDefault(newValue int32, oldValue int32) int32 {
	if newValue != 0 {
		return newValue
	}

	return oldValue
}

func boolOrDefault(newValue bool, oldValue bool) bool {
	if newValue {
		return newValue
	}

	return oldValue
}
