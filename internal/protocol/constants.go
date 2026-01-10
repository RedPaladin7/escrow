package protocol

// Game variants
const (
	GameVariantTexasHoldem = "TEXAS_HOLDEM"
	GameVariantOmaha       = "OMAHA"
	GameVariantSevenCard   = "SEVEN_CARD_STUD"
)

// Error codes
const (
	ErrCodeInvalidMessage    = "INVALID_MESSAGE"
	ErrCodeInvalidAction     = "INVALID_ACTION"
	ErrCodeNotYourTurn       = "NOT_YOUR_TURN"
	ErrCodeInsufficientFunds = "INSUFFICIENT_FUNDS"
	ErrCodeGameNotStarted    = "GAME_NOT_STARTED"
	ErrCodePlayerNotFound    = "PLAYER_NOT_FOUND"
	ErrCodeAlreadyInGame     = "ALREADY_IN_GAME"
	ErrCodeGameFull          = "GAME_FULL"
	ErrCodeInternalError     = "INTERNAL_ERROR"
)

// Action types
const (
	ActionFold  = "fold"
	ActionCheck = "check"
	ActionCall  = "call"
	ActionBet   = "bet"
	ActionRaise = "raise"
	ActionAllIn = "all_in"
)

// Game states
const (
	StateWaiting  = "WAITING"
	StateDealing  = "DEALING"
	StatePreFlop  = "PRE_FLOP"
	StateFlop     = "FLOP"
	StateTurn     = "TURN"
	StateRiver    = "RIVER"
	StateShowdown = "SHOWDOWN"
)

// Default values
const (
	DefaultSmallBlind = 10
	DefaultBigBlind   = 20
	DefaultStack      = 1000
	DefaultMaxPlayers = 6
)
