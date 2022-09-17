package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

func decrypt(key, ciphertext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < aead.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	// Split nonce and ciphertext.
	nonce, ciphertext := ciphertext[:aead.NonceSize()], ciphertext[aead.NonceSize():]

	// Decrypt the message and check it wasn't tampered with.
	return aead.Open(nil, nonce, ciphertext, nil)
}

func encrypt(key, plaintext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	totalLen := aead.NonceSize() + len(plaintext) + aead.Overhead()
	nonce := make([]byte, aead.NonceSize(), totalLen)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt the message and append the ciphertext to the nonce.
	return aead.Seal(nonce, nonce, plaintext, nil), nil
}

func main() {
	plaintext := []byte("Hello, World!")
	key := hex2bytes("a5ad47647a055644c4d86c840c60c51f7955a3fc9a5cc85acee0e82b808beb64")
	ciphertext, err := encrypt(key, plaintext)
	if err != nil {
		panic(err)
	}
	fmt.Println("encrypted:", bytes2hex(ciphertext))

	decrypted, err := decrypt(key, ciphertext)
	if err != nil {
		panic(err)
	}
	fmt.Println("decrytped:", string(decrypted))

	// fmt.Println("plain text len:", len(plaintext))
	// fmt.Println("ciphertext len:", len(ciphertext))

	// If we use wrong password
	wrongKey := bytes.Repeat([]byte{'b'}, 32)
	decrypted, err = decrypt(wrongKey, ciphertext)
	if err != nil {
		fmt.Println("Err: Wrong Key: ", err)
	}
	// You'll get error: chacha20poly1305: message authentication failed
	fmt.Println("decrytped:", bytes2hex(decrypted))
}

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
