package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-rest-api-example/model"
	"github.com/go-rest-api-example/service"

	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
)

// UserHandler represents user handler.
type UserHandler struct {
	_           struct{}
	userService service.UserService
}

// NewUserHandler returns new user handler instance.
func NewUserHandler(userSvc service.UserService) UserHandler {
	return UserHandler{userService: userSvc}
}

// FindProfile responds user profile.
//
// @Summary   Find profile
// @Tags      User
// @Security  BearerToken
// @Produce   json
// @Success   200  {object}  model.ClientUserResponse
// @Failure   401  {object}  model.ErrorResponse
// @Failure   500  {object}  model.ErrorResponse
// @Router    /users/info [GET]
func (u UserHandler) FindProfile(c *gin.Context) {
	team := c.MustGet("team").(model.Team)
	user := c.MustGet("user").(model.User)

	cu := model.ClientUser{
		ID:               user.ID,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
		Name:             user.Name,
		OrganizationID:   team.ID,
		OrganizationName: team.DisplayName,
		Email:            user.Email,
		EmailVerifiedAt:  user.EmailVerifiedAt,
		PhoneNumber:      user.PhoneNumber,
		Lang:             user.Lang,
		CurrentTeamID:    user.CurrentTeamID,
		Channel:          team.Name,
		Topic:            fmt.Sprintf("organization/%s/messages", team.Name),
		PackageID:        team.PackageID,
		Role:             user.Role,
	}

	c.JSON(http.StatusOK, model.ClientUserResponse{
		Message: "success",
		Result:  cu,
	})
}

// Update updates user.
//
// @Summary      Update user
// @Tags         User
// @Security     BearerToken
// @Description  🔒 Require `PermissionUserUpdate` permission.
// @Description  Only specify attributes that need to be updated.
// @Description  User attributes that can be updated: `name`, `email`, `phoneNumber`, `lang`, and `lineId`.
// @Accept       json
// @Produce      json
// @Param        reqBody  body      map[string]interface{}  true  "Update request body"
// @Success      200      {object}  model.UserResponse
// @Failure      400      {object}  model.ErrorResponse
// @Failure      401      {object}  model.ErrorResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /users [PATCH]
func (u UserHandler) Update(c *gin.Context) {
	var reqBody map[string]interface{}

	if err := c.BindJSON(&reqBody); err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusBadRequest, model.ErrRespFailedBodyRead)

		return
	}

	user := c.MustGet("user").(model.User)
	updateMap := make(map[string]interface{})

	if name, ok := reqBody["name"].(string); ok {
		updateMap["name"] = name
	}

	if email, ok := reqBody["email"].(string); ok {
		updateMap["email"] = email
	}

	if phoneNumber, ok := reqBody["phoneNumber"].(string); ok {
		updateMap["phone_number"] = phoneNumber
	}

	if lang, ok := reqBody["lang"].(string); ok {
		updateMap["lang"] = lang
	}

	if lineID, ok := reqBody["lineId"].(string); ok {
		updateMap["lineId"] = lineID
	}

	ctx := c.MustGet("context").(context.Context)

	updatedUsr, err := u.userService.Update(ctx, user.ID, updateMap)
	if err != nil {
		_ = c.Error(err)

		if errors.Is(err, service.ErrRoleNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{
				Message: "Role not found",
			})

			return
		}

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Message: "Update user failed",
		})

		return
	}

	c.JSON(http.StatusOK, model.UserResponse{
		Data: updatedUsr,
	})
}

// FindUsersByTeamID responds users that team ID matches with the given team ID.
//
// @Summary      Find users
// @Tags         User
// @Security     BearerToken
// @Description  🔒 Require `PermissionUserRead` permission.
// @Description  Find users of team with offset and limit applied.
// @Produce      json
// @Param        teamId  path      uint  true   "Team ID"
// @Param        offset  query     uint  false  "Data offset"  default(0)
// @Param        limit   query     int   false  "Data limit"   default(10)
// @Success      200     {object}  object{data=[]model.User,meta=model.PagingMeta}
// @Failure      400     {object}  model.ErrorResponse
// @Failure      404     {object}  model.ErrorResponse
// @Failure      500     {object}  model.ErrorResponse
// @Router       /teams/:teamId/users [GET]
func (u UserHandler) FindUsersByTeamID(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("teamID"), 10, 64)
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	q := model.PagingQuery{Limit: 10}
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	ctx := c.MustGet("context").(context.Context)

	users, count, err := u.userService.GetUsersByTeamID(ctx,
		uint(teamID),
		q.Offset,
		q.Limit)
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"meta": model.PagingMeta{
			Offset:     q.Offset,
			Limit:      q.Limit,
			TotalCount: count,
		},
	})
}

// CreateUserToTeam creates user to existing team.
//
// @Summary      Create user to team
// @Tags         User
// @Security     BearerToken
// @Description  🔒 Require `PermissionUserCreate` permission.
// @Accept       json
// @Produce      json
// @Param        teamId   path      uint                           true  "Team ID"
// @Param        reqBody  body      model.CreateUserToTeamReqBody  true  "Request body"
// @Success      201      {object}  object{data=model.User}
// @Failure      400      {object}  model.ErrorResponse
// @Failure      401      {object}  model.ErrorResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /teams/:teamId/users [POST]
func (u UserHandler) CreateUserToTeam(c *gin.Context) {
	var reqBody model.CreateUserToTeamReqBody
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
	usr := model.User{
		Name:   reqBody.Name,
		Email:  reqBody.Email,
		RoleID: &reqBody.RoleID,
	}

	createdUser, err := u.userService.CreateUserToTeam(ctx, team.ID, usr)

	switch {
	case errors.Is(err, service.ErrDuplicateUserEmail):
		_ = c.Error(err)

		c.JSON(http.StatusConflict, model.ErrorResponse{Message: "Email already taken"})

		return
	case errors.Is(err, service.ErrRoleNotFound):
		_ = c.Error(err)

		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Role not found"})

		return
	case err != nil:
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Create user failed"})

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": createdUser,
	})
}

// DeleteByID deletes user by ID.
//
// @Summary      Delete user
// @Tags         User
// @Security     BearerToken
// @Description  🔒 Require `PermissionUserDelete` permission.
// @Param        teamId  path  uint  true  "Team ID"
// @Param        userId  path  uint  true  "User ID"
// @Success      204
// @Failure      400  {object}  model.ErrorResponse
// @Failure      401  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /teams/:teamId/users/:userId [DELETE]
func (u UserHandler) DeleteByID(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userID"), 10, 64)
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "Invalid user ID"})

		return
	}

	ctx := c.MustGet("context").(context.Context)

	err = u.userService.DeleteByID(ctx, uint(userID))

	switch {
	case errors.Is(err, service.ErrUserNotFound):
		_ = c.Error(err)

		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "User not found"})

		return
	case err != nil:
		_ = c.Error(err)

		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "Delete user failed"})

		return
	}

	c.Status(http.StatusNoContent)
}
