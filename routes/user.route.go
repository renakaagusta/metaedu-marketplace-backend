package routes

import (
	"metaedu-marketplace/controllers"
	"metaedu-marketplace/middlewares"

	"github.com/gin-gonic/gin"
)

type UserRoutes struct {
	authorizationMiddleware middlewares.AuthorizationMiddleware
	userController          controllers.UserController
}

func NewUserRoutes(authorizationMiddleware middlewares.AuthorizationMiddleware, userController controllers.UserController) UserRoutes {
	return UserRoutes{authorizationMiddleware, userController}
}

func (rc *UserRoutes) UserRoute(rg *gin.RouterGroup) {

	router := rg.Group("/user")
	router.GET("/me", rc.authorizationMiddleware.VerifyToken, rc.userController.GetMyUserData)
	router.GET("/:id", rc.userController.GetUserData)
	router.PUT("/:id", rc.authorizationMiddleware.VerifyToken, rc.userController.UpdateUser)
}
