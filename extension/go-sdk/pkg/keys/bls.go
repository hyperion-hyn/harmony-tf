package keys

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto/bls"
	bls_core "github.com/hyperion-hyn/bls/ffi/go/bls"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	"io"
	"os"
	"path"
)

//BlsKey - struct to represent bls key data
type BlsKey struct {
	PrivateKey    *bls_core.SecretKey
	PublicKey     *bls_core.PublicKey
	PublicKeyHex  string
	PrivateKeyHex string
	Passphrase    string
	FilePath      string
	//ShardPublicKey *bls.SerializedPublicKey
}

//Initialize - initialize a bls key and assign a random private bls key if not already done
func (blsKey *BlsKey) Initialize() {
	if blsKey.PrivateKey == nil {
		blsKey.PrivateKey = bls.RandPrivateKey()
		blsKey.PrivateKeyHex = blsKey.PrivateKey.SerializeToHexStr()
		blsKey.PublicKey = blsKey.PrivateKey.GetPublicKey()
		blsKey.PublicKeyHex = blsKey.PublicKey.SerializeToHexStr()
	}
}

//Reset - resets the currently assigned private and public key fields
func (blsKey *BlsKey) Reset() {
	blsKey.PrivateKey = nil
	blsKey.PrivateKeyHex = ""
	blsKey.PublicKey = nil
	blsKey.PublicKeyHex = ""
}

// GenBlsKey - generate a random bls key using the supplied passphrase, write it to disk at the given filePath
func GenBlsKey(blsKey *BlsKey) error {
	blsKey.Initialize()
	out, err := writeBlsKeyToFile(blsKey)
	if err != nil {
		return err
	}
	fmt.Println(common.JSONPrettyFormat(out))
	return nil
}

func writeBlsKeyToFile(blsKey *BlsKey) (string, error) {
	if blsKey.FilePath == "" {
		cwd, _ := os.Getwd()
		blsKey.FilePath = fmt.Sprintf("%s/%s.key", cwd, blsKey.PublicKeyHex)
	}
	if !path.IsAbs(blsKey.FilePath) {
		return "", common.ErrNotAbsPath
	}
	encryptedPrivateKeyStr, err := encrypt([]byte(blsKey.PrivateKeyHex), blsKey.Passphrase)
	if err != nil {
		return "", err
	}
	err = writeToFile(blsKey.FilePath, encryptedPrivateKeyStr)
	if err != nil {
		return "", err
	}
	out := fmt.Sprintf(`
{"public-key" : "%s", "private-key" : "%s", "encrypted-private-key-path" : "%s"}`,
		blsKey.PublicKeyHex, blsKey.PrivateKeyHex, blsKey.FilePath)

	return out, nil
}

func writeToFile(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	return file.Sync()
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) (string, error) {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return hex.EncodeToString(ciphertext), nil
}

func decrypt(encrypted []byte, passphrase string) (decrypted []byte, err error) {
	unhexed := make([]byte, hex.DecodedLen(len(encrypted)))
	if _, err = hex.Decode(unhexed, encrypted); err == nil {
		if decrypted, err = decryptRaw(unhexed, passphrase); err == nil {
			return decrypted, nil
		}
	}
	// At this point err != nil, either from hex decode or from decryptRaw.
	decrypted, binErr := decryptRaw(encrypted, passphrase)
	if binErr != nil {
		// Disregard binary decryption error and return the original error,
		// because our canonical form is hex and not binary.
		return nil, err
	}
	return decrypted, nil
}

func decryptRaw(data []byte, passphrase string) ([]byte, error) {
	var err error
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	return plaintext, err
}
