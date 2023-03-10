package config

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/web3-storage/go-w3s-client"
)

func CreateWeb3StorageClient() w3s.Client {
	fmt.Printf("Initialize WEB3 storage client...")

	web3StorageToken, success := os.LookupEnv("WEB3_STORAGE_TOKEN")
	if !success {
		fmt.Fprintln(os.Stderr, "No API token - set the WEB3_STORAGE_TOKEN environment var and try again.")
		os.Exit(1)
	}

	web3StorageClient, err := w3s.NewClient(w3s.WithToken(web3StorageToken))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	return web3StorageClient
}
