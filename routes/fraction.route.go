package routes

import (
	"metaedu-marketplace/controllers"
	"metaedu-marketplace/middlewares"

	"github.com/gin-gonic/gin"
)

type FractionRoutes struct {
	authorizationMiddleware middlewares.AuthorizationMiddleware
	fractionController      controllers.FractionController
}

func NewFractionRoutes(authorizationMiddleware middlewares.AuthorizationMiddleware, fractionController controllers.FractionController) FractionRoutes {
	return FractionRoutes{authorizationMiddleware, fractionController}
}

func (rc *FractionRoutes) FractionRoute(rg *gin.RouterGroup) {

	router := rg.Group("/fraction")
	router.GET("/:id", rc.fractionController.GetFractionData)
	router.PUT("/:id", rc.authorizationMiddleware.VerifyToken, rc.fractionController.UpdateFraction)
	router.DELETE("/:id", rc.authorizationMiddleware.VerifyToken, rc.fractionController.DeleteFraction)
	router.POST("/", rc.authorizationMiddleware.VerifyToken, rc.fractionController.InsertFraction)
	router.GET("/", rc.fractionController.GetFractionList)
}
