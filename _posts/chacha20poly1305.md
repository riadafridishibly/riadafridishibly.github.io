---
published: false
comments: true
layout: post
title: Symmetric Encryption With ChaCha20Poly1305
author: Riad Afridi Shibly
categories: programming
tags: [golang, encryption, decryption]
image: 2022-07-14-chacha20poly1305/banner.jpg
---

# Table Of Contents

{:.no_toc}
- 
{:toc}

In this note, I'll try to explain Symmetric Encryption with ChaCha20 and Poly1305. We'll build up the required knowledge to understand and implement Symmetric Encryption in golang. I've chosen golang mostly because I've been working with golang for quite some time and I know its crypto library. `golang` is easy to understand. I think you won't have any problems reading the code. With that being said let's dive into the madness.

READ THE PRACTICAL CRYPTO BOOK

# Symmetric Encryption

Before going any further we need to understand what Symmetric Encryption is. It's a process where encryption and decryption are performed with one single key. So we'll be using a `key` to encrypt and as well as decrypt a message.

## Terminology

<!-- Refactor This -->

Let's learn some terminology first. We'll call our data (what we want to encrypt) `plaintext`. We'll encrypt the `plaintext` and get `ciphertext`. We'll use a `key` which is our secret. So, we'll encrypt `plaintext` using the `key` and get `ciphertext`.

## ChaCha20

Now that we know the terminologies and understand what Symmetric Encryption is, let's see how can we perform this in our case.

> ChaCha20 is the succsessor of Salsa20. Both are designed by `djb` (Daniel J. Bernstein) in 2005. More on this guy (`djb`) later.

Let's dive into code. We'll start small. `ChaCha20` uses 32 bytes key size. We want our key to be cryptographically secure. We'll see the security aspect of this later. Let's roll with the current implementation.

```go
var plaintext = []byte("hello, world!")
var key = []byte("bb roy good boy very good worker")
var nonce = bytes.Repeat([]byte{'s'}, 12)
```

We've defined our plaintext and key. Now we want to encrypt the plaintext. To do that we'll use `x/crypto/chacha20` library. 

Let's start with this function.

```go
func NewUnauthenticatedCipher(key, nonce []byte) (*Cipher, error)
```

Why is this called Unauthenticated? We'll come to that in a moment.

The `NewUnauthenticatedCipher` function takes two arguments, one is key and another is nonce. We can provide `nonce` as **12** or **24** bytes slice. We're currently dealing with ChaCha20 so we'll use **12** bytes of nonce.

Now we've all things set. You can run the code here. [GoPlayground](https://go.dev/play/p/LBh9LBiK8XJ)

```go
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/chacha20"
)

func encrypt(key, nonce, plaintext []byte) (ciphertext []byte) {
	ciphertext = make([]byte, len(plaintext))
	cipher, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		panic(err)
	}
	cipher.XORKeyStream(ciphertext, plaintext)
	return
}

func decrypt(key, nonce, ciphertext []byte) (plaintext []byte) {
	plaintext = make([]byte, len(ciphertext))
	cipher, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		panic(err)
	}
	cipher.XORKeyStream(plaintext, ciphertext)
	return
}

func main() {
	var plaintext = []byte("hello, world!")
	var key = []byte("bb roy good boy very good worker")
	var nonce = bytes.Repeat([]byte{'s'}, 12)

	ciphertext := encrypt(key, nonce, plaintext)
	plaintext = decrypt(key, nonce, ciphertext)

	fmt.Println("encrypted:", hex.EncodeToString(ciphertext))
	fmt.Println("plaintext:", string(plaintext))
}
```

কী পয়েন্টস,

- সিমেট্রিক এনক্রিপশনের বেসিক।
    - আমার একটা কী এবং টেক্সট আছে। আমি কী দিয়ে টেক্সটটা এনক্রিপ্ট করতে চাই

ChaCha20 এর বেসিক

- আমার একটা ৩২ বাইটের কী লাগবে
- আমি যেকোন টেক্সট এনক্রিপ্ট করবো
- তারপর আবার ডিক্রিপ্ট করবো‌।

কী ডেরিভেশন

- পাসওয়ারড তো যেকোন সাইজের হতে পারে। তাই আমরা পাসওয়ার্ড থেকে ফিক্সড সাইজের কী ডিরাইভ করবো
- আমরা scrypt এবং argon2id এই দুইটা এলগরিদম দেখবো

এখন ChaCha20 তে ভুল কী দিয়ে ডেটা ডিক্রিপ্ট করবো।

- ডেটা ডিক্রিপ্ট করার পর কিভাবে বুঝবো আমার পাসওয়ার্ড ভুল কিনা!

অথেন্টিকেশন কোড, ম্যাকের ইন্ট্রো এখানে

- HMAC
- Poly1305

এনক্রিপ্ট দেন অথেন্টিকেট

## Key Derivation

Previously we've seen that we must use exactly `32 bytes` (i.e. 256 bit) key to encrypt with ChaCha20. But in general, passwords are not exactly 32 bytes. Users can choose a random password which may or may not be 32 bytes. So, we need to tackle this problem. Luckily there's already a solution to this problem. Which is using `KDF` or `Key Derivation Functions`. **In simple terms, `KDF` is a type of function which takes an arbitrary-sized password as input and returns a fixed-size key.** Which is _cryptographically_ secure of course. 

```go
// The go function definition should look like this
// KDF takes arbitrary sized password and returns `keyLen` sized key
func KDF(password []byte,  keyLen int) (key []byte)
```

```go
// key should be randomly generated or derived from a function like Argon2.
key := make([]byte, KeySize)
if _, err := cryptorand.Read(key); err != nil {
  panic(err)
}

aead, err := NewX(key)
if err != nil {
  panic(err)
}

// Encryption.
var encryptedMsg []byte
{
  msg := []byte("Gophers, gophers, gophers everywhere!")

  // Select a random nonce, and leave capacity for the ciphertext.
  nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(msg)+aead.Overhead())
  if _, err := cryptorand.Read(nonce); err != nil {
    panic(err)
  }

  // Encrypt the message and append the ciphertext to the nonce.
  encryptedMsg = aead.Seal(nonce, nonce, msg, nil)
}

// Decryption.
{
  if len(encryptedMsg) < aead.NonceSize() {
    panic("ciphertext too short")
  }

  // Split nonce and ciphertext.
  nonce, ciphertext := encryptedMsg[:aead.NonceSize()], encryptedMsg[aead.NonceSize():]

  // Decrypt the message and check it wasn't tampered with.
  plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
  if err != nil {
    panic(err)
  }

  fmt.Printf("%s\n", plaintext)
}
```

## ChaCha20

### Encrypt

### Decrypt

## MAC (Message Authentication Code)

### HMAC-SHA-256

### Poly1305

## The End

## While writing this article...
