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

//  ECDSA key pair once
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


func GetPrivateKey() *ecdsa.PrivateKey {
    initializeKeys()
    return privateKey
}


func GetPublicKey() *ecdsa.PublicKey {
    initializeKeys()
    return &privateKey.PublicKey
}


func IsKeyInitialized() bool {
    return privateKey != nil
}