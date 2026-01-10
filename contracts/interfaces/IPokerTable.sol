// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

interface IPokerTable {
    enum GameStatus {
        Waiting,
        Active,
        Ended,
        Cancelled
    }

    function createGame(
        uint256 _buyIn,
        uint256 _smallBlind,
        uint256 _bigBlind,
        uint256 _maxPlayers
    ) external returns (bytes32);

    function joinGame(bytes32 _gameId) external payable;
    
    function leaveGame(bytes32 _gameId) external;
    
    function startGame(bytes32 _gameId) external;
    
    function endGame(
        bytes32 _gameId,
        address[] calldata _winners,
        uint256[] calldata _amounts
    ) external;
    
    function getGame(bytes32 _gameId) external view returns (
        address creator,
        uint256 buyIn,
        uint256 smallBlind,
        uint256 bigBlind,
        uint256 maxPlayers,
        uint256 totalPot,
        uint256 playerCount,
        GameStatus status
    );
    
    function isPlayerInGame(bytes32 _gameId, address _player) external view returns (bool);
}
