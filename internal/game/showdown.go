package game

import (
	"fmt"
	"math/big"

	"github.com/RedPaladin7/peerpoker/internal/blockchain"
	"github.com/RedPaladin7/peerpoker/internal/deck"
	"github.com/ethereum/go-ethereum/common"
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
		winAmount := g.currentPot
		g.playerStates[winnerAddr].Stack += winAmount

		logrus.Infof("🏆 WINNER BY DEFAULT: %s wins %d chips (everyone else folded)!",
			winnerAddr, winAmount)

		// Blockchain: Distribute winnings on-chain
		if g.blockchainEnabled && g.blockchainGameID != [32]byte{} {
			g.distributeWinningsOnChain([]string{winnerAddr}, []int{winAmount})
		}

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

	// Track all winners and amounts for blockchain
	allWinners := []string{}
	allAmounts := []int{}

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

				// Collect for blockchain payout
				for _, winner := range potWinners {
					share := pot.Amount / len(potWinners)
					allWinners = append(allWinners, winner.Addr)
					allAmounts = append(allAmounts, share)
				}
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

			// Collect for blockchain payout
			for _, winner := range winners {
				share := g.currentPot / len(winners)
				allWinners = append(allWinners, winner.Addr)
				allAmounts = append(allAmounts, share)
			}
		}
	}

	// Blockchain: Distribute all winnings on-chain
	if g.blockchainEnabled && g.blockchainGameID != [32]byte{} && len(allWinners) > 0 {
		g.distributeWinningsOnChain(allWinners, allAmounts)
	}

	g.resetHandState()
}

// distributeWinningsOnChain sends payout transaction to smart contract
func (g *Game) distributeWinningsOnChain(winners []string, amounts []int) {
	winnerAddrs := make([]common.Address, len(winners))
	winnerAmounts := make([]*big.Int, len(amounts))

	for i := range winners {
		winnerAddrs[i] = common.HexToAddress(winners[i])
		winnerAmounts[i] = big.NewInt(int64(amounts[i]))
	}

	logrus.WithFields(logrus.Fields{
		"game_id": fmt.Sprintf("0x%x", g.blockchainGameID),
		"winners": len(winners),
	}).Info("Distributing winnings on blockchain...")

	err := g.blockchain.EndGame(g.blockchainGameID, winnerAddrs, winnerAmounts)
	if err != nil {
		logrus.Errorf("Failed to distribute winnings on blockchain: %v", err)
		logrus.Warn("Winnings distributed in-game only (blockchain transaction failed)")
	} else {
		logrus.WithFields(logrus.Fields{
			"game_id": fmt.Sprintf("0x%x", g.blockchainGameID),
			"winners": len(winners),
		}).Info("✅ Winnings distributed on blockchain successfully")
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

// resetHandState resets the game state for a new hand
func (g *Game) resetHandState() {
	logrus.Info("=== Resetting for new hand ===")

	g.currentPot = 0
	g.highestBet = 0
	g.lastRaiseAmount = BigBlind
	g.myHand = make([]deck.Card, 0, 2)
	g.communityCards = make([]deck.Card, 0, 5)
	g.currentDeck = nil
	g.sidePots = []SidePot{}
	g.revealedKeys = make(map[string]*crypto.CardKeys)
	g.foldedPlayerKeys = make(map[string]*crypto.CardKeys)

	// Reset blockchain game ID for next hand
	if g.blockchainEnabled {
		g.blockchainGameID = [32]byte{}
	}

	// Remove players with no chips
	for addr, state := range g.playerStates {
		if state.Stack <= 0 {
			state.IsActive = false
			logrus.Infof("Player %s eliminated (no chips)", addr)
		}
	}

	// Check if we have enough players
	if len(g.getReadyActivePlayers()) >= 2 {
		g.setStatus(GameStatusWaiting)
		// Auto-start next hand after a delay if all players are still ready
		// For now, we'll wait for ready signals again
	} else {
		g.setStatus(GameStatusWaiting)
		logrus.Info("Not enough players, waiting for more")
	}
}
