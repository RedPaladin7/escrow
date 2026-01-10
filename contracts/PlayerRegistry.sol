// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title PlayerRegistry
 * @dev Registry for tracking player statistics and reputation
 */
contract PlayerRegistry {
    // Events
    event PlayerRegistered(address indexed player, uint256 timestamp);
    event StatsUpdated(address indexed player, uint256 gamesPlayed, uint256 totalWinnings);
    event PlayerBanned(address indexed player, string reason);
    event PlayerUnbanned(address indexed player);

    // Structs
    struct PlayerStats {
        uint256 gamesPlayed;
        uint256 gamesWon;
        uint256 totalWinnings;
        uint256 totalLosses;
        uint256 registeredAt;
        uint256 lastPlayedAt;
        bool isBanned;
        string banReason;
    }

    // State variables
    mapping(address => PlayerStats) public playerStats;
    mapping(address => bool) public isRegistered;
    address[] public registeredPlayers;
    
    address public owner;
    address public pokerTable;

    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner");
        _;
    }

    modifier onlyPokerTable() {
        require(msg.sender == pokerTable, "Only poker table");
        _;
    }

    constructor() {
        owner = msg.sender;
    }

    /**
     * @dev Set poker table address
     */
    function setPokerTable(address _pokerTable) external onlyOwner {
        pokerTable = _pokerTable;
    }

    /**
     * @dev Register a new player
     */
    function registerPlayer(address _player) external onlyPokerTable {
        if (!isRegistered[_player]) {
            playerStats[_player] = PlayerStats({
                gamesPlayed: 0,
                gamesWon: 0,
                totalWinnings: 0,
                totalLosses: 0,
                registeredAt: block.timestamp,
                lastPlayedAt: 0,
                isBanned: false,
                banReason: ""
            });

            isRegistered[_player] = true;
            registeredPlayers.push(_player);

            emit PlayerRegistered(_player, block.timestamp);
        }
    }

    /**
     * @dev Update player stats after game
     */
    function updateStats(
        address _player,
        bool _won,
        uint256 _winnings,
        uint256 _losses
    ) external onlyPokerTable {
        require(isRegistered[_player], "Player not registered");

        PlayerStats storage stats = playerStats[_player];
        stats.gamesPlayed++;
        stats.lastPlayedAt = block.timestamp;

        if (_won) {
            stats.gamesWon++;
            stats.totalWinnings += _winnings;
        }

        stats.totalLosses += _losses;

        emit StatsUpdated(_player, stats.gamesPlayed, stats.totalWinnings);
    }

    /**
     * @dev Ban a player
     */
    function banPlayer(address _player, string calldata _reason) external onlyOwner {
        require(isRegistered[_player], "Player not registered");
        
        playerStats[_player].isBanned = true;
        playerStats[_player].banReason = _reason;

        emit PlayerBanned(_player, _reason);
    }

    /**
     * @dev Unban a player
     */
    function unbanPlayer(address _player) external onlyOwner {
        require(isRegistered[_player], "Player not registered");
        
        playerStats[_player].isBanned = false;
        playerStats[_player].banReason = "";

        emit PlayerUnbanned(_player);
    }

    /**
     * @dev Get player stats
     */
    function getPlayerStats(address _player) external view returns (
        uint256 gamesPlayed,
        uint256 gamesWon,
        uint256 totalWinnings,
        uint256 totalLosses,
        uint256 winRate,
        bool isBanned
    ) {
        PlayerStats memory stats = playerStats[_player];
        uint256 rate = stats.gamesPlayed > 0 ? (stats.gamesWon * 100) / stats.gamesPlayed : 0;
        
        return (
            stats.gamesPlayed,
            stats.gamesWon,
            stats.totalWinnings,
            stats.totalLosses,
            rate,
            stats.isBanned
        );
    }

    /**
     * @dev Check if player is banned
     */
    function isPlayerBanned(address _player) external view returns (bool) {
        return playerStats[_player].isBanned;
    }

    /**
     * @dev Get total registered players
     */
    function getTotalPlayers() external view returns (uint256) {
        return registeredPlayers.length;
    }

    /**
     * @dev Get all registered players
     */
    function getAllPlayers() external view returns (address[] memory) {
        return registeredPlayers;
    }
}
