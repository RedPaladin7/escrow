package deck

import (
	"crypto/rand"
	"math/big"
)

// Deck represents a deck of cards
type Deck struct {
	Cards []Card
}

// NewDeck creates a standard 52-card deck
func NewDeck() *Deck {
	cards := make([]Card, 52)
	index := 0

	for suit := Hearts; suit <= Spades; suit++ {
		for value := 2; value <= 14; value++ {
			cards[index] = Card{Suit: suit, Value: value}
			index++
		}
	}

	return &Deck{Cards: cards}
}

// Shuffle shuffles the deck using crypto/rand
func (d *Deck) Shuffle() {
	n := len(d.Cards)
	for i := n - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			continue
		}
		j := int(jBig.Int64())
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	}
}

// Draw draws the top card from the deck
func (d *Deck) Draw() (Card, bool) {
	if len(d.Cards) == 0 {
		return Card{}, false
	}

	card := d.Cards[0]
	d.Cards = d.Cards[1:]
	return card, true
}

// DrawN draws n cards from the deck
func (d *Deck) DrawN(n int) []Card {
	if n > len(d.Cards) {
		n = len(d.Cards)
	}

	cards := make([]Card, n)
	copy(cards, d.Cards[:n])
	d.Cards = d.Cards[n:]
	return cards
}

// Reset resets the deck to a full 52-card deck
func (d *Deck) Reset() {
	*d = *NewDeck()
}

// Remaining returns the number of cards remaining in the deck
func (d *Deck) Remaining() int {
	return len(d.Cards)
}

// ToBytes converts the entire deck to a byte representation
func (d *Deck) ToBytes() [][]byte {
	bytes := make([][]byte, len(d.Cards))
	for i, card := range d.Cards {
		bytes[i] = card.ToBytes()
	}
	return bytes
}

// FromBytes creates a deck from byte representation
func FromBytes(bytes [][]byte) *Deck {
	cards := make([]Card, len(bytes))
	for i, b := range bytes {
		if len(b) > 0 {
			cards[i] = NewCardFromByte(b[0])
		}
	}
	return &Deck{Cards: cards}
}

// Contains checks if the deck contains a specific card
func (d *Deck) Contains(card Card) bool {
	for _, c := range d.Cards {
		if c.Suit == card.Suit && c.Value == card.Value {
			return true
		}
	}
	return false
}

// Remove removes a specific card from the deck
func (d *Deck) Remove(card Card) bool {
	for i, c := range d.Cards {
		if c.Suit == card.Suit && c.Value == card.Value {
			d.Cards = append(d.Cards[:i], d.Cards[i+1:]...)
			return true
		}
	}
	return false
}

// Clone creates a copy of the deck
func (d *Deck) Clone() *Deck {
	cards := make([]Card, len(d.Cards))
	copy(cards, d.Cards)
	return &Deck{Cards: cards}
}
