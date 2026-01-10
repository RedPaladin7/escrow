// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./PotManager.sol";
import "./PlayerRegistry.sol";

/**
 * @title PokerTable
 * @dev Main contract for managing a poker game table with buy-ins and payouts
 */
contract PokerTable {
    // Events
    event GameCreated(bytes32 indexed gameId, address indexed creator, uint256 buyIn, uint256 maxPlayers);
    event PlayerJoined(bytes32 indexed gameId, address indexed player, uint256 amount);
    event PlayerLeft(bytes32 indexed gameId, address indexed player, uint256 refund);
    event GameStarted(bytes32 indexed gameId, uint256 totalPot);
    event GameEnded(bytes32 indexed gameId, address[] winners, uint256[] payouts);
    event FundsLocked(bytes32 indexed gameId, address indexed player, uint256 amount);
    event FundsReleased(bytes32 indexed gameId, address indexed player, uint256 amount);

    // Structs
    struct Game {
        bytes32 gameId;
        address creator;
        uint256 buyIn;
        uint256 smallBlind;
        uint256 bigBlind;
        uint256 maxPlayers;
        uint256 totalPot;
        address[] players;
        mapping(address => uint256) playerBalances;
        mapping(address => bool) hasJoined;
        GameStatus status;
        uint256 createdAt;
        uint256 startedAt;
        uint256 endedAt;
    }

    enum GameStatus {
        Waiting,
        Active,
        Ended,
        Cancelled
    }

    // State variables
    mapping(bytes32 => Game) public games;
    mapping(address => bytes32[]) public playerGames;
    bytes32[] public activeGameIds;
    
    PotManager public potManager;
    PlayerRegistry public playerRegistry;

    uint256 public constant MIN_BUY_IN = 0.001 ether;
    uint256 public constant MAX_BUY_IN = 100 ether;
    uint256 public constant PLATFORM_FEE_PERCENT = 2; // 2% platform fee
    address public platformFeeAddress;

    // Modifiers
    modifier onlyGameCreator(bytes32 _gameId) {
        require(games[_gameId].creator == msg.sender, "Not game creator");
        _;
    }

    modifier gameExists(bytes32 _gameId) {
        require(games[_gameId].creator != address(0), "Game does not exist");
        _;
    }

    modifier gameInStatus(bytes32 _gameId, GameStatus _status) {
        require(games[_gameId].status == _status, "Invalid game status");
        _;
    }

    constructor(address _potManager, address _playerRegistry, address _platformFeeAddress) {
        potManager = PotManager(_potManager);
        playerRegistry = PlayerRegistry(_playerRegistry);
        platformFeeAddress = _platformFeeAddress;
    }

    /**
     * @dev Create a new poker game
     */
    function createGame(
        uint256 _buyIn,
        uint256 _smallBlind,
        uint256 _bigBlind,
        uint256 _maxPlayers
    ) external returns (bytes32) {
        require(_buyIn >= MIN_BUY_IN && _buyIn <= MAX_BUY_IN, "Invalid buy-in amount");
        require(_maxPlayers >= 2 && _maxPlayers <= 10, "Invalid max players");
        require(_smallBlind > 0 && _bigBlind > _smallBlind, "Invalid blinds");

        bytes32 gameId = keccak256(abi.encodePacked(msg.sender, block.timestamp, _buyIn));
        
        Game storage game = games[gameId];
        game.gameId = gameId;
        game.creator = msg.sender;
        game.buyIn = _buyIn;
        game.smallBlind = _smallBlind;
        game.bigBlind = _bigBlind;
        game.maxPlayers = _maxPlayers;
        game.status = GameStatus.Waiting;
        game.createdAt = block.timestamp;

        activeGameIds.push(gameId);
        playerRegistry.registerPlayer(msg.sender);

        emit GameCreated(gameId, msg.sender, _buyIn, _maxPlayers);
        return gameId;
    }

    /**
     * @dev Join a game with the required buy-in
     */
    function joinGame(bytes32 _gameId) external payable gameExists(_gameId) gameInStatus(_gameId, GameStatus.Waiting) {
        Game storage game = games[_gameId];
        
        require(!game.hasJoined[msg.sender], "Already joined");
        require(game.players.length < game.maxPlayers, "Game is full");
        require(msg.value == game.buyIn, "Incorrect buy-in amount");

        game.players.push(msg.sender);
        game.hasJoined[msg.sender] = true;
        game.playerBalances[msg.sender] = msg.value;
        game.totalPot += msg.value;

        playerGames[msg.sender].push(_gameId);
        playerRegistry.registerPlayer(msg.sender);

        // Lock funds in PotManager
        potManager.lockFunds{value: msg.value}(_gameId, msg.sender);

        emit PlayerJoined(_gameId, msg.sender, msg.value);
        emit FundsLocked(_gameId, msg.sender, msg.value);
    }

    /**
     * @dev Leave a game before it starts (get refund)
     */
    function leaveGame(bytes32 _gameId) external gameExists(_gameId) gameInStatus(_gameId, GameStatus.Waiting) {
        Game storage game = games[_gameId];
        
        require(game.hasJoined[msg.sender], "Not in game");

        uint256 refundAmount = game.playerBalances[msg.sender];
        game.playerBalances[msg.sender] = 0;
        game.hasJoined[msg.sender] = false;
        game.totalPot -= refundAmount;

        // Remove player from array
        for (uint256 i = 0; i < game.players.length; i++) {
            if (game.players[i] == msg.sender) {
                game.players[i] = game.players[game.players.length - 1];
                game.players.pop();
                break;
            }
        }

        // Release funds from PotManager
        potManager.releaseFunds(_gameId, msg.sender, refundAmount);

        emit PlayerLeft(_gameId, msg.sender, refundAmount);
        emit FundsReleased(_gameId, msg.sender, refundAmount);
    }

    /**
     * @dev Start the game (only creator can start)
     */
    function startGame(bytes32 _gameId) external onlyGameCreator(_gameId) gameExists(_gameId) gameInStatus(_gameId, GameStatus.Waiting) {
        Game storage game = games[_gameId];
        
        require(game.players.length >= 2, "Need at least 2 players");

        game.status = GameStatus.Active;
        game.startedAt = block.timestamp;

        emit GameStarted(_gameId, game.totalPot);
    }

    /**
     * @dev End game and distribute winnings
     */
    function endGame(
        bytes32 _gameId,
        address[] calldata _winners,
        uint256[] calldata _amounts
    ) external onlyGameCreator(_gameId) gameExists(_gameId) gameInStatus(_gameId, GameStatus.Active) {
        Game storage game = games[_gameId];
        
        require(_winners.length == _amounts.length, "Mismatched arrays");
        require(_winners.length > 0, "No winners");

        uint256 totalPayout = 0;
        for (uint256 i = 0; i < _amounts.length; i++) {
            totalPayout += _amounts[i];
        }

        // Calculate platform fee
        uint256 platformFee = (game.totalPot * PLATFORM_FEE_PERCENT) / 100;
        uint256 availablePot = game.totalPot - platformFee;

        require(totalPayout <= availablePot, "Payout exceeds pot");

        game.status = GameStatus.Ended;
        game.endedAt = block.timestamp;

        // Distribute winnings through PotManager
        potManager.distributePot(_gameId, _winners, _amounts);

        // Transfer platform fee
        if (platformFee > 0) {
            payable(platformFeeAddress).transfer(platformFee);
        }

        emit GameEnded(_gameId, _winners, _amounts);
    }

    /**
     * @dev Cancel game and refund all players (only if not started)
     */
    function cancelGame(bytes32 _gameId) external onlyGameCreator(_gameId) gameExists(_gameId) gameInStatus(_gameId, GameStatus.Waiting) {
        Game storage game = games[_gameId];
        
        for (uint256 i = 0; i < game.players.length; i++) {
            address player = game.players[i];
            uint256 refund = game.playerBalances[player];
            
            if (refund > 0) {
                game.playerBalances[player] = 0;
                potManager.releaseFunds(_gameId, player, refund);
            }
        }

        game.status = GameStatus.Cancelled;
        emit GameEnded(_gameId, new address[](0), new uint256[](0));
    }

    /**
     * @dev Get game details
     */
    function getGame(bytes32 _gameId) external view returns (
        address creator,
        uint256 buyIn,
        uint256 smallBlind,
        uint256 bigBlind,
        uint256 maxPlayers,
        uint256 totalPot,
        uint256 playerCount,
        GameStatus status
    ) {
        Game storage game = games[_gameId];
        return (
            game.creator,
            game.buyIn,
            game.smallBlind,
            game.bigBlind,
            game.maxPlayers,
            game.totalPot,
            game.players.length,
            game.status
        );
    }

    /**
     * @dev Get all players in a game
     */
    function getGamePlayers(bytes32 _gameId) external view returns (address[] memory) {
        return games[_gameId].players;
    }

    /**
     * @dev Get player balance in a game
     */
    function getPlayerBalance(bytes32 _gameId, address _player) external view returns (uint256) {
        return games[_gameId].playerBalances[_player];
    }

    /**
     * @dev Get all active games
     */
    function getActiveGames() external view returns (bytes32[] memory) {
        return activeGameIds;
    }

    /**
     * @dev Check if player is in game
     */
    function isPlayerInGame(bytes32 _gameId, address _player) external view returns (bool) {
        return games[_gameId].hasJoined[_player];
    }
}
