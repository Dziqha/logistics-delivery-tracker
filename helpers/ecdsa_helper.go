package helpers

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "log"
    "sync"
)

var (
    privateKey *ecdsa.PrivateKey
    once       sync.Once
)

// initializeKeys generates ECDSA key pair once
func initializeKeys() {
    once.Do(func() {
        key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
        if err != nil {
            log.Fatalf("Failed to generate ECDSA private key: %v", err)
        }
        privateKey = key
        log.Println("ECDSA key pair generated successfully")
    })
}

// GetPrivateKey returns the ECDSA private key for token signing
// This should only be used in secure contexts (like token generation)
func GetPrivateKey() *ecdsa.PrivateKey {
    initializeKeys()
    return privateKey
}

// GetPublicKey returns the ECDSA public key for token verification
// This is safe to use in middleware and verification contexts
func GetPublicKey() *ecdsa.PublicKey {
    initializeKeys()
    return &privateKey.PublicKey
}

// IsKeyInitialized checks if the ECDSA key pair has been initialized
func IsKeyInitialized() bool {
    return privateKey != nil
}