package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// CardKeys represents encryption/decryption keys for mental poker
type CardKeys struct {
	EncKey *big.Int
	DecKey *big.Int
	Prime  *big.Int
}

// GenerateCardKeys generates a new pair of encryption/decryption keys
func GenerateCardKeys() (*CardKeys, error) {
	// Use a large shared prime for the mental poker protocol
	// In production, this should be agreed upon by all players
	sharedPrime, success := new(big.Int).SetString("C7970CEDCC5226685694605929849D3D", 16)
	if !success {
		return nil, fmt.Errorf("failed to set shared prime")
	}

	return GenerateCardKeysWithPrime(sharedPrime)
}

// GenerateCardKeysWithPrime generates keys with a specific prime
func GenerateCardKeysWithPrime(prime *big.Int) (*CardKeys, error) {
	// Generate random encryption key
	encKey, err := generateRandomKey(prime)
	if err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Calculate decryption key (modular inverse)
	phiN := new(big.Int).Sub(prime, big.NewInt(1))
	decKey := new(big.Int).ModInverse(encKey, phiN)
	if decKey == nil {
		return nil, fmt.Errorf("failed to calculate decryption key")
	}

	return &CardKeys{
		EncKey: encKey,
		DecKey: decKey,
		Prime:  prime,
	}, nil
}

// generateRandomKey generates a random key that is coprime with (prime - 1)
func generateRandomKey(prime *big.Int) (*big.Int, error) {
	phiN := new(big.Int).Sub(prime, big.NewInt(1))
	maxAttempts := 1000

	for i := 0; i < maxAttempts; i++ {
		// Generate random number in range [2, prime-2]
		key, err := rand.Int(rand.Reader, new(big.Int).Sub(prime, big.NewInt(2)))
		if err != nil {
			return nil, err
		}
		key.Add(key, big.NewInt(2))

		// Check if key is coprime with phi(n)
		gcd := new(big.Int).GCD(nil, nil, key, phiN)
		if gcd.Cmp(big.NewInt(1)) == 0 {
			return key, nil
		}
	}

	return nil, fmt.Errorf("failed to generate coprime key after %d attempts", maxAttempts)
}

// Encrypt encrypts a byte array using the encryption key
func (ck *CardKeys) Encrypt(data []byte) []byte {
	// Convert bytes to big.Int
	plaintext := new(big.Int).SetBytes(data)

	// Encrypt: ciphertext = plaintext^encKey mod prime
	ciphertext := new(big.Int).Exp(plaintext, ck.EncKey, ck.Prime)

	return ciphertext.Bytes()
}

// Decrypt decrypts a byte array using the decryption key
func (ck *CardKeys) Decrypt(data []byte) []byte {
	// Convert bytes to big.Int
	ciphertext := new(big.Int).SetBytes(data)

	// Decrypt: plaintext = ciphertext^decKey mod prime
	plaintext := new(big.Int).Exp(ciphertext, ck.DecKey, ck.Prime)

	return plaintext.Bytes()
}

// EncryptMultiple encrypts multiple data chunks
func (ck *CardKeys) EncryptMultiple(dataList [][]byte) [][]byte {
	encrypted := make([][]byte, len(dataList))
	for i, data := range dataList {
		encrypted[i] = ck.Encrypt(data)
	}
	return encrypted
}

// DecryptMultiple decrypts multiple data chunks
func (ck *CardKeys) DecryptMultiple(dataList [][]byte) [][]byte {
	decrypted := make([][]byte, len(dataList))
	for i, data := range dataList {
		decrypted[i] = ck.Decrypt(data)
	}
	return decrypted
}

// Clone creates a copy of the CardKeys
func (ck *CardKeys) Clone() *CardKeys {
	return &CardKeys{
		EncKey: new(big.Int).Set(ck.EncKey),
		DecKey: new(big.Int).Set(ck.DecKey),
		Prime:  new(big.Int).Set(ck.Prime),
	}
}

// Validate checks if the keys are valid
func (ck *CardKeys) Validate() error {
	if ck.EncKey == nil || ck.DecKey == nil || ck.Prime == nil {
		return fmt.Errorf("keys are not initialized")
	}

	// Verify that encKey * decKey ≡ 1 (mod prime-1)
	phiN := new(big.Int).Sub(ck.Prime, big.NewInt(1))
	product := new(big.Int).Mul(ck.EncKey, ck.DecKey)
	modResult := new(big.Int).Mod(product, phiN)

	if modResult.Cmp(big.NewInt(1)) != 0 {
		return fmt.Errorf("invalid key pair: encryption and decryption keys don't match")
	}

	return nil
}
