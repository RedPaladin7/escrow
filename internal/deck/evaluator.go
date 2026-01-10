package deck

import (
	"sort"
)

// HandRank represents the rank of a poker hand
type HandRank int

const (
	HighCard HandRank = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

func (hr HandRank) String() string {
	switch hr {
	case HighCard:
		return "High Card"
	case OnePair:
		return "One Pair"
	case TwoPair:
		return "Two Pair"
	case ThreeOfAKind:
		return "Three of a Kind"
	case Straight:
		return "Straight"
	case Flush:
		return "Flush"
	case FullHouse:
		return "Full House"
	case FourOfAKind:
		return "Four of a Kind"
	case StraightFlush:
		return "Straight Flush"
	case RoyalFlush:
		return "Royal Flush"
	default:
		return "Unknown"
	}
}

// EvaluateBestHand finds the best 5-card hand from hole cards and community cards
func EvaluateBestHand(holeCards, communityCards []Card) (int32, string) {
	allCards := append(holeCards, communityCards...)
	
	if len(allCards) < 5 {
		// Not enough cards to make a hand
		return 999999, "Invalid Hand"
	}

	// Generate all possible 5-card combinations
	combinations := generateCombinations(allCards, 5)
	
	bestRank := int32(999999)
	bestHandName := "High Card"

	for _, combo := range combinations {
		rank, handName := evaluateFiveCardHand(combo)
		if rank < bestRank {
			bestRank = rank
			bestHandName = handName
		}
	}

	return bestRank, bestHandName
}

// evaluateFiveCardHand evaluates a specific 5-card hand
func evaluateFiveCardHand(cards []Card) (int32, string) {
	if len(cards) != 5 {
		return 999999, "Invalid Hand"
	}

	// Sort cards by value (descending)
	sorted := make([]Card, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	isFlush := checkFlush(sorted)
	isStraight, straightHigh := checkStraight(sorted)
	valueCounts := getValueCounts(sorted)

	// Royal Flush
	if isFlush && isStraight && straightHigh == 14 {
		return int32(RoyalFlush)*1000000 + int32(straightHigh), "Royal Flush"
	}

	// Straight Flush
	if isFlush && isStraight {
		return int32(StraightFlush)*1000000 + int32(straightHigh), "Straight Flush"
	}

	// Four of a Kind
	if valueCounts[0].count == 4 {
		rank := int32(FourOfAKind)*1000000 + int32(valueCounts[0].value)*1000 + int32(valueCounts[1].value)
		return rank, "Four of a Kind"
	}

	// Full House
	if valueCounts[0].count == 3 && valueCounts[1].count == 2 {
		rank := int32(FullHouse)*1000000 + int32(valueCounts[0].value)*1000 + int32(valueCounts[1].value)
		return rank, "Full House"
	}

	// Flush
	if isFlush {
		rank := int32(Flush) * 1000000
		for i, card := range sorted {
			rank += int32(card.Value) * int32(1000/(i+1))
		}
		return rank, "Flush"
	}

	// Straight
	if isStraight {
		return int32(Straight)*1000000 + int32(straightHigh), "Straight"
	}

	// Three of a Kind
	if valueCounts[0].count == 3 {
		rank := int32(ThreeOfAKind)*1000000 + int32(valueCounts[0].value)*1000 + 
			int32(valueCounts[1].value)*10 + int32(valueCounts[2].value)
		return rank, "Three of a Kind"
	}

	// Two Pair
	if valueCounts[0].count == 2 && valueCounts[1].count == 2 {
		rank := int32(TwoPair)*1000000 + int32(valueCounts[0].value)*1000 + 
			int32(valueCounts[1].value)*10 + int32(valueCounts[2].value)
		return rank, "Two Pair"
	}

	// One Pair
	if valueCounts[0].count == 2 {
		rank := int32(OnePair)*1000000 + int32(valueCounts[0].value)*1000 + 
			int32(valueCounts[1].value)*100 + int32(valueCounts[2].value)*10 + int32(valueCounts[3].value)
		return rank, "One Pair"
	}

	// High Card
	rank := int32(HighCard) * 1000000
	for i, card := range sorted {
		rank += int32(card.Value) * int32(10000/(i+1))
	}
	return rank, "High Card"
}

type valueCount struct {
	value int
	count int
}

func getValueCounts(cards []Card) []valueCount {
	counts := make(map[int]int)
	for _, card := range cards {
		counts[card.Value]++
	}

	result := make([]valueCount, 0, len(counts))
	for value, count := range counts {
		result = append(result, valueCount{value: value, count: count})
	}

	// Sort by count (descending), then by value (descending)
	sort.Slice(result, func(i, j int) bool {
		if result[i].count != result[j].count {
			return result[i].count > result[j].count
		}
		return result[i].value > result[j].value
	})

	return result
}

func checkFlush(cards []Card) bool {
	suit := cards[0].Suit
	for _, card := range cards[1:] {
		if card.Suit != suit {
			return false
		}
	}
	return true
}

func checkStraight(cards []Card) (bool, int) {
	// Check for regular straight
	for i := 0; i < len(cards)-1; i++ {
		if cards[i].Value-cards[i+1].Value != 1 {
			// Check for A-2-3-4-5 straight (wheel)
			if i == 0 && cards[0].Value == 14 {
				if cards[1].Value == 5 && cards[2].Value == 4 && 
					cards[3].Value == 3 && cards[4].Value == 2 {
					return true, 5 // Ace-low straight
				}
			}
			return false, 0
		}
	}
	return true, cards[0].Value
}

func generateCombinations(cards []Card, size int) [][]Card {
	var result [][]Card
	var current []Card
	
	var generate func(start int)
	generate = func(start int) {
		if len(current) == size {
			combo := make([]Card, size)
			copy(combo, current)
			result = append(result, combo)
			return
		}
		
		for i := start; i < len(cards); i++ {
			current = append(current, cards[i])
			generate(i + 1)
			current = current[:len(current)-1]
		}
	}
	
	generate(0)
	return result
}
