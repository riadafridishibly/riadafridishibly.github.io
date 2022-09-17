package main

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

func bytes2hex(b []byte) string {
	return hex.EncodeToString(b)
}

func hex2bytes(hexStr string) []byte {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return b
}

func decrypt(key, nonce, ciphertext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}
	return aead.Open(nil, nonce, ciphertext, nil)
}

func encrypt(key, nonce, plaintext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}
	return aead.Seal(nil, nonce, plaintext, nil), nil
}

func main() {
	plaintext := []byte("Hello, World!")
	key := hex2bytes("a5ad47647a055644c4d86c840c60c51f7955a3fc9a5cc85acee0e82b808beb64")
	nonce := hex2bytes("b0a6bb1fbd02787db8aa0059")
	ciphertext, err := encrypt(key, nonce, plaintext)
	if err != nil {
		panic(err)
	}
	fmt.Println("encrypted:", bytes2hex(ciphertext))

	decrypted, err := decrypt(key, nonce, ciphertext)
	if err != nil {
		panic(err)
	}
	fmt.Println("decrytped:", string(decrypted))

	// fmt.Println("plain text len:", len(plaintext))
	// fmt.Println("ciphertext len:", len(ciphertext))

	// If we use wrong password
	wrongKey := bytes.Repeat([]byte{'b'}, 32)
	decrypted, err = decrypt(wrongKey, nonce, ciphertext)
	if err != nil {
		fmt.Println("Err: Wrong Key: ", err)
	}
	// You'll get error: chacha20poly1305: message authentication failed
	fmt.Println("decrytped:", bytes2hex(decrypted))
}
