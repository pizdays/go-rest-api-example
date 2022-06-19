package model

import (
	"time"

	"gorm.io/gorm"
)

// RoleAdmin represents administrator role.
var RoleAdmin = "Admin"

// Role represents role schema.
type Role struct {
	ID        string         `json:"id" gorm:"primaryKey;size:36"`
	CreatedAt time.Time      `json:"createdAt" gorm:"type:TIMESTAMP DEFAULT CURRENT_TIMESTAMP();not null"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"type:TIMESTAMP DEFAULT CURRENT_TIMESTAMP();not null"`
	DeletedAt gorm.DeletedAt `json:"-"`

	// Name represents role name.
	Name        string       `json:"name" gorm:"size:128;not null"`
	Description string       `json:"description" gorm:"size:1024;not null"`
	TeamID      uint         `json:"teamId" gorm:"type:INT UNSIGNED"`
	Team        Team         `json:"-"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions"`
}


// RolePermission represents associative table for role and permission.
type RolePermission struct {
	RoleID       string    `gorm:"primaryKey"`
	PermissionID string    `gorm:"primaryKey"`
	CreatedAt    time.Time `json:"createdAt" gorm:"type:TIMESTAMP DEFAULT CURRENT_TIMESTAMP();not null"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"type:TIMESTAMP DEFAULT CURRENT_TIMESTAMP();not null"`
}

// RoleReqBody represents role request body for creating and updating.
type RoleReqBody struct {
	Name          string   `json:"name" binding:"required,ne=Admin"`
	Description   string   `json:"description"`
	PermissionIDs []string `json:"permissionIDs" binding:"gt=0,dive,uuid4"`
}

