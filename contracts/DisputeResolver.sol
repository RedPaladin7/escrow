// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title DisputeResolver
 * @dev Handles disputes in poker games through voting mechanism
 */
contract DisputeResolver {
    // Events
    event DisputeRaised(bytes32 indexed disputeId, bytes32 indexed gameId, address indexed raiser, string reason);
    event VoteCast(bytes32 indexed disputeId, address indexed voter, bool supportDispute);
    event DisputeResolved(bytes32 indexed disputeId, bool upheld, uint256 votesFor, uint256 votesAgainst);

    // Structs
    struct Dispute {
        bytes32 disputeId;
        bytes32 gameId;
        address raiser;
        string reason;
        uint256 createdAt;
        uint256 votingEndsAt;
        mapping(address => bool) hasVoted;
        mapping(address => bool) votes;
        uint256 votesFor;
        uint256 votesAgainst;
        DisputeStatus status;
        address[] voters;
    }

    enum DisputeStatus {
        Active,
        Resolved,
        Expired
    }

    // State variables
    mapping(bytes32 => Dispute) public disputes;
    bytes32[] public activeDisputes;
    
    uint256 public constant VOTING_PERIOD = 24 hours;
    uint256 public constant MIN_VOTES = 3;

    address public owner;

    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner");
        _;
    }

    constructor() {
        owner = msg.sender;
    }

    /**
     * @dev Raise a dispute
     */
    function raiseDispute(bytes32 _gameId, string calldata _reason) external returns (bytes32) {
        bytes32 disputeId = keccak256(abi.encodePacked(_gameId, msg.sender, block.timestamp));

        Dispute storage dispute = disputes[disputeId];
        dispute.disputeId = disputeId;
        dispute.gameId = _gameId;
        dispute.raiser = msg.sender;
        dispute.reason = _reason;
        dispute.createdAt = block.timestamp;
        dispute.votingEndsAt = block.timestamp + VOTING_PERIOD;
        dispute.status = DisputeStatus.Active;

        activeDisputes.push(disputeId);

        emit DisputeRaised(disputeId, _gameId, msg.sender, _reason);
        return disputeId;
    }

    /**
     * @dev Vote on a dispute
     */
    function vote(bytes32 _disputeId, bool _supportDispute) external {
        Dispute storage dispute = disputes[_disputeId];
        
        require(dispute.status == DisputeStatus.Active, "Dispute not active");
        require(block.timestamp < dispute.votingEndsAt, "Voting period ended");
        require(!dispute.hasVoted[msg.sender], "Already voted");

        dispute.hasVoted[msg.sender] = true;
        dispute.votes[msg.sender] = _supportDispute;
        dispute.voters.push(msg.sender);

        if (_supportDispute) {
            dispute.votesFor++;
        } else {
            dispute.votesAgainst++;
        }

        emit VoteCast(_disputeId, msg.sender, _supportDispute);

        // Auto-resolve if minimum votes reached
        if (dispute.votesFor + dispute.votesAgainst >= MIN_VOTES) {
            _resolveDispute(_disputeId);
        }
    }

    /**
     * @dev Resolve a dispute (internal)
     */
    function _resolveDispute(bytes32 _disputeId) internal {
        Dispute storage dispute = disputes[_disputeId];
        
        bool upheld = dispute.votesFor > dispute.votesAgainst;
        dispute.status = DisputeStatus.Resolved;

        emit DisputeResolved(_disputeId, upheld, dispute.votesFor, dispute.votesAgainst);
    }

    /**
     * @dev Manually resolve dispute (owner only, for expired disputes)
     */
    function resolveDispute(bytes32 _disputeId) external onlyOwner {
        Dispute storage dispute = disputes[_disputeId];
        
        require(dispute.status == DisputeStatus.Active, "Dispute not active");
        require(block.timestamp >= dispute.votingEndsAt, "Voting period not ended");

        if (dispute.votesFor + dispute.votesAgainst < MIN_VOTES) {
            dispute.status = DisputeStatus.Expired;
        } else {
            _resolveDispute(_disputeId);
        }
    }

    /**
     * @dev Get dispute details
     */
    function getDispute(bytes32 _disputeId) external view returns (
        bytes32 gameId,
        address raiser,
        string memory reason,
        uint256 votesFor,
        uint256 votesAgainst,
        DisputeStatus status,
        uint256 votingEndsAt
    ) {
        Dispute storage dispute = disputes[_disputeId];
        return (
            dispute.gameId,
            dispute.raiser,
            dispute.reason,
            dispute.votesFor,
            dispute.votesAgainst,
            dispute.status,
            dispute.votingEndsAt
        );
    }

    /**
     * @dev Check if address has voted
     */
    function hasVoted(bytes32 _disputeId, address _voter) external view returns (bool) {
        return disputes[_disputeId].hasVoted[_voter];
    }

    /**
     * @dev Get vote of an address
     */
    function getVote(bytes32 _disputeId, address _voter) external view returns (bool) {
        require(disputes[_disputeId].hasVoted[_voter], "Has not voted");
        return disputes[_disputeId].votes[_voter];
    }

    /**
     * @dev Get all active disputes
     */
    function getActiveDisputes() external view returns (bytes32[] memory) {
        return activeDisputes;
    }
}
