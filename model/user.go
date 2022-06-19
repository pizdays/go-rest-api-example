package model

import (
	"fmt"
	"regexp"
	"time"

	"gorm.io/gorm"
)

// User represents user schema.
type User struct {
	ID              uint           `json:"id" gorm:"type:BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY" example:"1"`
	OrganizationID  uint           `json:"organizationId" gorm:"type:INT NOT NULL;index" example:"1"`
	Name            string         `json:"name" gorm:"type:VARCHAR(191) NOT NULL;index" example:"John Doe"`
	Email           string         `json:"email" gorm:"type:VARCHAR(191) NOT NULL;unique" example:"john.doe@gmail.com"`
	EmailVerifiedAt *time.Time     `json:"emailVerifiedAt" example:"2021-02-01T01:01:00Z"`
	Password        string         `json:"-" gorm:"type:VARCHAR(191) NOT NULL" `
	CreatedAt       time.Time      `json:"createdAt" gorm:"not null" example:"2021-01-01T01:01:00Z"`
	UpdatedAt       time.Time      `json:"updatedAt" gorm:"not null" example:"2021-01-01T01:01:00Z"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index" swaggertype:"string"`
	PhoneNumber     string         `json:"phoneNumber" gorm:"type:VARCHAR(191) NULL" example:"0878889999"`
	Lang            string         `json:"lang" gorm:"type:VARCHAR(191) NULL" example:"th"`
	CurrentTeamID   uint           `json:"currentTeamId" gorm:"type:INT NOT NULL;index" example:"1"`

	// LineID represents user Line ID.
	LineID string `json:"lineId" gorm:"type:VARCHAR(255) DEFAULT '';not null" example:"johndoe"`
	// RoleID represents user role. Used for doing business logic.
	RoleID *string `json:"roleId"`
	Role   Role    `json:"-" gorm:"OnDelete:SET NULL;"`
}



var updatePwdSQLRe = regexp.MustCompile("^UPDATE.*password.*")

// ClientUser represents client format user.
type ClientUser struct {
	ID               uint       `json:"id" example:"1"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`

	Name             string     `json:"name"`
	Email            string     `json:"email"`
	EmailVerifiedAt  *time.Time `json:"emailVerifiedAt"`
	PhoneNumber      string     `json:"phoneNumber"`
	Lang             string     `json:"lang"`
	CurrentTeamID    uint       `json:"currentTeamId"`
	Role             Role       `json:"role"`
}

// ClientUserResponse represents user profile response.
type ClientUserResponse struct {
	Message string     `json:"message" example:"success"`
	Result  ClientUser `json:"result"`
}

// UserResponse represents response for user update.
type UserResponse struct {
	Data User `json:"data"`
}

// UsersResponse represents response including users with offset and limit
// applied and total user count.
type UsersResponse struct {
	Message string `json:"message" example:"success"`
	Result  struct {
		Count int64  `json:"count" example:"5"`
		Rows  []User `json:"rows"`
	} `json:"result"`
}

// CreateUserToTeamReqBody represents request body that creates user to team.
type CreateUserToTeamReqBody struct {
	Name   string `json:"name" binding:"required"`
	Email  string `json:"email" binding:"required,email"`
	RoleID string `json:"roleId" binding:"required,gt=0"`
}
