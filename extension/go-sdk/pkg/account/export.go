package account

import (
	"errors"
)

var (
	ErrAddressNotFound = errors.New("account was not found in keystore")
)
