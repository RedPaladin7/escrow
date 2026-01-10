// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title PotManager
 * @dev Manages poker pot funds, locking, and distribution
 */
contract PotManager {
    // Events
    event FundsLocked(bytes32 indexed gameId, address indexed player, uint256 amount);
    event FundsReleased(bytes32 indexed gameId, address indexed player, uint256 amount);
    event PotDistributed(bytes32 indexed gameId, address[] winners, uint256[] amounts);
    event SidePotCreated(bytes32 indexed gameId, uint256 potIndex, uint256 amount);

    // Structs
    struct LockedFunds {
        uint256 amount;
        uint256 lockedAt;
        bool isLocked;
    }

    struct SidePot {
        uint256 amount;
        uint256 cap;
        address[] eligiblePlayers;
    }

    // State variables
    mapping(bytes32 => mapping(address => LockedFunds)) public lockedFunds;
    mapping(bytes32 => uint256) public gamePots;
    mapping(bytes32 => SidePot[]) public sidePots;
    
    address public pokerTable;
    
    modifier onlyPokerTable() {
        require(msg.sender == pokerTable, "Only poker table can call");
        _;
    }

    constructor(address _pokerTable) {
        pokerTable = _pokerTable;
    }

    /**
     * @dev Lock funds for a player in a game
     */
    function lockFunds(bytes32 _gameId, address _player) external payable onlyPokerTable {
        require(msg.value > 0, "Amount must be greater than 0");
        require(!lockedFunds[_gameId][_player].isLocked, "Funds already locked");

        lockedFunds[_gameId][_player] = LockedFunds({
            amount: msg.value,
            lockedAt: block.timestamp,
            isLocked: true
        });

        gamePots[_gameId] += msg.value;

        emit FundsLocked(_gameId, _player, msg.value);
    }

    /**
     * @dev Release funds to a player
     */
    function releaseFunds(bytes32 _gameId, address _player, uint256 _amount) external onlyPokerTable {
        require(lockedFunds[_gameId][_player].isLocked, "No locked funds");
        require(lockedFunds[_gameId][_player].amount >= _amount, "Insufficient locked funds");

        lockedFunds[_gameId][_player].amount -= _amount;
        
        if (lockedFunds[_gameId][_player].amount == 0) {
            lockedFunds[_gameId][_player].isLocked = false;
        }

        gamePots[_gameId] -= _amount;

        payable(_player).transfer(_amount);

        emit FundsReleased(_gameId, _player, _amount);
    }

    /**
     * @dev Distribute pot to winners
     */
    function distributePot(
        bytes32 _gameId,
        address[] calldata _winners,
        uint256[] calldata _amounts
    ) external onlyPokerTable {
        require(_winners.length == _amounts.length, "Mismatched arrays");

        for (uint256 i = 0; i < _winners.length; i++) {
            require(_amounts[i] > 0, "Invalid amount");
            require(_amounts[i] <= gamePots[_gameId], "Amount exceeds pot");

            gamePots[_gameId] -= _amounts[i];
            payable(_winners[i]).transfer(_amounts[i]);
        }

        emit PotDistributed(_gameId, _winners, _amounts);
    }

    /**
     * @dev Create a side pot for all-in situations
     */
    function createSidePot(
        bytes32 _gameId,
        uint256 _amount,
        uint256 _cap,
        address[] calldata _eligiblePlayers
    ) external onlyPokerTable {
        SidePot memory newPot = SidePot({
            amount: _amount,
            cap: _cap,
            eligiblePlayers: _eligiblePlayers
        });

        sidePots[_gameId].push(newPot);

        emit SidePotCreated(_gameId, sidePots[_gameId].length - 1, _amount);
    }

    /**
     * @dev Get locked funds for a player
     */
    function getLockedFunds(bytes32 _gameId, address _player) external view returns (uint256, bool) {
        LockedFunds memory funds = lockedFunds[_gameId][_player];
        return (funds.amount, funds.isLocked);
    }

    /**
     * @dev Get game pot total
     */
    function getGamePot(bytes32 _gameId) external view returns (uint256) {
        return gamePots[_gameId];
    }

    /**
     * @dev Get side pots for a game
     */
    function getSidePots(bytes32 _gameId) external view returns (SidePot[] memory) {
        return sidePots[_gameId];
    }

    /**
     * @dev Get number of side pots
     */
    function getSidePotCount(bytes32 _gameId) external view returns (uint256) {
        return sidePots[_gameId].length;
    }

    /**
     * @dev Emergency withdrawal (only owner)
     */
    function emergencyWithdraw(bytes32 _gameId, address _to) external onlyPokerTable {
        uint256 amount = gamePots[_gameId];
        require(amount > 0, "No funds to withdraw");

        gamePots[_gameId] = 0;
        payable(_to).transfer(amount);
    }

    // Fallback and receive functions
    receive() external payable {}
    fallback() external payable {}
}
