package common

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
)

func KeyStoreForPath(p string) *keystore.KeyStore {
	return keystore.NewKeyStore(p, ScryptN, ScryptP)
}
