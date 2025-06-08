package services

import (
	"context"
	"database/sql"
	"fmt"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/repositories"
	"go_ecommerce/pkg/utils"
)

type IAuthService interface {
	ChangePassword(ctx context.Context, userID string, changePasswordDTO *models.ChangePasswordDTO) error
}

type AuthService struct {
	AuthRepo repositories.IAuthRepository
}

func NewAuthService(authRepo repositories.IAuthRepository) *AuthService {
	return &AuthService{
		AuthRepo: authRepo,
	}
}

func (s *AuthService) ChangePassword(ctx context.Context, userID string, changePasswordDTO *models.ChangePasswordDTO) error {
	if changePasswordDTO.CurrentPassword == "" || changePasswordDTO.NewPassword == "" {
		return utils.HandleServiceErrors(ctx, fmt.Errorf("current_password and new_password are required"), "service/ChangePassword")
	}
	if len(changePasswordDTO.NewPassword) < 8 {
		return utils.HandleServiceErrors(ctx, fmt.Errorf("new_password must be at least 8 characters long"), "service/ChangePassword")
	}
	if changePasswordDTO.CurrentPassword == changePasswordDTO.NewPassword {
		return utils.HandleServiceErrors(ctx, fmt.Errorf("new password cannot be the same as the current password"), "service/ChangePassword")
	}

	auth, err := s.AuthRepo.GetAuthByUserID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.HandleServiceErrors(ctx, fmt.Errorf("user not found"), "service/ChangePassword")
		}
		return utils.HandleServiceErrors(ctx, err, "service/ChangePassword")
	}

	if !utils.CheckPasswordHash(changePasswordDTO.CurrentPassword, auth.Password) {
		return utils.HandleServiceErrors(ctx, fmt.Errorf("incorrect current password"), "service/ChangePassword")
	}
	newHashedPassword := utils.HashPassword(changePasswordDTO.NewPassword)

	err = s.AuthRepo.UpdatePassword(ctx, userID, newHashedPassword)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/ChangePassword")
	}

	return nil
}
