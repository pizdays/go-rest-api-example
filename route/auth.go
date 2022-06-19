package route

import (
	"github.com/go-redis/redis"
	"github.com/go-rest-api-example/connection"
	"github.com/go-rest-api-example/handler"
	"github.com/go-rest-api-example/service"

	"github.com/gin-gonic/gin"
)

// addAuthRoutes adds authentication APIs to the given router.
func addAuthRoutes(routerGroup *gin.RouterGroup, redisClient *redis.Client) {
	mySQL := connection.NewMySQL()
	pSQL := connection.NewPostgreSQL()

	userService := service.NewUserService(mySQL)
	authService := service.NewAuthService(mySQL, userService)
	thingService := service.NewThingService(mySQL)
	deviceService := service.NewDeviceService(mySQL)
	cachingMeasurementRepo := repository.NewRedisMeasurementRepository(redisClient)
	measurementService := service.NewMeasurementService(mySQL,
		deviceService,
		cachingMeasurementRepo)
	shiftService := service.NewShiftService(mySQL)
	dashboardService := service.NewDashboardService(mySQL)
	sparePartService := service.NewSparePartService(mySQL, pSQL)
	operationService := service.NewOperationService(mySQL)
	downtimeCategoryService := service.NewDowntimeCategoryService(pSQL)
	widgetService := service.NewWidgetService(mySQL)
	teamService := service.NewTeamService(mySQL)
	messageService := service.NewMessageService(pSQL)
	downtimeTicketService := service.NewDowntimeTicketV2Service(pSQL)
	downtimeSparePartService := service.NewDowntimeSparePartService(pSQL)

	authH := handler.NewAuthHandler(authService,
		userService,
		thingService,
		deviceService,
		measurementService,
		shiftService,
		sparePartService,
		dashboardService,
		operationService,
		downtimeCategoryService,
		widgetService,
		teamService,
		messageService,
		downtimeTicketService,
		downtimeSparePartService)

	authR := routerGroup.Group("/auth")

	authR.POST("/login", authH.LogIn)
	authR.POST("/login/line", authH.LogInLine)

	authR.POST("/refresh-token", authH.RefreshToken)
	authR.DELETE("/logout", authH.LogOut)

	authR.POST("/register", authH.Register)
	authR.POST("/register/line", authH.RegisterByLine)
}
