package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
)

func CreateRedisClient() *redis.Client {
	fmt.Println("Initialize Redis client...")

	redisHost, success := os.LookupEnv("REDIS_HOST")
	if !success {
		fmt.Fprintln(os.Stderr, "No REDIS_HOST - set the REDIS_HOST environment var and try again.")
		os.Exit(1)
	}

	redisPort, success := os.LookupEnv("REDIS_PORT")
	if !success {
		fmt.Fprintln(os.Stderr, "No REDIS_PORT - set the REDIS_PORT environment var and try again.")
		os.Exit(1)
	}

	redisPassword, success := os.LookupEnv("REDIS_PASSWORD")
	if !success {
		fmt.Fprintln(os.Stderr, "No REDIS_PORT - set the REDIS_PORT environment var and try again.")
		os.Exit(1)
	}

	redisDB, success := os.LookupEnv("REDIS_DB")
	if !success {
		fmt.Fprintln(os.Stderr, "No REDIS_DB - set the REDIS_DB environment var and try again.")
		os.Exit(1)
	}

	redisDbFormat, err := strconv.Atoi(redisDB)

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	redisAddress := fmt.Sprintf("%s:%s", redisHost, redisPort)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: redisPassword,
		DB:       redisDbFormat,
	})

	return redisClient
}
