package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PasswordReset represents password reset schema.
type PasswordReset struct {
	Email     string         `json:"email" gorm:"type:VARCHAR(191);not null;index"`
	Token     string         `json:"token" gorm:"type:VARCHAR(191);not null"`
	CreatedAt time.Time      `json:"createdAt" gorm:"not null" `
	UpdatedAt time.Time      `json:"updatedAt" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (pr *PasswordReset) AfterCreate(tx *gorm.DB) error {
	v := tx.Statement.Context.Value(CtxUser)
	if v == nil {
		return nil
	}

	u := v.(User)

	err := tx.Create(&ActivityLog{
		UserID:       u.ID,
		ActivityType: ActTypePwdChangeReq,
	}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("model.PasswordReset.AfterCreate: %w", err)
	}

	return nil
}

// ResetPasswordParams represents parameters for resetting user password.
type ResetPasswordParams struct {
	Email    string `json:"email" form:"email" binding:"required,email" example:"john.doe@gmail.com" validate:"required"`
	Password string `json:"password" form:"password" binding:"required" example:"139482345" validate:"required"`
	Token    string `json:"token" form:"token" binding:"required" example:"7FkZA0yz0yYPXH7oqns16nc5BZMSgvkHxBdhWPkjrLjL8c4wEic1kH5Ms9WI5Eefh0heOrQdgZr0pCKxTS1BUJILPW3a36Kb4n1JnEoWhpVMQW_X9LRRHYwH04aRWkjTT_-UjYa63o1g6Lm-wv1Shcj7byB_Aryzv6L_kjdMzK4=" validate:"required"`
}

// PasswordResetValidationParams represents password reset validation parameters.
type PasswordResetValidationParams struct {
	Email string `json:"email" form:"email" binding:"required,email"`
	Token string `json:"token" form:"token" binding:"required"`
}

// PasswordResetValidityResponse represents validity of a combination of email
// and password reset token response.
type PasswordResetValidityResponse struct {
	// Valid represents password reset record with a combination of email and
	// password reset token validity.
	Valid bool `json:"valid" example:"true"`
}

// PasswordResetExpirationResponse represents password reset record expiration
// response.
type PasswordResetExpirationResponse struct {
	// Expired represents whether password reset record with a combination of
	// email and password reset token expired.
	Expired bool `json:"expired" example:"false"`
}

