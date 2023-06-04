package config

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func CreateDBClient() *sql.DB {
	fmt.Printf("Initialize Postgres client...")

	postgresHost, success := os.LookupEnv("POSTGRES_HOST")
	fmt.Println(postgresHost)
	if !success {
		fmt.Fprintln(os.Stderr, "No POSTGRES_HOST - set the POSTGRES_HOST environment var and try again.")
		os.Exit(1)
	}

	postgresPort, success := os.LookupEnv("POSTGRES_PORT")
	if !success {
		fmt.Fprintln(os.Stderr, "No POSTGRES_PORT - set the POSTGRES_PORT environment var and try again.")
		os.Exit(1)
	}

	postgresUsername, success := os.LookupEnv("POSTGRES_USERNAME")
	if !success {
		fmt.Fprintln(os.Stderr, "No POSTGRES_USERNAME - set the POSTGRES_USERNAME environment var and try again.")
		os.Exit(1)
	}

	postgresPassword, success := os.LookupEnv("POSTGRES_PASSWORD")
	if !success {
		fmt.Fprintln(os.Stderr, "No POSTGRES_PASSWORD - set the POSTGRES_PASSWORD environment var and try again.")
		os.Exit(1)
	}

	postgresDB, success := os.LookupEnv("POSTGRES_DB")
	if !success {
		fmt.Fprintln(os.Stderr, "No POSTGRES_DB - set the POSTGRES_DB environment var and try again.")
		os.Exit(1)
	}

	postgresSSLMode, success := os.LookupEnv("POSTGRES_SSL_MODE")
	if !success {
		fmt.Fprintln(os.Stderr, "No POSTGRES_SSL_MODE - set the POSTGRES_SSL_MODE environment var and try again.")
		os.Exit(1)
	}

	//	postgresSource := fmt.Sprintf("postgresql://admin:password123@localhost:6500/metaedu?sslmode=disable")
	postgresSource := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s", postgresUsername, postgresPassword, postgresHost, postgresPort, postgresDB, postgresSSLMode)

	db, err := sql.Open("postgres", postgresSource)

	if !success {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	err = db.Ping()

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	return db
}
