package deck

import "fmt"

// Suit represents a card suit
type Suit int

const (
	Hearts Suit = iota
	Diamonds
	Clubs
	Spades
)

func (s Suit) String() string {
	switch s {
	case Hearts:
		return "Hearts"
	case Diamonds:
		return "Diamonds"
	case Clubs:
		return "Clubs"
	case Spades:
		return "Spades"
	default:
		return "Unknown"
	}
}

func (s Suit) Symbol() string {
	switch s {
	case Hearts:
		return "♥"
	case Diamonds:
		return "♦"
	case Clubs:
		return "♣"
	case Spades:
		return "♠"
	default:
		return "?"
	}
}

// Card represents a playing card
type Card struct {
	Suit  Suit
	Value int // 2-14 (11=Jack, 12=Queen, 13=King, 14=Ace)
}

// NewCard creates a new card
func NewCard(suit Suit, value int) Card {
	return Card{
		Suit:  suit,
		Value: value,
	}
}

// NewCardFromByte creates a card from a byte representation
func NewCardFromByte(b byte) Card {
	value := int(b/4) + 2
	suit := Suit(b % 4)
	return Card{
		Suit:  suit,
		Value: value,
	}
}

// ToByte converts a card to byte representation
func (c Card) ToByte() byte {
	return byte((c.Value-2)*4 + int(c.Suit))
}

// String returns a string representation of the card
func (c Card) String() string {
	var valueName string
	switch c.Value {
	case 14:
		valueName = "A"
	case 13:
		valueName = "K"
	case 12:
		valueName = "Q"
	case 11:
		valueName = "J"
	case 10:
		valueName = "10"
	default:
		valueName = fmt.Sprintf("%d", c.Value)
	}
	return valueName + c.Suit.Symbol()
}

// FullName returns the full name of the card
func (c Card) FullName() string {
	var valueName string
	switch c.Value {
	case 14:
		valueName = "Ace"
	case 13:
		valueName = "King"
	case 12:
		valueName = "Queen"
	case 11:
		valueName = "Jack"
	default:
		valueName = fmt.Sprintf("%d", c.Value)
	}
	return valueName + " of " + c.Suit.String()
}

// ToBytes converts a card to a byte slice for encryption
func (c Card) ToBytes() []byte {
	return []byte{c.ToByte()}
}

// IsValid checks if a card is valid
func (c Card) IsValid() bool {
	return c.Value >= 2 && c.Value <= 14 && c.Suit >= Hearts && c.Suit <= Spades
}

// Compare compares two cards by value (for sorting)
func (c Card) Compare(other Card) int {
	if c.Value > other.Value {
		return 1
	} else if c.Value < other.Value {
		return -1
	}
	return 0
}
