package crypto

import (
	"encoding/hex"
	"fmt"
	"math/big"
)

// SerializeKeys converts CardKeys to string format for transmission
type SerializedKeys struct {
	EncKey string `json:"enc_key"`
	DecKey string `json:"dec_key"`
	Prime  string `json:"prime"`
}

// Serialize converts CardKeys to SerializedKeys
func (ck *CardKeys) Serialize() SerializedKeys {
	return SerializedKeys{
		EncKey: ck.EncKey.Text(16),
		DecKey: ck.DecKey.Text(16),
		Prime:  ck.Prime.Text(16),
	}
}

// DeserializeKeys converts SerializedKeys back to CardKeys
func DeserializeKeys(sk SerializedKeys) (*CardKeys, error) {
	encKey := new(big.Int)
	if _, ok := encKey.SetString(sk.EncKey, 16); !ok {
		return nil, fmt.Errorf("invalid encryption key format")
	}

	decKey := new(big.Int)
	if _, ok := decKey.SetString(sk.DecKey, 16); !ok {
		return nil, fmt.Errorf("invalid decryption key format")
	}

	prime := new(big.Int)
	if _, ok := prime.SetString(sk.Prime, 16); !ok {
		return nil, fmt.Errorf("invalid prime format")
	}

	return &CardKeys{
		EncKey: encKey,
		DecKey: decKey,
		Prime:  prime,
	}, nil
}

// ToHex converts a byte slice to hex string
func ToHex(data []byte) string {
	return hex.EncodeToString(data)
}

// FromHex converts a hex string to byte slice
func FromHex(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

// EncryptDeck encrypts an entire deck of cards
func EncryptDeck(deck [][]byte, keys *CardKeys) [][]byte {
	encrypted := make([][]byte, len(deck))
	for i, card := range deck {
		encrypted[i] = keys.Encrypt(card)
	}
	return encrypted
}

// DecryptDeck decrypts an entire deck of cards
func DecryptDeck(deck [][]byte, keys *CardKeys) [][]byte {
	decrypted := make([][]byte, len(deck))
	for i, card := range deck {
		decrypted[i] = keys.Decrypt(card)
	}
	return decrypted
}

// DecryptSpecificCards decrypts specific cards from a deck by indices
func DecryptSpecificCards(deck [][]byte, indices []int, keys *CardKeys) map[int][]byte {
	decrypted := make(map[int][]byte)
	for _, idx := range indices {
		if idx >= 0 && idx < len(deck) {
			decrypted[idx] = keys.Decrypt(deck[idx])
		}
	}
	return decrypted
}

// VerifyDecryption verifies that encryption/decryption is reversible
func VerifyDecryption(original []byte, keys *CardKeys) bool {
	encrypted := keys.Encrypt(original)
	decrypted := keys.Decrypt(encrypted)
	
	if len(original) != len(decrypted) {
		return false
	}
	
	for i := range original {
		if original[i] != decrypted[i] {
			return false
		}
	}
	
	return true
}

// CombineDecryption applies multiple decryption keys in sequence
func CombineDecryption(data []byte, keysList []*CardKeys) []byte {
	result := data
	for _, keys := range keysList {
		result = keys.Decrypt(result)
	}
	return result
}

// CombineEncryption applies multiple encryption keys in sequence
func CombineEncryption(data []byte, keysList []*CardKeys) []byte {
	result := data
	for _, keys := range keysList {
		result = keys.Encrypt(result)
	}
	return result
}
