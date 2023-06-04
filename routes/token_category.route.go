package routes

import (
	"metaedu-marketplace/controllers"
	"metaedu-marketplace/middlewares"

	"github.com/gin-gonic/gin"
)

type TokenCategoryRoutes struct {
	authorizationMiddleware middlewares.AuthorizationMiddleware
	tokenCategoryController controllers.TokenCategoryController
}

func NewTokenCategoryRoutes(authorizationMiddleware middlewares.AuthorizationMiddleware, tokenCategoryController controllers.TokenCategoryController) TokenCategoryRoutes {
	return TokenCategoryRoutes{authorizationMiddleware, tokenCategoryController}
}

func (rc *TokenCategoryRoutes) TokenCategoryRoute(rg *gin.RouterGroup) {

	router := rg.Group("/token-category")
	router.GET("/:id", rc.tokenCategoryController.GetTokenCategoryData)
	router.PUT("/:id", rc.authorizationMiddleware.VerifyToken, rc.tokenCategoryController.UpdateTokenCategory)
	router.DELETE("/:id", rc.authorizationMiddleware.VerifyToken, rc.tokenCategoryController.DeleteTokenCategory)
	router.POST("/", rc.authorizationMiddleware.VerifyToken, rc.tokenCategoryController.InsertTokenCategory)
	router.GET("/", rc.tokenCategoryController.GetTokenCategoryList)
}
