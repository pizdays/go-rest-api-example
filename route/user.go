package route

import (
	"github.com/go-rest-api-example/connection"
	"github.com/go-rest-api-example/handler"
	"github.com/go-rest-api-example/middleware"
	"github.com/go-rest-api-example/model"
	"github.com/go-rest-api-example/service"

	"github.com/gin-gonic/gin"
)

// AddUserRoutes add user APIs to the given router.
func AddUserRoutes(routerGroup *gin.RouterGroup) {
	mySQL := connection.NewMySQL()

	userService := service.NewUserService(mySQL)
	teamService := service.NewTeamService(mySQL)

	userH := handler.NewUserHandler(userService)

	teamIDPathPermMw := middleware.NewTeamIDPathPermMw("teamID")
	userPermMw := middleware.NewUserPermissionMiddleware(userService)

	userRouter := routerGroup.Group("")
	userRouter.Use(middleware.AuthRequired(userService, teamService))

	userRouter.GET("/users/info",
		userH.FindProfile)
	userRouter.GET("/teams/:teamID/users",
		userPermMw.HasPermission(model.PermUserRead),
		teamIDPathPermMw,
		userH.FindUsersByTeamID)
	userRouter.POST("/teams/:teamID/users",
		userPermMw.HasPermission(model.PermUserCreate),
		teamIDPathPermMw,
		userH.CreateUserToTeam)
	userRouter.PATCH("/users",
		userPermMw.HasPermission(model.PermUserUpdate),
		userH.Update)
	userRouter.DELETE("/teams/:teamID/users/:userID",
		userPermMw.HasPermission(model.PermUserDelete),
		teamIDPathPermMw,
		userH.DeleteByID)
}

