package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-rest-api-example/model"
	"github.com/go-rest-api-example/util"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"

	"github.com/openlyinc/pointy"

	"gorm.io/gorm"
)

var (
	// ErrUserNotFound represents user not found error.
	ErrUserNotFound = errors.New("user not found")

	// ErrDuplicateUserEmail means email already taken.
	ErrDuplicateUserEmail = errors.New("duplicate user email")
)

// UserService represents user service.
type UserService struct {
	_     struct{}
	mySQL *gorm.DB
}

// NewUserService returns new user service instance.
func NewUserService(mySQL *gorm.DB) UserService {
	return UserService{
		mySQL: mySQL,
	}
}

// GetByToken finds user by user ID that is attached to jwt token.
func (u UserService) GetByToken(ctx context.Context, tokStr string) (model.User, error) {
	claims, err := util.ParseToken(tokStr)
	if err != nil {
		return model.User{}, fmt.Errorf("service.UserService.GetByToken: %w", err)
	}

	t := claims["type"]
	if t != "access" && t != "refresh" {
		return model.User{}, errors.New("service.UserService.GetByToken: invalid token type")
	}

	var user model.User

	err = u.mySQL.WithContext(ctx).
		Preload("Role.Permissions").
		Joins("Role").
		First(&user, "users.id = ?", claims["id"]).
		Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return model.User{}, fmt.Errorf("service.UserService.GetByToken: %w", ErrUserNotFound)
	case err != nil:
		return model.User{}, fmt.Errorf("service.UserService.GetByToken: %w", err)
	}

	return user, nil
}

// GetByEmail returns user with matched email.
func (u UserService) GetByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User

	err := u.mySQL.WithContext(ctx).
		Joins("Role").
		First(&user, "email = ?", email).
		Error
	if err != nil {
		return model.User{}, fmt.Errorf("service.UserService.GetByEmail: %w", err)
	}

	return user, nil
}

// FindByID returns user with ID matches.
func (u UserService) FindByID(ctx context.Context, userID uint) (model.User, error) {
	var user model.User

	err := u.mySQL.WithContext(ctx).
		Joins("Role").
		First(&user, "`users`.`id` = ?", userID).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, fmt.Errorf("service.UserService.FindAllByTeamID: %w", ErrUserNotFound)
	}

	if err != nil {
		return model.User{}, fmt.Errorf("service.UserService.FindAllByTeamID: %w", err)
	}

	return user, nil
}

// GetUsersByTeamID returns users with matched team (organization) ID.
func (u UserService) GetUsersByTeamID(ctx context.Context, teamID uint, offset, limit int) ([]model.User, int64, error) {
	var count int64

	users := make([]model.User, 0)

	dbWithCtx := u.mySQL.WithContext(ctx)

	err := dbWithCtx.Model(&model.User{}).
		Where(&model.User{OrganizationID: teamID}).
		Count(&count).
		Error
	if err != nil {
		return []model.User{}, 0, fmt.Errorf("service.UserService.GetAllUsersByTeamID: %w", err)
	}

	err = dbWithCtx.Where(&model.User{OrganizationID: teamID}).
		Limit(limit).
		Offset(offset).
		Joins("Role").
		Order("created_at DESC").
		Find(&users).
		Error
	if err != nil {
		return []model.User{}, 0, fmt.Errorf("service.UserService.GetAllUsersByTeamID: %w", err)
	}

	return users, count, nil
}

// Create creates user record with new team and theme
// and returns created user.
func (u UserService) Create(ctx context.Context, teamName string, usr model.User) (model.User, error) {

	tx := u.mySQL.Begin().WithContext(ctx)

	var theme model.Theme
	if err := tx.Create(&theme).Error; err != nil {
		tx.Rollback()
		return model.User{}, fmt.Errorf("service.UserService.Create %w", err)
	}

	team := model.Team{
		Name:        util.GenerateUUIDv4(),
		DisplayName: teamName,
		ThemeID:     pointy.Uint(theme.ID),
		PackageID:   1,
	}
	if err := tx.Create(&team).Error; err != nil {
		tx.Rollback()
		return model.User{}, fmt.Errorf("service.UserService.Create %w", err)
	}

	hashedPwd, err := util.HashPassword(usr.Password)
	if err != nil {
		tx.Rollback()
		return model.User{}, fmt.Errorf("service.UserService.Create %w", err)
	}

	uuidV4, err := uuid.NewRandom()
	if err != nil {
		return model.User{}, fmt.Errorf("service.UserService.Create %w", err)
	}

	var role = model.Role{
		ID:          uuidV4.String(),
		Name:        model.RoleAdmin,
		Description: "Administrator",
		TeamID:      team.ID,
	}

	if err := tx.Create(&role).Error; err != nil {
		tx.Rollback()
		return model.User{}, fmt.Errorf("service.UserService.Create %w", err)
	}

	for _, permission := range model.AllPermissions {

		var rolePermission = model.RolePermission{
			PermissionID: permission.ID,
			RoleID:       role.ID,
		}

		if err := tx.Create(&rolePermission).Error; err != nil {
			tx.Rollback()
			return model.User{}, fmt.Errorf("service.UserService.Create %w", err)
		}

	}

	usr.Email = strings.ToLower(usr.Email)
	usr.Password = hashedPwd
	usr.OrganizationID = team.ID
	usr.CurrentTeamID = team.ID
	usr.RoleID = &role.ID

	if err := tx.Create(&usr).Error; err != nil {
		tx.Rollback()
		return model.User{}, fmt.Errorf("service.UserService.Create %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return model.User{}, fmt.Errorf("service.UserService.Create %w", err)
	}

	return usr, nil
}

// CreateTx creates user record with new team and theme with the given
// transaction but not commit or rollback and returns created user.
func (u UserService) CreateTx(_ context.Context, teamName string, usr model.User, mySQLTx *gorm.DB) (model.User, error) {
	var theme model.Theme
	if err := mySQLTx.Create(&theme).Error; err != nil {
		return model.User{}, fmt.Errorf("service.UserService.CreateTx %w", err)
	}

	team := model.Team{
		Name:        util.GenerateUUIDv4(),
		DisplayName: teamName,
		ThemeID:     pointy.Uint(theme.ID),
		PackageID:   1,
	}
	if err := mySQLTx.Create(&team).Error; err != nil {
		return model.User{}, fmt.Errorf("service.UserService.CreateTx %w", err)
	}

	hashedPwd, err := util.HashPassword(usr.Password)
	if err != nil {
		return model.User{}, fmt.Errorf("service.UserService.CreateTx %w", err)
	}

	usr.Email = strings.ToLower(usr.Email)
	usr.Password = hashedPwd
	usr.OrganizationID = team.ID
	usr.CurrentTeamID = team.ID

	if err := mySQLTx.Create(&usr).Error; err != nil {
		return model.User{}, fmt.Errorf("service.UserService.CreateTx %w", err)
	}

	return usr, nil
}

// GetOrCreateUser returns user with matched email. Create one if not exist.
func (u UserService) GetOrCreateUser(ctx context.Context, email, pwd, name, teamName string) (model.User, error) {
	user := model.User{
		Name:     name,
		Email:    email,
		Password: pwd,
	}

	emailLower := strings.ToLower(email)

	err := u.mySQL.WithContext(ctx).
		Preload("Role.Permissions").
		Joins("Role").
		FirstOrCreate(&user, "email = ?", emailLower).
		Error
	if err != nil {
		return model.User{}, fmt.Errorf("service.UserService.GetOrCreateUser %w", err)
	}

	return user, nil
}

// Update updates user with specified update field(s).
func (u UserService) Update(ctx context.Context, usrID uint, updateFields map[string]interface{}) (model.User, error) {
	// Prevent "id" field in update fields from overwriting user ID.
	delete(updateFields, "id")

	tx := u.mySQL.Begin().WithContext(ctx)

	err := tx.Model(&model.User{ID: usrID}).Updates(updateFields).Error
	if err != nil {
		tx.Rollback()
		return model.User{}, fmt.Errorf("service.UserService.Update: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return model.User{}, fmt.Errorf("service.UserService.Update: %w", err)
	}

	usr, err := u.FindByID(ctx, usrID)
	if err != nil {
		return model.User{}, fmt.Errorf("service.UserService.Update: %w", err)
	}

	return usr, nil
}

// CreateUserToTeam creates user record with existing team ID
// and returns user record ID.
func (u UserService) CreateUserToTeam(ctx context.Context, teamID uint, usr model.User) (model.User, error) {
	hashedPwd, err := util.HashPassword(usr.Password)
	if err != nil {
		return model.User{}, fmt.Errorf("service.UserService.CreateUserToTeam: %w", err)
	}

	// Check role exist.
	if usr.RoleID != nil {
		var count int64

		err = u.mySQL.Model(&model.Role{}).
			Where("team_id = ? AND id = ?", teamID, usr.RoleID).
			Count(&count).
			Error
		if err != nil {
			return model.User{}, fmt.Errorf("service.UserService.CreateUserToTeam: %w", err)
		}

		if count == 0 {
			return model.User{}, fmt.Errorf("service.UserService.CreateUserToTeam: %w", ErrRoleNotFound)
		}
	}

	usr.Email = strings.ToLower(usr.Email)
	usr.Password = hashedPwd
	usr.OrganizationID = teamID
	usr.CurrentTeamID = teamID

	err = u.mySQL.WithContext(ctx).Create(&usr).Error

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return model.User{}, fmt.Errorf("service.UserService.CreateUserToTeam: %w", ErrDuplicateUserEmail)
	}

	if err != nil {
		return model.User{}, fmt.Errorf("service.UserService.CreateUserToTeam: %w", err)
	}

	return usr, nil
}

// DeleteByID deletes user by ID.
func (u UserService) DeleteByID(ctx context.Context, userID uint) error {
	var count int64

	err := u.mySQL.WithContext(ctx).
		Model(&model.User{}).
		Where(&model.User{ID: userID}).
		Count(&count).
		Error

	switch {
	case count == 0:
		return fmt.Errorf("service.UserService.DeleteByID: %w", ErrUserNotFound)
	case err != nil:
		return fmt.Errorf("service.UserService.DeleteByID: %w", err)
	}

	err = u.mySQL.WithContext(ctx).Delete(&model.User{ID: userID}).Error
	if err != nil {
		return fmt.Errorf("service.UserService.DeleteByID: %w", err)
	}

	return nil
}

func (u UserService) HasPermission(ctx context.Context, userID uint, perm model.Permission) (bool, error) {
	var count int64

	err := u.mySQL.WithContext(ctx).
		Preload("Permissions").
		Model(&model.Role{}).
		Joins("JOIN users ON users.role_id = roles.id").
		Joins("JOIN role_permissions rp ON rp.role_id = roles.id").
		Joins("JOIN permissions p ON p.id = rp.permission_id").
		Where("users.id = ? AND p.name = ?",
			userID,
			perm.Name).
		Count(&count).
		Error
	if err != nil {
		return false, fmt.Errorf("service.RoleService.HasPermission: %w", err)
	}

	return count == 1, nil
}
