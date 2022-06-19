package route

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/go-rest-api-example/middleware"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func NewApp() *gin.Engine {

	r := gin.New()

	r.Use(func(c *gin.Context) {
		defer func() {
			if recov := recover(); recov != nil {
				err := fmt.Errorf("panic: %s\nstacktrace: %s",
					recov,
					debug.Stack())
				_ = c.Error(err)

				c.JSON(http.StatusInternalServerError,
					model.ErrRespUnhandledPanic)
			}
		}()
		c.Next()
	})
	r.Use(middleware.CORS())

	// Serve Swagger UI on non-production environment.
	if os.Getenv("MODE") != "prod" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	v1 := r.Group("/v1")

	addAuthRoutes(v1)
	addPasswordRoutes(v1)
	AddUserRoutes(v1)
	AddRoleEndpoints(v1)

	return r
}
