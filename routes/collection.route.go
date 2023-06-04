package routes

import (
	"metaedu-marketplace/controllers"
	"metaedu-marketplace/middlewares"

	"github.com/gin-gonic/gin"
)

type CollectionRoutes struct {
	authorizationMiddleware middlewares.AuthorizationMiddleware
	collectionController    controllers.CollectionController
}

func NewCollectionRoutes(authorizationMiddleware middlewares.AuthorizationMiddleware, collectionController controllers.CollectionController) CollectionRoutes {
	return CollectionRoutes{authorizationMiddleware, collectionController}
}

func (rc *CollectionRoutes) CollectionRoute(rg *gin.RouterGroup) {

	router := rg.Group("/collection")

	router.GET("/:id/transaction", rc.collectionController.GetCollectionTransactionList)
	router.GET("/:id", rc.collectionController.GetCollectionData)
	router.PUT("/:id", rc.authorizationMiddleware.VerifyToken, rc.collectionController.UpdateCollection)
	router.DELETE("/:id", rc.authorizationMiddleware.VerifyToken, rc.collectionController.DeleteCollection)
	router.POST("/", rc.authorizationMiddleware.VerifyToken, rc.collectionController.InsertCollection)
	router.GET("/", rc.collectionController.GetCollectionList)
}
