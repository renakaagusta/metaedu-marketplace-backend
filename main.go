package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"metaedu-marketplace/config"
	"metaedu-marketplace/controllers"
	"metaedu-marketplace/middlewares"
	"metaedu-marketplace/repositories"
	"metaedu-marketplace/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

var (
	server *gin.Engine

	UserRepository  *repositories.UserRepository
	TokenRepository *repositories.TokenRepository

	AuthorizationMiddleware *middlewares.AuthorizationMiddleware

	AuthenticationController controllers.AuthenticationController
	TokenController          controllers.TokenController

	AuthenticationRoutes routes.AuthenticationRoutes
	TokenRoutes          routes.TokenRoutes
)

func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
		os.Exit(1)
	}

	dbClient := config.CreateDBClient()
	jwtHmacProvider := config.CreateJwtHmacProvider()
	web3StorageClient := config.CreateWeb3StorageClient()
	redisClient := config.CreateRedisClient()

	UserRepository = repositories.NewUserRepository(dbClient)
	TokenRepository = repositories.NewTokenRepository(dbClient)

	AuthorizationMiddleware = middlewares.NewAuthorizationMiddleware(jwtHmacProvider)

	AuthenticationController = *controllers.NewAuthenticationController(UserRepository, jwtHmacProvider)
	TokenController = *controllers.NewTokenController(TokenRepository, web3StorageClient, redisClient)

	AuthenticationRoutes = routes.NewAuthenticationRoutes(AuthenticationController)
	TokenRoutes = routes.NewTokenRoutes(*AuthorizationMiddleware, TokenController)

	server = gin.Default()
}

func main() {
	server.MaxMultipartMemory = 10 << 20

	server.Use(gin.Logger())

	server.Use(gin.Recovery())

	router := server.Group("/api/v1")

	router.GET("/healthchecker", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Welcome to MetaEdu Marketplace"})
	})

	AuthenticationRoutes.AuthenticationRoute(router)
	TokenRoutes.TokenRoute(router)

	port, success := os.LookupEnv("PORT")
	if !success {
		fmt.Fprintln(os.Stderr, "No REDIS_PORT - set the REDIS_PORT environment var and try again.")
		os.Exit(1)
	}

	log.Fatal(server.Run(":" + port))
}
