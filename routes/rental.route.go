package routes

import (
	"metaedu-marketplace/controllers"
	"metaedu-marketplace/middlewares"

	"github.com/gin-gonic/gin"
)

type RentalRoutes struct {
	authorizationMiddleware middlewares.AuthorizationMiddleware
	rentalController        controllers.RentalController
}

func NewRentalRoutes(authorizationMiddleware middlewares.AuthorizationMiddleware, rentalController controllers.RentalController) RentalRoutes {
	return RentalRoutes{authorizationMiddleware, rentalController}
}

func (rc *RentalRoutes) RentalRoute(rg *gin.RouterGroup) {

	router := rg.Group("/rental")
	router.GET("/:id", rc.rentalController.GetRentalData)
	router.PUT("/:id", rc.authorizationMiddleware.VerifyToken, rc.rentalController.UpdateRental)
	router.DELETE("/:id", rc.authorizationMiddleware.VerifyToken, rc.rentalController.DeleteRental)
	router.POST("/", rc.authorizationMiddleware.VerifyToken, rc.rentalController.InsertRental)
	router.GET("/", rc.rentalController.GetRentalList)
}
