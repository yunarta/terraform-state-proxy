package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log"
)

type State struct {
	AESIV     string `json:"aes_iv"`
	Encrypted string `json:"encrypted"`
}

func (s State) Decrypt(key string) ([]byte, error) {
	var (
		hash                      [32]byte
		aesKey, iv, encryptedData []byte
		block                     cipher.Block
		mode                      cipher.BlockMode
		decrypted                 []byte
		length, padding           int
		err                       error
	)

	// Convert the key to a 32-byte AES key using SHA256
	hash = sha256.Sum256([]byte(key))
	aesKey = hash[:]

	// Decode the AES IV and Encrypted data from Base64
	iv, err = base64.StdEncoding.DecodeString(s.AESIV)
	if err != nil {
		log.Println("Error decoding IV:", err)
		return nil, err
	}

	encryptedData, err = base64.StdEncoding.DecodeString(s.Encrypted)
	if err != nil {
		return nil, err
	}

	// Create a new AES cipher block
	block, err = aes.NewCipher(aesKey)
	if err != nil {
		log.Println("Error creating AES cipher:", err)
		return nil, err
	}

	// Ensure IV length matches the block size
	if len(iv) != aes.BlockSize {
		log.Println("Invalid IV length")
		return nil, err
	}

	// Use CBC mode for decryption
	mode = cipher.NewCBCDecrypter(block, iv)

	// Decrypt the data
	decrypted = make([]byte, len(encryptedData))
	mode.CryptBlocks(decrypted, encryptedData)

	// Unpad the decrypted data (PKCS7 padding)
	length = len(decrypted)
	padding = int(decrypted[length-1])
	if padding > aes.BlockSize || padding > length {
		log.Println("Invalid padding")
		return nil, err
	}
	decrypted = decrypted[:length-padding]

	return decrypted, nil
}

func NewEncryptedState(key string, plaintext []byte) (*State, error) {
	var (
		hash       [32]byte
		aesKey     []byte
		block      cipher.Block
		iv         []byte
		paddedData []byte
		encrypted  []byte
		err        error
	)

	// Convert the key to a 32-byte AES key using SHA256
	hash = sha256.Sum256([]byte(key))
	aesKey = hash[:]

	// Create a new AES cipher block
	block, err = aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	// Generate an IV with the same length as AES block size
	iv = make([]byte, aes.BlockSize)
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return nil, err
	}

	// Apply PKCS7 padding to the plaintext
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	paddedData = make([]byte, len(plaintext)+padding)
	copy(paddedData, plaintext)
	for i := len(plaintext); i < len(paddedData); i++ {
		paddedData[i] = byte(padding)
	}

	// Encrypt the padded data
	encrypted = make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, paddedData)

	// Encode the IV and encrypted data to Base64
	ivBase64 := base64.StdEncoding.EncodeToString(iv)
	encryptedBase64 := base64.StdEncoding.EncodeToString(encrypted)

	return &State{
		AESIV:     ivBase64,
		Encrypted: encryptedBase64,
	}, err
}

type Request struct {
	Project    string `uri:"project" binding:"required"`
	Repository string `uri:"repo" binding:"required"`
	Path       string `uri:"path" binding:"required"`
}
