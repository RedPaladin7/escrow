package game

import (
	"fmt"
	"math/big"

	"github.com/RedPaladin7/peerpoker/internal/blockchain"
	"github.com/RedPaladin7/peerpoker/internal/crypto"
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

	// Multiple players - evaluate hands
	playerHands := make([]PlayerHand, 0, len(nonFoldedPlayers))

	for _, playerAddr := range nonFoldedPlayers {
		// Decrypt player's hole cards using revealed keys
		holeCards := g.decryptPlayerCards(playerAddr)

		// Evaluate hand
		rank, handName := deck.EvaluateBestHand(holeCards, g.communityCards)

		logrus.Infof("Player %s: %v - %s (Rank: %d)",
			playerAddr, holeCards, handName, rank)

		playerHands = append(playerHands, PlayerHand{
			Addr:     playerAddr,
			Hand:     holeCards,
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

// decryptPlayerCards decrypts a player's hole cards using all revealed keys
func (g *Game) decryptPlayerCards(playerAddr string) []deck.Card {
	// Get player's card indices (first two cards for this player)
	state := g.playerStates[playerAddr]
	cardIndices := []int{state.RotationID * 2, state.RotationID*2 + 1}

	cards := make([]deck.Card, 0, 2)

	for _, idx := range cardIndices {
		if idx >= len(g.currentDeck) {
			logrus.Warnf("Card index %d out of bounds", idx)
			continue
		}

		encryptedCard := g.currentDeck[idx]

		// Decrypt using all revealed keys (from folded players and this player)
		decryptedCard := encryptedCard

		// Apply folded player keys
		for _, keys := range g.foldedPlayerKeys {
			decryptedCard = keys.Decrypt(decryptedCard)
		}

		// Apply revealed keys
		for _, keys := range g.revealedKeys {
			decryptedCard = keys.Decrypt(decryptedCard)
		}

		// Convert decrypted bytes to card
		if len(decryptedCard) > 0 {
			card := deck.NewCardFromByte(decryptedCard[0])
			cards = append(cards, card)
		}
	}

	return cards
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

	// Step 1: Create initial deck
	initialDeck := deck.NewDeck()
	g.currentDeck = initialDeck.ToBytes()

	logrus.Infof("Created initial deck with %d cards", len(g.currentDeck))

	// Step 2: Encrypt deck with our keys
	g.currentDeck = crypto.EncryptDeck(g.currentDeck, g.deckKeys)
	logrus.Info("Encrypted deck with our keys")

	// Step 3: Shuffle the deck
	g.currentDeck = crypto.ShuffleDeck(g.currentDeck)
	logrus.Info("Shuffled deck")

	// Step 4: In a real P2P game, each player would:
	// - Receive the deck
	// - Encrypt with their keys
	// - Shuffle
	// - Pass to next player
	// For now, we simulate this with multiple encryption rounds

	activePlayers := g.getReadyActivePlayers()
	for i, playerAddr := range activePlayers {
		if playerAddr == g.listenAddr {
			continue // Skip self, already encrypted
		}

		// Simulate other players encrypting (in real implementation, this happens via P2P)
		logrus.Infof("Simulating encryption by player %d (%s)", i, playerAddr)
		
		// Generate temporary keys for this player (in reality, they would use their own)
		tempKeys, _ := crypto.GenerateCardKeys()
		g.currentDeck = crypto.EncryptDeck(g.currentDeck, tempKeys)
		g.currentDeck = crypto.ShuffleDeck(g.currentDeck)
		
		// Store keys for later decryption
		g.revealedKeys[playerAddr] = tempKeys
	}

	logrus.Infof("Deck fully encrypted and shuffled by %d players", len(activePlayers))

	// Step 5: Deal cards (encrypt indices are known to all players)
	g.dealHoleCards()
	
	// Update game status
	g.setStatus(GameStatusPreFlop)
	logrus.Info("Cards dealt, starting pre-flop betting")
}

// dealHoleCards deals 2 cards to each player
func (g *Game) dealHoleCards() {
	activePlayers := g.getReadyActivePlayers()
	
	for i, playerAddr := range activePlayers {
		card1Idx := i * 2
		card2Idx := i*2 + 1

		logrus.Infof("Player %s assigned cards at indices [%d, %d]", playerAddr, card1Idx, card2Idx)

		// If this is us, decrypt our cards
		if playerAddr == g.listenAddr {
			g.myHand = g.decryptPlayerCards(g.listenAddr)
			logrus.Infof("Our hand: %v", g.myHand)
		}
	}

	logrus.Info("Hole cards dealt to all players")
}

// dealCommunityCards deals community cards (flop, turn, or river)
func (g *Game) dealCommunityCards(count int) {
	numPlayers := len(g.getReadyActivePlayers())
	startIdx := numPlayers*2 + len(g.communityCards)

	for i := 0; i < count; i++ {
		cardIdx := startIdx + i
		if cardIdx >= len(g.currentDeck) {
			logrus.Warnf("Not enough cards in deck for community card %d", i)
			continue
		}

		encryptedCard := g.currentDeck[cardIdx]
		decryptedCard := encryptedCard

		// Decrypt using all player keys
		for _, keys := range g.revealedKeys {
			decryptedCard = keys.Decrypt(decryptedCard)
		}

		// Decrypt with our keys
		decryptedCard = g.deckKeys.Decrypt(decryptedCard)

		if len(decryptedCard) > 0 {
			card := deck.NewCardFromByte(decryptedCard[0])
			g.communityCards = append(g.communityCards, card)
			logrus.Infof("Dealt community card: %s", card.String())
		}
	}
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
	} else {
		g.setStatus(GameStatusWaiting)
		logrus.Info("Not enough players, waiting for more")
	}
}
