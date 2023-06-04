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

	CollectionRepository    *repositories.CollectionRepository
	FractionRepository      *repositories.FractionRepository
	OwnershipRepository     *repositories.OwnershipRepository
	RentalRepository        *repositories.RentalRepository
	TokenRepository         *repositories.TokenRepository
	TokenCategoryRepository *repositories.TokenCategoryRepository
	TransactionRepository   *repositories.TransactionRepository
	UserRepository          *repositories.UserRepository

	AuthorizationMiddleware *middlewares.AuthorizationMiddleware

	AuthenticationController controllers.AuthenticationController
	CollectionController     controllers.CollectionController
	FractionController       controllers.FractionController
	OwnershipController      controllers.OwnershipController
	RentalController         controllers.RentalController
	TokenController          controllers.TokenController
	TokenCategoryController  controllers.TokenCategoryController
	TransactionController    controllers.TransactionController
	UserController           controllers.UserController

	AuthenticationRoutes routes.AuthenticationRoutes
	CollectionRoutes     routes.CollectionRoutes
	FractionRoutes       routes.FractionRoutes
	OwnershipRoutes      routes.OwnershipRoutes
	RentalRoutes         routes.RentalRoutes
	TokenRoutes          routes.TokenRoutes
	TokenCategoryRoutes  routes.TokenCategoryRoutes
	TransactionRoutes    routes.TransactionRoutes
	UserRoutes           routes.UserRoutes
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

	CollectionRepository = repositories.NewCollectionRepository(dbClient)
	FractionRepository = repositories.NewFractionRepository(dbClient)
	OwnershipRepository = repositories.NewOwnershipRepository(dbClient)
	RentalRepository = repositories.NewRentalRepository(dbClient)
	TokenRepository = repositories.NewTokenRepository(dbClient)
	TokenCategoryRepository = repositories.NewTokenCategoryRepository(dbClient)
	TransactionRepository = repositories.NewTransactionRepository(dbClient)
	UserRepository = repositories.NewUserRepository(dbClient)

	AuthorizationMiddleware = middlewares.NewAuthorizationMiddleware(*UserRepository, jwtHmacProvider)

	AuthenticationController = *controllers.NewAuthenticationController(UserRepository, jwtHmacProvider)
	CollectionController = *controllers.NewCollectionController(CollectionRepository, TransactionRepository, web3StorageClient, redisClient)
	FractionController = *controllers.NewFractionController(FractionRepository, TokenRepository, OwnershipRepository, RentalRepository, UserRepository, web3StorageClient, redisClient)
	OwnershipController = *controllers.NewOwnershipController(OwnershipRepository, TokenRepository, RentalRepository, web3StorageClient, redisClient)
	RentalController = *controllers.NewRentalController(RentalRepository, web3StorageClient, redisClient)
	TokenController = *controllers.NewTokenController(TokenRepository, OwnershipRepository, CollectionRepository, TransactionRepository, web3StorageClient, redisClient)
	TokenCategoryController = *controllers.NewTokenCategoryController(TokenCategoryRepository, web3StorageClient, redisClient)
	TransactionController = *controllers.NewTransactionController(TransactionRepository, TokenRepository, CollectionRepository, OwnershipRepository, RentalRepository, web3StorageClient, redisClient)
	UserController = *controllers.NewUserController(UserRepository, web3StorageClient, redisClient)

	AuthenticationRoutes = routes.NewAuthenticationRoutes(AuthenticationController)
	CollectionRoutes = routes.NewCollectionRoutes(*AuthorizationMiddleware, CollectionController)
	FractionRoutes = routes.NewFractionRoutes(*AuthorizationMiddleware, FractionController)
	OwnershipRoutes = routes.NewOwnershipRoutes(*AuthorizationMiddleware, OwnershipController)
	RentalRoutes = routes.NewRentalRoutes(*AuthorizationMiddleware, RentalController)
	TokenRoutes = routes.NewTokenRoutes(*AuthorizationMiddleware, TokenController)
	TokenCategoryRoutes = routes.NewTokenCategoryRoutes(*AuthorizationMiddleware, TokenCategoryController)
	TransactionRoutes = routes.NewTransactionRoutes(*AuthorizationMiddleware, TransactionController)
	UserRoutes = routes.NewUserRoutes(*AuthorizationMiddleware, UserController)

	server = gin.Default()
}

func main() {
	// Set maximum uploaded file size
	server.MaxMultipartMemory = 10 << 20

	// Add gin logger
	server.Use(gin.Logger())

	// Add gin recovery
	server.Use(gin.Recovery())

	router := server.Group("/api/v1")

	router.GET("/healthchecker", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Welcome to MetaEdu Marketplace"})
	})

	AuthenticationRoutes.AuthenticationRoute(router)
	CollectionRoutes.CollectionRoute(router)
	FractionRoutes.FractionRoute(router)
	OwnershipRoutes.OwnershipRoute(router)
	RentalRoutes.RentalRoute(router)
	TokenRoutes.TokenRoute(router)
	TokenCategoryRoutes.TokenCategoryRoute(router)
	TransactionRoutes.TransactionRoute(router)
	UserRoutes.UserRoute(router)

	port, success := os.LookupEnv("PORT")
	if !success {
		fmt.Fprintln(os.Stderr, "No REDIS_PORT - set the REDIS_PORT environment var and try again.")
		os.Exit(1)
	}

	log.Fatal(server.Run(":" + port))
}
