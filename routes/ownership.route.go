package routes

import (
	"metaedu-marketplace/controllers"
	"metaedu-marketplace/middlewares"

	"github.com/gin-gonic/gin"
)

type OwnershipRoutes struct {
	authorizationMiddleware middlewares.AuthorizationMiddleware
	ownershipController     controllers.OwnershipController
}

func NewOwnershipRoutes(authorizationMiddleware middlewares.AuthorizationMiddleware, ownershipController controllers.OwnershipController) OwnershipRoutes {
	return OwnershipRoutes{authorizationMiddleware, ownershipController}
}

func (rc *OwnershipRoutes) OwnershipRoute(rg *gin.RouterGroup) {

	router := rg.Group("/ownership")
	router.GET("/:id", rc.ownershipController.GetOwnershipData)
	router.PUT("/:id", rc.authorizationMiddleware.VerifyToken, rc.ownershipController.UpdateOwnership)
	router.DELETE("/:id", rc.authorizationMiddleware.VerifyToken, rc.ownershipController.DeleteOwnership)
	router.POST("/", rc.authorizationMiddleware.VerifyToken, rc.ownershipController.InsertOwnership)
	router.GET("/", rc.ownershipController.GetOwnershipList)
}
