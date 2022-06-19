package route

import (
	"crypto/tls"
	"net/http"
	"os"

	"github.com/go-rest-api-example/connection"
	"github.com/go-rest-api-example/handler"
	"github.com/go-rest-api-example/service"

	"github.com/gin-gonic/gin"
	"github.com/mailgun/mailgun-go/v4"
)

// addPasswordRoutes adds password reset APIs to the given router.
func addPasswordRoutes(routerGroup *gin.RouterGroup) {
	// Create an instance of the Mailgun Client
	transCfg := &http.Transport{
		//nolint:gosec // Ask the project owner why this line exist.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	httpClient := &http.Client{Transport: transCfg}
	mg := mailgun.NewMailgun(os.Getenv("MAILGUN_DOMAIN"), os.Getenv("MAILGUN_API_KEY"))
	mg.SetClient(httpClient)

	mySQL := connection.NewMySQL()

	userService := service.NewUserService(mySQL)

	passwordHandler := handler.NewPasswordHandler(service.NewPasswordService(mySQL, userService),
		service.NewEmailService(mg))

	passwordRoute := routerGroup.Group("/password")
	passwordRoute.POST("/password-reset", passwordHandler.CreatePasswordReset)
	passwordRoute.GET("/validate-password-reset",
		passwordHandler.ValidatePasswordReset)
	passwordRoute.GET("/check-password-reset-expire",
		passwordHandler.CheckPasswordResetExpire)
	passwordRoute.PATCH("/password", passwordHandler.ResetPassword)
}

