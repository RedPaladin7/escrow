package game

type GameStatus int

const (
	GameStatusWaiting GameStatus = iota
	GameStatusDealing
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
	GameStatusShowdown
)

func (gs GameStatus) String() string {
	switch gs {
	case GameStatusWaiting:
		return "WAITING"
	case GameStatusDealing:
		return "DEALING"
	case GameStatusPreFlop:
		return "PRE_FLOP"
	case GameStatusFlop:
		return "FLOP"
	case GameStatusTurn:
		return "TURN"
	case GameStatusRiver:
		return "RIVER"
	case GameStatusShowdown:
		return "SHOWDOWN"
	default:
		return "UNKNOWN"
	}
}
