package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"metaedu-marketplace/config"
	"metaedu-marketplace/helpers"
	"metaedu-marketplace/repositories"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-co-op/gocron"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
)

var (
	ctx         = context.Background()
	rpcUrl      string
	ethClient   *ethclient.Client
	dbClient    *sql.DB
	redisClient *redis.Client

	collectionRepository    *repositories.CollectionRepository
	fractionRepository      *repositories.FractionRepository
	ownershipRepository     *repositories.OwnershipRepository
	rentalRepository        *repositories.RentalRepository
	tokenRepository         *repositories.TokenRepository
	tokenCategoryRepository *repositories.TokenCategoryRepository
	transactionRepository   *repositories.TransactionRepository
	userRepository          *repositories.UserRepository
)

func checkBlock() {
	status := "waiting_confirmation"
	monitoringOnly := false

	var empty string = ""

	// Get pending tokens
	tokens, err := tokenRepository.GetTokenList(0, 100000000, "", nil, nil, nil, &empty, 0, 10000000000, &status, "created_at", "ASC")

	if err != nil {
		fmt.Println("Error getting token list: ", err)
		return
	}

	// Check all pending tokens
	for _, token := range tokens {
		// Check if is in the monitoring only mode
		if monitoringOnly {
			break
		}

		transactionId := common.HexToHash(token.TransactionHash)

		// Get transaction receipt
		receipt, err := ethClient.TransactionReceipt(ctx, transactionId)

		if err != nil {
			log.Print(err)
			continue
		}

		if receipt.Status == 1 {
			if token.PreviousID != helpers.GetEmptyUUID() {

				oldToken, err := tokenRepository.GetTokenData(token.PreviousID)

				if err != nil {
					fmt.Println("Failed to get, token: ", token.PreviousID, ", from: ", token.ID, ", error: ", err)
				}

				oldToken.LastPrice = token.LastPrice
				oldToken.NumberOfTransactions = token.NumberOfTransactions
				oldToken.VolumeTransactions = token.VolumeTransactions
				oldToken.FractionID = token.FractionID
				oldToken.SourceID = token.SourceID
				oldToken.Views = token.Views

				err = tokenRepository.UpdateToken(oldToken.ID, oldToken)

				if err != nil {
					fmt.Println("Updating failed, token: ", token.PreviousID, ", from: ", token.ID, ", error: ", err)
				}

				err = tokenRepository.DeleteToken(token.ID)

				if err != nil {
					fmt.Println("Deleting failed, token: ", token.PreviousID, ", from: ", token.ID, ", error: ", err)
				}
			} else {
				token.Status = "active"

				err = tokenRepository.UpdateToken(token.ID, token)

				if err != nil {
					fmt.Println("Updating failed, token: ", token.PreviousID, ", from: ", token.ID, ", error: ", err)
				}
			}
		} else {
			err = tokenRepository.DeleteToken(token.ID)

			if err != nil {
				fmt.Println("Deleting failed, token: ", token.PreviousID, ", from: ", token.ID, ", error: ", err)
			}
		}

		// Remove token cache
		cacheKey := "token-*"
		keys, err := redisClient.Keys(cacheKey).Result()
		if err == nil {
			for _, key := range keys {
				err = redisClient.Del(key).Err()

				if err != nil {
					return
				}
			}
		}
	}

	// Get pending ownerships
	ownerships, err := ownershipRepository.GetOwnershipList(0, 100000000, "", nil, &empty, nil, &empty, nil, status, "created_at", "ASC")

	if err != nil {
		fmt.Println("ownership")
		fmt.Println(err)
		return
	}

	// Check all pending ownerships
	for _, ownership := range ownerships {
		// Check if is in the monitoring only mode
		if monitoringOnly {
			break
		}

		transactionId := common.HexToHash(ownership.TransactionHash)

		// Get transaction receipt
		receipt, err := ethClient.TransactionReceipt(ctx, transactionId)

		if err != nil {
			log.Print(err)
			continue
		}

		if receipt.Status == 1 {
			if ownership.PreviousID != helpers.GetEmptyUUID() {
				oldOwnership, err := ownershipRepository.GetOwnershipData(ownership.PreviousID)

				if err != nil {
					fmt.Println("Failed to get, ownership: ", ownership.PreviousID, ", from: ", ownership.ID, ", error: ", err)
				}

				oldOwnership.Quantity = ownership.Quantity
				oldOwnership.SalePrice = ownership.SalePrice
				oldOwnership.RentCost = ownership.RentCost
				oldOwnership.AvailableForSale = ownership.AvailableForSale
				oldOwnership.AvailableForRent = ownership.AvailableForRent
				oldOwnership.UserID = ownership.UserID

				if ownership.Quantity > 0 {
					oldOwnership.Status = "active"
				} else {
					oldOwnership.Status = "inactive"
				}

				err = ownershipRepository.UpdateOwnership(oldOwnership.ID, oldOwnership)

				if err != nil {
					fmt.Println("Updating failed, ownership: ", ownership.PreviousID, ", from: ", ownership.ID, ", error: ", err)
				}

				err = ownershipRepository.DeleteOwnership(ownership.ID)

				if err != nil {
					fmt.Println("Deleting failed, ownership: ", ownership.PreviousID, ", from: ", ownership.ID, ", error: ", err)
				}
			} else {
				if ownership.Quantity > 0 {
					ownership.Status = "active"
				} else {
					ownership.Status = "inactive"
				}

				err = ownershipRepository.UpdateOwnership(ownership.ID, ownership)

				if err != nil {
					fmt.Println("Updating failed, ownership: ", ownership.PreviousID, ", from: ", ownership.ID, ", error: ", err)
				}
			}
		} else {
			err = ownershipRepository.DeleteOwnership(ownership.ID)

			if err != nil {
				fmt.Println("Deleting failed, ownership: ", ownership.PreviousID, ", from: ", ownership.ID, ", error: ", err)
			}
		}

		// Remove ownership cache
		cacheKey := "ownership-*"
		keys, err := redisClient.Keys(cacheKey).Result()
		if err != nil {
			for _, key := range keys {
				err = redisClient.Del(key).Err()

				if err != nil {

					return
				}
			}
		}
	}

	// Get pending collections
	collections, err := collectionRepository.GetCollectionList(0, 100000000, "", nil, &status, "created_at", "ASC")

	if err != nil {
		fmt.Println("collection")
		fmt.Println(err)
		return
	}

	// Check all pending collections
	for _, collection := range collections {
		// Check if is in the monitoring only mode
		if monitoringOnly {
			break
		}

		transactionId := common.HexToHash(collection.TransactionHash.String)

		// Get transaction receipt
		receipt, err := ethClient.TransactionReceipt(ctx, transactionId)

		if err != nil {
			log.Print(err)
			continue
		}

		if receipt.Status == 1 {
			if collection.PreviousID != helpers.GetEmptyUUID() {
				oldCollection, err := collectionRepository.GetCollectionData(collection.PreviousID)

				if err != nil {
					fmt.Println("Failed to get, collection: ", collection.PreviousID, ", from: ", collection.ID, ", error: ", err)
				}

				oldCollection.NumberOfItems = collection.NumberOfItems
				oldCollection.NumberOfTransactions = collection.NumberOfTransactions
				oldCollection.VolumeTransactions = collection.VolumeTransactions
				oldCollection.Floor = collection.Floor
				oldCollection.Views = collection.Views

				err = collectionRepository.UpdateCollection(oldCollection.ID, oldCollection)

				if err != nil {
					fmt.Println("Updating failed, collection: ", collection.PreviousID, ", from: ", collection.ID, ", error: ", err)
				}

				err = collectionRepository.DeleteCollection(collection.ID)

				if err != nil {
					fmt.Println("Deleting failed, collection: ", collection.PreviousID, ", from: ", collection.ID, ", error: ", err)
				}
			} else {
				collection.Status = sql.NullString{String: "active", Valid: true}

				err = collectionRepository.UpdateCollection(collection.ID, collection)

				if err != nil {
					fmt.Println("Updating failed, collection: ", collection.PreviousID, ", from: ", collection.ID, ", error: ", err)
				}
			}
		} else {
			err = collectionRepository.DeleteCollection(collection.ID)

			if err != nil {
				fmt.Println("Deleting failed, collection: ", collection.PreviousID, ", from: ", collection.ID, ", error: ", err)
			}
		}

		// Remove collection cache
		cacheKey := "collection-*"
		keys, err := redisClient.Keys(cacheKey).Result()
		if err == nil {
			for _, key := range keys {
				err = redisClient.Del(key).Err()

				if err != nil {

					return
				}
			}
		}
	}

	// Get pending transactions
	transactions, err := transactionRepository.GetTransactionList(0, 100000000, nil, &status, "created_at", "ASC")

	if err != nil {
		fmt.Println("transaction")
		fmt.Println(err)
		return
	}

	// Check all pending transactions
	for _, transaction := range transactions {
		// Check if is in the monitoring only mode
		if monitoringOnly {
			break
		}

		transactionId := common.HexToHash(transaction.TransactionHash)

		// Get transaction receipt
		receipt, err := ethClient.TransactionReceipt(ctx, transactionId)

		if err != nil {
			log.Print(err)
			continue
		}

		if receipt.Status == 1 {
			transaction.Status = "active"

			err = transactionRepository.UpdateTransaction(transaction.ID, transaction)

			if err != nil {
				fmt.Println("Updating failed, transaction: ", transaction.PreviousID, ", from: ", transaction.ID, ", error: ", err)
			}
		} else {
			err = transactionRepository.DeleteTransaction(transaction.ID)

			if err != nil {
				fmt.Println("Deleting failed, transaction: ", transaction.PreviousID, ", from: ", transaction.ID, ", error: ", err)
			}
		}

		// Remove transaction cache
		cacheKey := "transaction-*"
		keys, err := redisClient.Keys(cacheKey).Result()
		if err == nil {
			for _, key := range keys {
				err = redisClient.Del(key).Err()

				if err != nil {

					return
				}
			}
		}
	}

	// Get pending rentals
	rentals, err := rentalRepository.GetRentalList(0, 100000000, "", nil, &empty, nil, &empty, nil, &empty, nil, &status, "created_at", "ASC")

	if err != nil {
		fmt.Println("rental")
		fmt.Println(err)
		return
	}

	// Check all pending rentals
	for _, rental := range rentals {
		// Check if is in the monitoring only mode
		if monitoringOnly {
			break
		}

		transactionId := common.HexToHash(rental.TransactionHash)

		// Get transaction receipt
		receipt, err := ethClient.TransactionReceipt(ctx, transactionId)

		if err != nil {
			log.Print(err)
			continue
		}

		if receipt.Status == 1 {
			rental.Status = "active"

			err = rentalRepository.UpdateRental(rental.ID, rental)

			if err != nil {
				fmt.Println("Updating failed, rental: ", rental.PreviousID, ", from: ", rental.ID, ", error: ", err)
			}
		} else {
			err = rentalRepository.DeleteRental(rental.ID)

			if err != nil {
				fmt.Println("Deleting failed, rental: ", rental.PreviousID, ", from: ", rental.ID, ", error: ", err)
			}
		}
	}

	// Get pending fractions
	fractions, err := fractionRepository.GetFractionList(0, 100000000, "", nil, &empty, &status, "created_at", "ASC")

	if err != nil {
		fmt.Println("fraction")
		fmt.Println(err)
		return
	}

	// Check all pending fractions
	for _, fraction := range fractions {
		// Check if is in the monitoring only mode
		if monitoringOnly {
			break
		}

		transactionId := common.HexToHash(fraction.TransactionHash)

		fmt.Println("fraction", fraction.TransactionHash)

		// Get transaction receipt
		receipt, err := ethClient.TransactionReceipt(ctx, transactionId)

		if err != nil {
			log.Print(err)
			continue
		}

		if receipt.Status == 1 {
			fmt.Println("fraction", 1)
			fraction.Status = "active"

			err = fractionRepository.UpdateFraction(fraction.ID, fraction)

			if err != nil {
				fmt.Println("Updating failed, fraction: ", fraction.PreviousID, ", from: ", fraction.ID, ", error: ", err)
			}
		} else {
			fmt.Println("fraction", 0)
			err = fractionRepository.DeleteFraction(fraction.ID)

			if err != nil {
				fmt.Println("Deleting failed, fraction: ", fraction.PreviousID, ", from: ", fraction.ID, ", error: ", err)
			}
		}

		// Remove ownership cache
		cacheKey := "fraction-*"
		keys, err := redisClient.Keys(cacheKey).Result()
		if err == nil {
			for _, key := range keys {
				err = redisClient.Del(key).Err()

				if err != nil {

					return
				}
			}
		}
	}

	fmt.Println("Number of pending tokens : ", len(tokens))
	fmt.Println("Number of pending ownerships : ", len(ownerships))
	fmt.Println("Number of pending collections : ", len(collections))
	fmt.Println("Number of pending transactions : ", len(transactions))
	fmt.Println("Number of pending rentals : ", len(rentals))
	fmt.Println("Number of pending fractions : ", len(fractions))
	fmt.Println("--------------------------------------------")
}

func runCronJobs() {
	s := gocron.NewScheduler(time.UTC)

	s.Every(5).Seconds().Do(func() {
		checkBlock()
	})

	s.StartBlocking()
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
		os.Exit(1)
	}

	dbClient = config.CreateDBClient()
	redisClient = config.CreateRedisClient()

	var success bool
	rpcUrl, success = os.LookupEnv("RPC_URL")
	if !success {
		fmt.Fprintln(os.Stderr, "No RPC_URL - set the RPC_URL environment var and try again.")
		os.Exit(1)
	}

	var err error
	ethClient, err = ethclient.DialContext(ctx, rpcUrl)

	if err != nil {
		fmt.Println("Failed to connect with blockchain network: ", err)
		os.Exit(1)
	}

	collectionRepository = repositories.NewCollectionRepository(dbClient)
	fractionRepository = repositories.NewFractionRepository(dbClient)
	ownershipRepository = repositories.NewOwnershipRepository(dbClient)
	rentalRepository = repositories.NewRentalRepository(dbClient)
	tokenRepository = repositories.NewTokenRepository(dbClient)
	tokenCategoryRepository = repositories.NewTokenCategoryRepository(dbClient)
	transactionRepository = repositories.NewTransactionRepository(dbClient)
	userRepository = repositories.NewUserRepository(dbClient)

	runCronJobs()
}
