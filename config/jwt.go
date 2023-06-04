package config

import (
	"fmt"
	"metaedu-marketplace/utils"
	"os"
	"time"
)

func CreateJwtHmacProvider() *utils.JwtHmacProvider {
	fmt.Printf("Initialize JWT provider...")

	jwtHmacSecretKey, success := os.LookupEnv("JWT_HMAC_SECRET_KEY")
	if !success {
		fmt.Fprintln(os.Stderr, "No JWT_HMAC_SECRET_KEY - set the JWT_HMAC_SECRET_KEY environment var and try again.")
		os.Exit(1)
	}

	jwtIssuer, success := os.LookupEnv("JWT_ISSUER")
	if !success {
		fmt.Fprintln(os.Stderr, "No JWT_ISSUER - set the JWT_ISSUER environment var and try again.")
		os.Exit(1)
	}

	jwtHmacProvider := utils.NewJwtHmacProvider(
		jwtHmacSecretKey,
		jwtIssuer,
		time.Hour*600,
	)

	return jwtHmacProvider
}
