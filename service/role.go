package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-rest-api-example/model"

	"gorm.io/gorm"
)

var (
	// ErrRoleNotFound represents role not found error.
	ErrRoleNotFound = errors.New("role not found")

	// ErrAdminRoleCannotBeModified represents error that occurs when admin role record are going
	// to be modified (update, deletion).
	ErrAdminRoleCannotBeModified = errors.New("admin role cannot be modified")
)

// RoleService represents role service.
type RoleService struct {
	mySQL *gorm.DB
}

// NewRoleService returns new role service instance.
func NewRoleService(mySQL *gorm.DB) RoleService {
	return RoleService{mySQL: mySQL}
}

func (r RoleService) Create(ctx context.Context, role model.Role) (model.Role, error) {
	permIDs := make([]string, len(role.Permissions))

	for i, perm := range role.Permissions {
		permIDs[i] = perm.ID
	}

	var count int64

	tx := r.mySQL.WithContext(ctx).Begin()

	err := tx.Model(&model.Permission{}).
		Where("id IN (?)", permIDs).
		Count(&count).
		Error
	if err != nil {
		tx.Rollback()
		return model.Role{}, fmt.Errorf("service.RoleService.Create: %w", err)
	}

	if int(count) != len(role.Permissions) {
		return model.Role{}, fmt.Errorf("service.RoleService.Create: %w", ErrPermissionNotFound)
	}

	err = tx.Create(&role).Error
	if err != nil {
		tx.Rollback()
		return model.Role{}, fmt.Errorf("service.RoleService.Create: %w", err)
	}

	err = tx.Model(&role).
		Association("Permissions").
		Find(&role.Permissions)
	if err != nil {
		tx.Rollback()
		return model.Role{}, fmt.Errorf("service.RoleService.Create: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return model.Role{}, fmt.Errorf("service.RoleService.Create: %w", err)
	}

	return role, nil
}

func (r RoleService) CountByTeamID(ctx context.Context, teamID uint) (int64, error) {
	var count int64

	err := r.mySQL.WithContext(ctx).
		Model(&model.Role{}).
		Where("team_id = ?", teamID).
		Count(&count).
		Error
	if err != nil {
		return 0, fmt.Errorf("service.RoleService.CountByTeamID: %w", err)
	}

	return count, nil
}

func (r RoleService) FindAllByTeamID(ctx context.Context, teamID uint, query model.PagingQuery) ([]model.Role, error) {
	var roles []model.Role

	err := r.mySQL.WithContext(ctx).
		Preload("Permissions").
		Offset(query.Offset).
		Limit(query.Limit).
		Order("created_at DESC").
		Find(&roles, "team_id = ?", teamID).
		Error
	if err != nil {
		return nil, fmt.Errorf("service.RoleService.FindAllByTeamID: %w", err)
	}

	return roles, nil
}

func (r RoleService) Update(ctx context.Context, role model.Role) (model.Role, error) {
	tx := r.mySQL.WithContext(ctx).Begin()

	var tmpRole model.Role

	err := tx.Select("name", "created_at").
		Where(&model.Role{ID: role.ID}).
		First(&tmpRole).
		Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return model.Role{}, fmt.Errorf("service.RoleService.Update: %w", ErrRoleNotFound)
	case err != nil:
		return model.Role{}, fmt.Errorf("service.RoleService.Update: %w", err)
	}

	if tmpRole.Name == model.RoleAdmin {
		return model.Role{}, fmt.Errorf("service.RoleService.Update: %w", ErrAdminRoleCannotBeModified)
	}

	permIDs := make([]string, len(role.Permissions))

	for i, perm := range role.Permissions {
		permIDs[i] = perm.ID
	}

	var count int64

	err = tx.Model(&model.Permission{}).
		Where("id IN (?)", permIDs).
		Count(&count).
		Error
	if err != nil {
		return model.Role{}, fmt.Errorf("service.RoleService.Update: %w", err)
	}

	if int(count) != len(role.Permissions) {
		return model.Role{}, fmt.Errorf("service.RoleService.Update: %w", ErrPermissionNotFound)
	}

	role.CreatedAt = tmpRole.CreatedAt

	err = tx.Save(&role).Error
	if err != nil {
		tx.Rollback()
		return model.Role{}, fmt.Errorf("service.RoleService.Update: %w", err)
	}

	err = tx.Model(&role).
		Association("Permissions").
		Replace(role.Permissions)
	if err != nil {
		tx.Rollback()
		return model.Role{}, fmt.Errorf("service.RoleService.Update: %w", err)
	}

	err = tx.Model(&role).
		Association("Permissions").
		Find(&role.Permissions)
	if err != nil {
		tx.Rollback()
		return model.Role{}, fmt.Errorf("service.RoleService.Update: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return model.Role{}, fmt.Errorf("service.RoleService.Update: %w", err)
	}

	return role, nil
}

func (r RoleService) DeleteByID(ctx context.Context, roleID string) error {
	var tmpRole model.Role

	err := r.mySQL.WithContext(ctx).
		Select("name", "created_at").
		Where(&model.Role{ID: roleID}).
		First(&tmpRole).
		Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return fmt.Errorf("service.RoleService.Update: %w", ErrRoleNotFound)
	case err != nil:
		return fmt.Errorf("service.RoleService.Update: %w", err)
	}

	if tmpRole.Name == model.RoleAdmin {
		return fmt.Errorf("service.RoleService.Update: %w", ErrAdminRoleCannotBeModified)
	}

	err = r.mySQL.WithContext(ctx).Delete(&model.Role{ID: roleID}).Error
	if err != nil {
		return fmt.Errorf("service.RoleService.Delete: %w", err)
	}

	return nil
}
