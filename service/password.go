package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-rest-api-example/model"
	"github.com/go-rest-api-example/util"

	"gorm.io/gorm"
)

const (
	passwordResetRecordExpiryTime int64 = 1 * 3600 // Seconds
)

// PasswordService represents password reset service.
type PasswordService struct {
	_       struct{}
	mySQL   *gorm.DB
	userSvc UserService
}

// NewPasswordService return password reset service instance.
func NewPasswordService(mySQL *gorm.DB, userSvc UserService) PasswordService {
	return PasswordService{
		mySQL:   mySQL,
		userSvc: userSvc,
	}
}

func (s PasswordService) CreatePasswordReset(ctx context.Context, email, token string) (model.PasswordReset, error) {
	pr := model.PasswordReset{
		Email: email,
		Token: token,
	}
	tx := s.mySQL.WithContext(ctx)

	err := tx.Create(&pr).Error
	if err != nil {
		tx.Rollback()
		return model.PasswordReset{}, fmt.Errorf("service.PasswordService.CreatePasswordReset: %w", err)
	}

	return pr, nil
}

func (s PasswordService) GetPasswordReset(ctx context.Context, email string) (model.PasswordReset, error) {
	var pr model.PasswordReset

	err := s.mySQL.WithContext(ctx).First(&pr, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.PasswordReset{}, nil
	}
	if err != nil {
		return model.PasswordReset{}, fmt.Errorf("service.PasswordService.GetPasswordReset: %w", err)
	}

	return pr, nil
}

func (s PasswordService) ValidateEmailAndToken(ctx context.Context, email, token string) (bool, error) {
	exist, err := s.checkPasswordResetExist(ctx, email, token)
	if err != nil {
		return false, fmt.Errorf("service.PasswordService.ValidateEmailAndToken: %w", err)
	}

	expired, err := s.CheckPasswordResetExpire(ctx, email, token)
	if err != nil {
		return false, fmt.Errorf("service.PasswordService.ValidateEmailAndToken: %w", err)
	}

	return exist && !expired, nil
}

func (s PasswordService) CheckPasswordResetExpire(ctx context.Context, email, token string) (bool, error) {
	var pr model.PasswordReset

	err := s.mySQL.WithContext(ctx).
		Where("email = ? AND token = ?", email, token).
		First(&pr).
		Error
	if err != nil {
		return false, fmt.Errorf("service.PasswordService.CheckPasswordResetExpire: %w", err)
	}

	// PasswordService reset expired.
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil
	}

	expired := time.Now().Unix()-pr.CreatedAt.Unix() > passwordResetRecordExpiryTime

	return expired, nil
}

func (s PasswordService) checkPasswordResetExist(ctx context.Context, email, token string) (bool, error) {
	var count int64

	err := s.mySQL.WithContext(ctx).
		Model(&model.PasswordReset{}).
		Where("email = ? AND token = ?", email, token).
		Count(&count).
		Error
	if err != nil {
		return false, fmt.Errorf("service.PasswordService.checkPasswordResetExist: %w", err)
	}

	return count >= 1, nil
}

func (s PasswordService) CheckPasswordResetEmailExpire(ctx context.Context, email string) bool {
	var pr model.PasswordReset

	err := s.mySQL.WithContext(ctx).
		Where("email = ?", email).
		First(&pr).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}

	expired := time.Now().Unix()-pr.CreatedAt.Unix() > passwordResetRecordExpiryTime
	if expired {
		_ = s.DeletePasswordResetRecord(ctx, email, pr.Token)
	}

	return expired
}

func (s PasswordService) CheckPasswordResetEmailExist(ctx context.Context, email string) bool {
	var count int64

	s.mySQL.WithContext(ctx).
		Model(&model.PasswordReset{}).
		Where("email = ?", email).
		Count(&count)

	return count >= 1
}

func (s PasswordService) EmailExist(ctx context.Context, email string) bool {
	var count int64

	s.mySQL.WithContext(ctx).
		Model(&model.User{}).
		Where("email = ?", email).
		Count(&count)

	return count >= 1
}

func (s PasswordService) DeletePasswordResetRecord(ctx context.Context, email, token string) error {
	err := s.mySQL.WithContext(ctx).
		Delete(&model.PasswordReset{},
			"email = ? AND token = ?",
			email,
			token).
		Error
	if err != nil {
		return fmt.Errorf("service.PasswordService.DeletePasswordResetRecord: %w", err)
	}

	return nil
}

// ResetPassword updates user password.
func (s PasswordService) ResetPassword(ctx context.Context, rp model.ResetPasswordParams) error {
	newHashedPassword, err := util.HashPassword(rp.Password)
	if err != nil {
		return fmt.Errorf("PasswordService.ResetPassword: %w", err)
	}

	usr, err := s.userSvc.GetByEmail(ctx, rp.Email)
	if err != nil {
		return fmt.Errorf("service.PasswordService.ResetPassword: %w", err)
	}

	// Update password
	_, err = s.userSvc.Update(ctx, usr.ID, map[string]interface{}{
		"password": newHashedPassword,
	})
	if err != nil {
		return fmt.Errorf("service.PasswordService.ResetPassword: %w", err)
	}

	return nil
}
