package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-rest-api-example/model"
	"github.com/go-rest-api-example/service"

	"github.com/dgrijalva/jwt-go"

	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
)

// AuthRequired returns authentication middleware that also sets user and team
// in Gin context for further use.
func AuthRequired(userService service.UserService, teamService service.TeamService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) <= len("Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
				Message: "Authentication failed",
			})

			return
		}

		tokenString := authHeader[7:]
		ctx := c.MustGet("context").(context.Context)

		user, err := userService.GetByToken(ctx, tokenString)

		var jwtValidationErr *jwt.ValidationError

		switch {
		case errors.As(err, &jwtValidationErr):
			_ = c.Error(err)

			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
				Message: "Authentication failed",
			})

			return
		case errors.Is(err, service.ErrUserNotFound):
			_ = c.Error(err)

			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
				Message: "Authentication failed",
			})

			return
		case err != nil:
			_ = c.Error(err)

			c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
				Message: "Authentication failed",
			})

			return
		}

		team, err := teamService.FindByID(ctx, user.OrganizationID)
		if err != nil {
			_ = c.Error(err)

			c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
				Message: "Find user failed",
			})

			return
		}

		c.Set("user", user)

		withUsrCtx := context.WithValue(ctx, model.CtxUser, user)
		c.Set("context", withUsrCtx)

		log.Debug().
			Interface("user", user).
			Msg("Set context values")
	}
}
