package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-rest-api-example/model"
	"github.com/go-rest-api-example/service"

	"github.com/google/uuid"

	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
)

// RoleHandler represents role handler.
type RoleHandler struct {
	roleService service.RoleService
}

// NewRoleHandler returns new role handler instance.
func NewRoleHandler(roleService service.RoleService) RoleHandler {
	return RoleHandler{roleService: roleService}
}

// FindAll finds all roles of team with offset and limit applied.
// @Summary      Find roles
// @Tags         Role
// @Security     BearerToken
// @Description  🔒 Require `PermissionUserRead` permission.
// @Param        offset  query     uint  false  "Data offset"  default(0)
// @Param        limit   query     uint  false  "Data limit"   default(10)
// @Param        teamId  path      uint  true   "Team ID"
// @Success      200     {object}  object{data=[]model.Role,meta=model.PagingMeta}
// @Failure      400     {object}  model.ErrorResponse
// @Failure      401     {object}  model.ErrorResponse
// @Failure      500     {object}  model.ErrorResponse
// @Router       /teams/:teamId/roles [GET]
func (r RoleHandler) FindAll(c *gin.Context) {
	ctx := c.MustGet("context").(context.Context)

	teamID, err := strconv.ParseUint(c.Param("teamID"), 10, 64)
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "Invalid Team ID"})

		return
	}

	q := model.PagingQuery{Limit: 10}
	if err := c.BindQuery(&q); err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusBadRequest, model.ErrRespFailedQueryRead)

		return
	}

	roleCount, err := r.roleService.CountByTeamID(ctx, uint(teamID))
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Find roles failed"})

		return
	}

	roles, err := r.roleService.FindAllByTeamID(ctx, uint(teamID), q)
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Find roles failed"})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": roles,
		"meta": model.PagingMeta{
			Offset:     q.Offset,
			Limit:      q.Limit,
			TotalCount: roleCount,
		},
	})
}

// Create creates role for team.
//
// @Summary      Create role
// @Tags         Role
// @Security     BearerToken
// @Description  🔒 Require `PermissionUserCreate` permission.
// @Description  Cannot create role with name `Admin`.
// @Accept       json
// @Produce      json
// @Param        teamId   path      uint               true  "Team ID"
// @Param        reqBody  body      model.RoleReqBody  true  "Request body"
// @Success      201      {object}  object{data=model.Role}
// @Failure      400      {object}  model.ErrorResponse
// @Failure      401      {object}  model.ErrorResponse
// @Failure      404      {object}  model.ErrorResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /teams/:teamId/roles [POST]
func (r RoleHandler) Create(c *gin.Context) {
	var reqBody model.RoleReqBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		_ = c.Error(err)

		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			c.JSON(http.StatusBadRequest, model.ErrRespFailedBodyValidation)
			return
		}

		c.JSON(http.StatusBadRequest, model.ErrRespFailedBodyRead)

		return
	}

	ctx := c.MustGet("context").(context.Context)
	team := c.MustGet("team").(model.Team)

	perms := make([]model.Permission, len(reqBody.PermissionIDs))

	for i, permID := range reqBody.PermissionIDs {
		perms[i] = model.Permission{ID: permID}
	}

	role := model.Role{
		Name:        reqBody.Name,
		Description: reqBody.Description,
		Permissions: perms,
		TeamID:      team.ID,
	}

	createdRole, err := r.roleService.Create(ctx, role)

	switch {
	case errors.Is(err, service.ErrPermissionNotFound):
		_ = c.Error(err)

		c.JSON(http.StatusNotFound,
			model.ErrorResponse{Message: "Permission(s) not found"})

		return
	case err != nil:
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError,
			model.ErrorResponse{Message: "Create role failed"})

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": createdRole,
	})
}

// Update updates role.
//
// @Summary      Update role
// @Tags         Role
// @Security     BearerToken
// @Description  🔒 Require `PermissionUserUpdate` permission.
// @Description  Cannot update `Admin` role.
// @Accept       json
// @Produce      json
// @Param        teamId   path      uint               true  "Team ID"
// @Param        roleId   path      string             true  "Role ID"
// @Param        reqBody  body      model.RoleReqBody  true  "Request body"
// @Success      200      {object}  object{data=model.Role}
// @Failure      400      {object}  model.ErrorResponse
// @Failure      401      {object}  model.ErrorResponse
// @Failure      404      {object}  model.ErrorResponse
// @Failure      409      {object}  model.ErrorResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /teams/:teamId/roles/:roleId [PUT]
func (r RoleHandler) Update(c *gin.Context) {
	var reqBody model.RoleReqBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		_ = c.Error(err)

		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			c.JSON(http.StatusBadRequest, model.ErrRespFailedBodyValidation)
			return
		}

		c.JSON(http.StatusBadRequest, model.ErrRespFailedBodyRead)

		return
	}

	roleID, err := uuid.Parse(c.Param("roleID"))
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusBadRequest,
			model.ErrorResponse{Message: "Invalid role ID"})

		return
	}

	ctx := c.MustGet("context").(context.Context)
	team := c.MustGet("team").(model.Team)

	perms := make([]model.Permission, len(reqBody.PermissionIDs))

	for i, permID := range reqBody.PermissionIDs {
		perms[i] = model.Permission{ID: permID}
	}

	role := model.Role{
		ID:          roleID.String(),
		Name:        reqBody.Name,
		Description: reqBody.Description,
		TeamID:      team.ID,
		Permissions: perms,
	}

	updatedRole, err := r.roleService.Update(ctx, role)

	switch {
	case errors.Is(err, service.ErrAdminRoleCannotBeModified):
		_ = c.Error(err)

		c.JSON(http.StatusConflict, model.ErrorResponse{Message: "Admin role can't be updated"})

		return
	case errors.Is(err, service.ErrRoleNotFound):
		_ = c.Error(err)

		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Role not found"})

		return
	case errors.Is(err, service.ErrPermissionNotFound):
		_ = c.Error(err)

		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Permission(s) not found"})

		return
	case err != nil:
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Update role failed"})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": updatedRole,
	})
}

// Delete deletes role.
//
// @Summary      Delete role
// @Tags         Role
// @Security     BearerToken
// @Description  🔒 Require `PermissionUserDelete` permission.
// @Description  Cannot delete `Admin` role.
// @Param        teamId  path  uint    true  "Team ID"
// @Param        roleId  path  string  true  "Role ID"
// @Success      204
// @Failure      400  {object}  model.ErrorResponse
// @Failure      401  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /teams/:teamId/roles/:roleId [DELETE]
func (r RoleHandler) Delete(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("roleID"))
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "Invalid role ID"})

		return
	}

	ctx := c.MustGet("context").(context.Context)

	err = r.roleService.DeleteByID(ctx, roleID.String())

	switch {
	case errors.Is(err, service.ErrRoleNotFound):
		_ = c.Error(err)

		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Role not found"})

		return
	case errors.Is(err, service.ErrAdminRoleCannotBeModified):
		_ = c.Error(err)

		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "Admin role can't be deleted"})

		return
	case err != nil:
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Delete role failed"})

		return
	}

	c.Status(http.StatusNoContent)
}
