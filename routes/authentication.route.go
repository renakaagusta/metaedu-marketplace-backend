package routes

import (
	"metaedu-marketplace/controllers"

	"github.com/gin-gonic/gin"
)

type AuthenticationRoutes struct {
	authController controllers.AuthenticationController
}

func NewAuthenticationRoutes(authController controllers.AuthenticationController) AuthenticationRoutes {
	return AuthenticationRoutes{authController}
}

func (rc *AuthenticationRoutes) AuthenticationRoute(rg *gin.RouterGroup) {

	router := rg.Group("/auth")
	router.POST("/sign-up", rc.authController.SignUp)
	router.POST("/sign-in", rc.authController.SignIn)
	router.GET("/users/:address/nonce", rc.authController.GetUserNonce)
}
