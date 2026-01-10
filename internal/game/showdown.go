package game

import (
	"github.com/RedPaladin7/peerpoker/internal/deck"
	"github.com/sirupsen/logrus"
)

type PlayerHand struct {
	Addr     string
	Hand     []deck.Card
	Rank     int32
	HandName string
}

// ResolveWinner determines the winner(s) and distributes pots
func (g *Game) ResolveWinner() {
	logrus.Info("=== RESOLVING WINNER ===")

	activePlayers := g.getReadyActivePlayers()
	nonFoldedPlayers := []string{}

	for _, playerAddr := range activePlayers {
		state := g.playerStates[playerAddr]
		if !state.IsFolded {
			nonFoldedPlayers = append(nonFoldedPlayers, playerAddr)
		}
	}

	// Only one player left (everyone else folded)
	if len(nonFoldedPlayers) == 1 {
		winnerAddr := nonFoldedPlayers[0]
		g.playerStates[winnerAddr].Stack += g.currentPot
		logrus.Infof("🏆 WINNER BY DEFAULT: %s wins %d chips (everyone else folded)!",
			winnerAddr, g.currentPot)
		g.resetHandState()
		return
	}

	// Multiple players - need to evaluate hands
	playerHands := make([]PlayerHand, 0, len(nonFoldedPlayers))

	for _, playerAddr := range nonFoldedPlayers {
		state := g.playerStates[playerAddr]

		// TODO: Decrypt player's hole cards using revealed keys
		// For now, we'll use placeholder logic
		playerHand := []deck.Card{
			// These would be decrypted from the encrypted deck
		}

		// Evaluate hand
		rank, handName := deck.EvaluateBestHand(playerHand, g.communityCards)

		logrus.Infof("Player %s: %v - %s (Rank: %d)",
			playerAddr, playerHand, handName, rank)

		playerHands = append(playerHands, PlayerHand{
			Addr:     playerAddr,
			Hand:     playerHand,
			Rank:     rank,
			HandName: handName,
		})
	}

	// Calculate side pots
	sidePots := g.calculateSidePots()

	if len(sidePots) > 0 {
		logrus.Infof("Distributing %d pot(s)...", len(sidePots))

		for i, pot := range sidePots {
			logrus.Infof("Pot #%d: %d chips (cap: %d)", i+1, pot.Amount, pot.Cap)

			// Find best hand among eligible players
			bestRank := int32(999999)
			potWinners := []*PlayerHand{}

			for idx := range playerHands {
				ph := &playerHands[idx]

				// Check if player is eligible for this pot
				isEligible := false
				for _, eligibleAddr := range pot.EligiblePlayers {
					if ph.Addr == eligibleAddr {
						isEligible = true
						break
					}
				}

				if isEligible {
					if ph.Rank < bestRank {
						bestRank = ph.Rank
						potWinners = []*PlayerHand{ph}
					} else if ph.Rank == bestRank {
						potWinners = append(potWinners, ph)
					}
				}
			}

			if len(potWinners) > 0 {
				g.distributePot(pot.Amount, potWinners, i+1)
			}
		}
	} else {
		// Single main pot
		bestRank := int32(999999)
		winners := []*PlayerHand{}

		for idx := range playerHands {
			ph := &playerHands[idx]
			if ph.Rank < bestRank {
				bestRank = ph.Rank
				winners = []*PlayerHand{ph}
			} else if ph.Rank == bestRank {
				winners = append(winners, ph)
			}
		}

		if len(winners) > 0 {
			g.distributePot(g.currentPot, winners, 0)
		}
	}
}

// InitiateShuffleAndDeal starts the mental poker protocol
func (g *Game) InitiateShuffleAndDeal() {
	logrus.Info("Initiating shuffle and deal protocol...")
	// TODO: Implement mental poker shuffle and deal
	// This would involve:
	// 1. Creating encrypted deck
	// 2. Each player shuffles and encrypts
	// 3. Distribute encrypted cards
	// 4. Players decrypt their own cards
}
