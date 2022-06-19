package route

import (
	"github.com/go-rest-api-example/connection"
	"github.com/go-rest-api-example/handler"
	"github.com/go-rest-api-example/middleware"
	"github.com/go-rest-api-example/model"
	"github.com/go-rest-api-example/service"

	"github.com/gin-gonic/gin"
)

// AddRoleEndpoints adds role endpoints to routerGroup.
func AddRoleEndpoints(routerGroup *gin.RouterGroup) {
	mySQL := connection.NewMySQL()

	roleService := service.NewRoleService(mySQL)
	userService := service.NewUserService(mySQL)
	teamService := service.NewTeamService(mySQL)

	roleHandler := handler.NewRoleHandler(roleService)

	teamPathPermMW := middleware.NewTeamIDPathPermMw("teamID")
	userPermMW := middleware.NewUserPermissionMiddleware(userService)

	roleRouter := routerGroup.Group("/teams/:teamID/roles").
		Use(middleware.AuthRequired(userService, teamService)).
		Use(teamPathPermMW)

	roleRouter.GET("", userPermMW.HasPermission(model.PermUserRead), roleHandler.FindAll)
	roleRouter.POST("", userPermMW.HasPermission(model.PermUserCreate), roleHandler.Create)
	roleRouter.PUT("/:roleID", userPermMW.HasPermission(model.PermUserUpdate), roleHandler.Update)
	roleRouter.DELETE("/:roleID",
		userPermMW.HasPermission(model.PermUserDelete),
		roleHandler.Delete)
}
