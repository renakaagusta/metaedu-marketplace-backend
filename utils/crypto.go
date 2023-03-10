package utils

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func Verify(address string, text string, sigHex string) error {
	sig := hexutil.MustDecode(sigHex)
	// https://github.com/ethereum/go-ethereum/blob/master/internal/ethapi/api.go#L516
	// check here why I am subtracting 27 from the last byte
	sig[crypto.RecoveryIDOffset] -= 27
	msg := accounts.TextHash([]byte(text))
	recovered, err := crypto.SigToPub(msg, sig)
	if err != nil {
		return err
	}
	recoveredAddr := crypto.PubkeyToAddress(*recovered)

	if address != strings.ToLower(recoveredAddr.Hex()) {
		return ErrAuthError
	}

	return nil
}
