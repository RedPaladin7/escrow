package game

import (
	"sort"

	"github.com/sirupsen/logrus"
)

// calculateSidePots calculates all side pots for the hand
func (g *Game) calculateSidePots() []SidePot {
	type PlayerContribution struct {
		Addr   string
		Amount int
	}

	contributions := []PlayerContribution{}
	for addr, state := range g.playerStates {
		if state.IsActive && state.TotalBetThisHand > 0 {
			contributions = append(contributions, PlayerContribution{
				Addr:   addr,
				Amount: state.TotalBetThisHand,
			})
		}
	}

	// Sort by contribution amount
	sort.Slice(contributions, func(i, j int) bool {
		return contributions[i].Amount < contributions[j].Amount
	})

	pots := []SidePot{}
	previousCap := 0

	for i, contrib := range contributions {
		if contrib.Amount <= previousCap {
			continue
		}

		cap := contrib.Amount
		eligible := []string{}
		potAmount := 0

		// All players from this position onwards are eligible
		for j := i; j < len(contributions); j++ {
			eligible = append(eligible, contributions[j].Addr)
			potAmount += (cap - previousCap)
		}

		if potAmount > 0 {
			pots = append(pots, SidePot{
				Amount:          potAmount,
				Cap:             cap,
				EligiblePlayers: eligible,
			})
		}

		previousCap = cap
	}

	return pots
}

// distributePot distributes a pot among winners
func (g *Game) distributePot(amount int, winners []*PlayerHand, potNum int) {
	share := amount / len(winners)
	remainder := amount % len(winners)

	for i, winner := range winners {
		winAmount := share
		if i == 0 {
			winAmount += remainder // Give remainder to first winner
		}

		state := g.playerStates[winner.Addr]
		state.Stack += winAmount

		logrus.WithFields(logrus.Fields{
			"pot":       potNum,
			"player":    winner.Addr,
			"hand":      winner.HandName,
			"rank":      winner.Rank,
			"win_amount": winAmount,
			"new_stack": state.Stack,
		}).Info("Pot distributed")
	}
}
