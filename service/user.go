package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-rest-api-example/model"
	"github.com/go-rest-api-example/util"

	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"
)

// ErrRevokedToken represents revoked refresh token error.
var ErrRevokedToken = errors.New("token is revoked")

// AuthService represents authentication service.
type AuthService struct {
	mySQL  *gorm.DB
	usrSvc UserService
}

// NewAuthService instantiates AuthService service.
func NewAuthService(mySQL *gorm.DB, usrSvc UserService) AuthService {
	return AuthService{mySQL: mySQL, usrSvc: usrSvc}
}

// LogIn returns access and refresh token for valid user credential.
func (s AuthService) LogIn(ctx context.Context, cred model.LoginParams) (model.TokenSet, error) {
	var user model.User

	err := s.mySQL.WithContext(ctx).
		First(&user,
			"email = ?",
			strings.ToLower(cred.Email)).
		Error
	if err != nil {
		return model.TokenSet{}, fmt.Errorf("service.AuthService.LogIn: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(cred.Password)); err != nil {
		return model.TokenSet{}, fmt.Errorf("service.AuthService.LogIn: %w", err)
	}

	accessTokenLifespan := time.Now().Add(time.Hour * 24 * 30) // 1 month ชั่วคราว
	accessTokenExpiredAt := accessTokenLifespan.Unix()
	if cred.IsLongLiveToken {
		accessTokenExpiredAt = 0 //Set long live token
	}

	accessTokenStr, err := util.CreateAccessToken(user.ID, accessTokenExpiredAt)
	if err != nil {
		return model.TokenSet{}, fmt.Errorf("service.AuthService.LogIn: %w", err)
	}

	refreshTokenLifespan := time.Now().Add(time.Hour * 24 * 30) // 1 month
	refreshTokenStr, err := util.CreateRefreshToken(user.ID, refreshTokenLifespan.Unix())
	if err != nil {
		return model.TokenSet{}, fmt.Errorf("service.AuthService.LogIn: %w", err)
	}

	tx := s.mySQL.Begin().WithContext(ctx)

	err = tx.Create(&model.RefreshToken{
		Token:   refreshTokenStr,
		Revoked: false,
	}).Error
	if err != nil {
		tx.Rollback()
		return model.TokenSet{}, fmt.Errorf("service.AuthService.LogIn: %w", err)
	}

	// Create role for old (backward compatibility).
	if user.RoleID == nil {
		user.Role = model.Role{
			Name:        model.RoleAdmin,
			Description: "Administrator",
			Permissions: model.AllPermissions,
			TeamID:      user.OrganizationID,
		}

		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			return model.TokenSet{}, fmt.Errorf("service.AuthService.LogIn: %w", err)
		}
	}

	err = tx.Create(&model.ActivityLog{
		UserID:       user.ID,
		ActivityType: model.ActTypeLogin,
	}).Error
	if err != nil {
		tx.Rollback()
		return model.TokenSet{}, fmt.Errorf("service.AuthService.LogIn: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return model.TokenSet{}, fmt.Errorf("service.AuthService.LogIn: %w", err)
	}

	return model.TokenSet{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
	}, nil
}

// RefreshToken returns new access token.
func (s AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	var rt model.RefreshToken

	err := s.mySQL.WithContext(ctx).
		First(&rt, "token = ?", refreshToken).Error

	if err != nil {
		return "", fmt.Errorf("service.AuthService.RefreshToken: %w", err)
	}

	if rt.Revoked {
		return "", fmt.Errorf("service.AuthService.RefreshToken: %w", ErrRevokedToken)
	}

	claims, err := util.ParseToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("service.AuthService.RefreshToken: %w", err)
	}

	if claims["type"] != "refresh" {
		err = errors.New("wrong token type")
		return "", fmt.Errorf("service.AuthService.RefreshToken: %w", err)
	}

	accessTokenLifespan := time.Now().Add(time.Hour * 24 * 30) // 1 month ชั่วคราว
	accessTokenStr, err := util.CreateAccessToken(uint(claims["id"].(float64)), accessTokenLifespan.Unix())
	if err != nil {
		return "", fmt.Errorf("service.AuthService.RefreshToken: %w", err)
	}

	return accessTokenStr, nil
}

// LogOut revokes refresh token if not expired.
func (s AuthService) LogOut(ctx context.Context, refreshToken string) error {
	tx := s.mySQL.WithContext(ctx)

	err := tx.Model(model.RefreshToken{}).
		Where("token = ?", refreshToken).
		Update("revoked", true).
		Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("service.AuthService.LogOut: %w", err)
	}

	usr, err := s.usrSvc.GetByToken(ctx, refreshToken)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("service.AuthService.LogOut: %w", err)
	}

	err = tx.Create(&model.ActivityLog{
		UserID:       usr.ID,
		ActivityType: model.ActTypeLogout,
	}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("service.AuthService.LogOut: %w", err)
	}

	return nil
}

// LogInByLine returns access and refresh token for valid Line ID and Line
// access token.
func (s AuthService) LogInByLine(ctx context.Context, lineID string) (model.TokenSet, error) {
	var usr model.User

	err := s.mySQL.WithContext(ctx).
		First(&usr, "line_id = ?", lineID).
		Error
	if err != nil {
		return model.TokenSet{}, fmt.Errorf("service.AuthService.LogInByLine: %w", err)
	}
	accessTokenLifespan := time.Now().Add(time.Hour * 24 * 30) // 1 month ชั่วคราว
	// Generate access token.
	accTok, err := util.CreateAccessToken(usr.ID, accessTokenLifespan.Unix())
	if err != nil {
		return model.TokenSet{}, fmt.Errorf("service.AuthService.LogInByLine: %w", err)
	}

	refreshTokenLifespan := time.Now().Add(time.Hour * 24 * 7)

	// Generate refresh token.
	refreshTok, err := util.CreateRefreshToken(usr.ID, refreshTokenLifespan.Unix())
	if err != nil {
		return model.TokenSet{}, fmt.Errorf("service.AuthService.LogInByLine: %w", err)
	}

	tx := s.mySQL.WithContext(ctx).Begin()

	// Create refresh token record.
	err = tx.Create(&model.RefreshToken{
		Token:   refreshTok,
		Revoked: false,
	}).Error
	if err != nil {
		tx.Rollback()
		return model.TokenSet{}, fmt.Errorf("service.AuthService.LogInByLine: %w", err)
	}

	err = tx.Create(&model.ActivityLog{
		UserID:       usr.ID,
		ActivityType: model.ActTypeLogin,
	}).Error
	if err != nil {
		tx.Rollback()
		return model.TokenSet{}, fmt.Errorf("service.AuthService.LogInByLine: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return model.TokenSet{}, fmt.Errorf("service.AuthService.LogInByLine: %w", err)
	}

	return model.TokenSet{
		AccessToken:  accTok,
		RefreshToken: refreshTok,
	}, nil
}
