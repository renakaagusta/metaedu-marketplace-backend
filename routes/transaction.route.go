package routes

import (
	"metaedu-marketplace/controllers"
	"metaedu-marketplace/middlewares"

	"github.com/gin-gonic/gin"
)

type TransactionRoutes struct {
	authorizationMiddleware middlewares.AuthorizationMiddleware
	transactionController   controllers.TransactionController
}

func NewTransactionRoutes(authorizationMiddleware middlewares.AuthorizationMiddleware, transactionController controllers.TransactionController) TransactionRoutes {
	return TransactionRoutes{authorizationMiddleware, transactionController}
}

func (rc *TransactionRoutes) TransactionRoute(rg *gin.RouterGroup) {

	router := rg.Group("/transaction")
	router.GET("/:id", rc.transactionController.GetTransactionData)
	router.PUT("/:id", rc.authorizationMiddleware.VerifyToken, rc.transactionController.UpdateTransaction)
	router.DELETE("/:id", rc.authorizationMiddleware.VerifyToken, rc.transactionController.DeleteTransaction)
	router.POST("/", rc.authorizationMiddleware.VerifyToken, rc.transactionController.InsertTransaction)
	router.GET("/", rc.transactionController.GetTransactionList)
}
