package crypto

import (
	"crypto/rand"
	"math/big"
)

// ShuffleDeck performs a cryptographically secure shuffle of the deck
func ShuffleDeck(deck [][]byte) [][]byte {
	n := len(deck)
	shuffled := make([][]byte, n)
	copy(shuffled, deck)

	// Fisher-Yates shuffle with crypto/rand
	for i := n - 1; i > 0; i-- {
		// Generate random index j where 0 <= j <= i
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			// Fallback to non-crypto random on error (should never happen)
			continue
		}
		j := int(jBig.Int64())

		// Swap elements
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}

// ShuffleIndices generates a random permutation of indices
func ShuffleIndices(n int) []int {
	indices := make([]int, n)
	for i := 0; i < n; i++ {
		indices[i] = i
	}

	// Fisher-Yates shuffle
	for i := n - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			continue
		}
		j := int(jBig.Int64())
		indices[i], indices[j] = indices[j], indices[i]
	}

	return indices
}

// ApplyPermutation applies a permutation to a deck
func ApplyPermutation(deck [][]byte, permutation []int) [][]byte {
	if len(deck) != len(permutation) {
		return deck
	}

	shuffled := make([][]byte, len(deck))
	for i, idx := range permutation {
		shuffled[i] = deck[idx]
	}

	return shuffled
}

// VerifyShuffle checks if a deck has been properly shuffled (not in original order)
func VerifyShuffle(original, shuffled [][]byte) bool {
	if len(original) != len(shuffled) {
		return false
	}

	changedCount := 0
	for i := range original {
		if !bytesEqual(original[i], shuffled[i]) {
			changedCount++
		}
	}

	// At least 80% of cards should be in different positions
	return changedCount >= len(original)*4/5
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
