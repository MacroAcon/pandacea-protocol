// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

/**
 * @title Reputation
 * @dev Manages reputation scores for Earners in the Pandacea Protocol
 * @dev Implements automated "just-in-time" reputation decay mechanism and dispute resolution
 */
contract Reputation is Ownable, ReentrancyGuard {
    
    // Events
    event ReputationUpdated(address indexed earner, uint256 oldScore, uint256 newScore, string reason);
    event DisputeRaised(address indexed spender, address indexed earner, uint256 leaseId, string reason);
    event ReputationDecay(address indexed earner, uint256 oldScore, uint256 newScore);
    
    // Structs
    struct ReputationData {
        uint256 score;                    // Current reputation score (0-1000)
        uint256 lastUpdatedTimestamp;    // Timestamp of last update (for decay calculation)
        uint256 totalDisputes;           // Total number of disputes raised against this earner
        uint256 resolvedDisputes;        // Number of disputes resolved in favor of spender
        bool isActive;                   // Whether the earner is currently active
    }
    
    // State variables
    mapping(address => ReputationData) public reputationData;
    mapping(uint256 => Dispute) public disputes;
    
    uint256 public constant MAX_REPUTATION = 1000;
    uint256 public constant MIN_REPUTATION = 0;
    uint256 public constant REPUTATION_DECAY_RATE = 1; // 1 point per day
    uint256 public constant DISPUTE_COOLDOWN = 7 days; // Cooldown between disputes
    
    uint256 public disputeCounter;
    
    struct Dispute {
        address spender;
        address earner;
        uint256 leaseId;
        uint256 timestamp;
        string reason;
        bool resolved;
        bool inFavorOfSpender;
    }
    
    // Modifiers
    modifier onlyValidAddress(address addr) {
        require(addr != address(0), "Invalid address");
        _;
    }
    
    modifier onlyValidReputation(uint256 score) {
        require(score <= MAX_REPUTATION, "Reputation score too high");
        _;
    }
    
    modifier onlyActiveEarner(address earner) {
        require(reputationData[earner].isActive, "Earner not active");
        _;
    }
    
    constructor() {
        disputeCounter = 0;
    }
    
    /**
     * @dev Initialize reputation for a new earner
     * @param earner Address of the earner
     * @param initialScore Initial reputation score
     */
    function initializeReputation(address earner, uint256 initialScore) 
        external 
        onlyOwner 
        onlyValidAddress(earner)
        onlyValidReputation(initialScore)
    {
        require(reputationData[earner].lastUpdatedTimestamp == 0, "Reputation already initialized");
        
        reputationData[earner] = ReputationData({
            score: initialScore,
            lastUpdatedTimestamp: block.timestamp,
            totalDisputes: 0,
            resolvedDisputes: 0,
            isActive: true
        });
        
        emit ReputationUpdated(earner, 0, initialScore, "Initialization");
    }
    
    /**
     * @dev Update reputation score for an earner based on lease success and value
     * @dev Applies "just-in-time" decay before applying new score changes
     * @param earner Address of the earner
     * @param successfulLease Whether the lease was successful
     * @param leaseValue Value of the lease in wei
     */
    function updateReputation(address earner, bool successfulLease, uint256 leaseValue) 
        external 
        onlyOwner 
        onlyValidAddress(earner)
    {
        ReputationData storage data = reputationData[earner];
        require(data.lastUpdatedTimestamp > 0, "Earner not initialized");
        
        // Apply "just-in-time" decay first
        uint256 currentScore = _applyDecayAndGetCurrentScore(earner);
        
        // Calculate reputation change based on lease value
        uint256 reputationChange = calculateReputationChange(leaseValue);
        
        uint256 newScore;
        if (successfulLease) {
            // Positive reputation change for successful lease
            newScore = currentScore + reputationChange;
            if (newScore > MAX_REPUTATION) {
                newScore = MAX_REPUTATION;
            }
        } else {
            // Negative reputation change for failed lease
            if (currentScore > reputationChange) {
                newScore = currentScore - reputationChange;
            } else {
                newScore = 0;
            }
        }
        
        // Update the reputation data
        data.score = newScore;
        data.lastUpdatedTimestamp = block.timestamp;
        
        string memory reason = successfulLease ? "Successful lease completion" : "Failed lease";
        emit ReputationUpdated(earner, currentScore, newScore, reason);
    }
    
    /**
     * @dev Internal function to apply decay and return current score
     * @param earner Address of the earner
     * @return Current score after applying decay
     */
    function _applyDecayAndGetCurrentScore(address earner) internal view returns (uint256) {
        ReputationData storage data = reputationData[earner];
        uint256 timePassed = block.timestamp - data.lastUpdatedTimestamp;
        uint256 daysPassed = timePassed / 1 days;
        
        if (daysPassed == 0) {
            return data.score;
        }
        
        uint256 decayAmount = daysPassed * REPUTATION_DECAY_RATE;
        if (decayAmount >= data.score) {
            return 0;
        } else {
            return data.score - decayAmount;
        }
    }
    
    /**
     * @dev Get current reputation score with "just-in-time" decay applied
     * @param earner Address of the earner
     * @return Current reputation score after applying decay
     */
    function getReputation(address earner) external view returns (uint256) {
        return _applyDecayAndGetCurrentScore(earner);
    }
    
    /**
     * @dev Calculate reputation change based on lease value tier
     * @param leaseValue Value of the lease in wei
     * @return Reputation points to add/subtract
     */
    function calculateReputationChange(uint256 leaseValue) internal pure returns (uint256) {
        // Tier 1: < 1 ETH (1e18 wei) = +/- 25 points
        if (leaseValue < 1e18) {
            return 25;
        }
        // Tier 2: 1 ETH <= leaseValue < 10 ETH = +/- 50 points
        else if (leaseValue < 10e18) {
            return 50;
        }
        // Tier 3: leaseValue >= 10 ETH = +/- 100 points
        else {
            return 100;
        }
    }
    
    /**
     * @dev Raise a dispute against an earner
     * @param earner Address of the earner being disputed
     * @param leaseId ID of the lease being disputed
     * @param reason Reason for the dispute
     */
    function raiseDispute(address earner, uint256 leaseId, string memory reason) 
        external 
        nonReentrant
        onlyValidAddress(earner)
        onlyActiveEarner(earner)
    {
        require(msg.sender != earner, "Cannot dispute yourself");
        
        // Check cooldown period
        uint256 lastDispute = 0;
        for (uint256 i = 0; i < disputeCounter; i++) {
            if (disputes[i].spender == msg.sender && disputes[i].earner == earner) {
                if (disputes[i].timestamp > lastDispute) {
                    lastDispute = disputes[i].timestamp;
                }
            }
        }
        
        require(block.timestamp >= lastDispute + DISPUTE_COOLDOWN, "Dispute cooldown not met");
        
        disputes[disputeCounter] = Dispute({
            spender: msg.sender,
            earner: earner,
            leaseId: leaseId,
            timestamp: block.timestamp,
            reason: reason,
            resolved: false,
            inFavorOfSpender: false
        });
        
        reputationData[earner].totalDisputes++;
        
        emit DisputeRaised(msg.sender, earner, leaseId, reason);
        disputeCounter++;
    }
    
    /**
     * @dev Resolve a dispute (only owner)
     * @param disputeId ID of the dispute to resolve
     * @param inFavorOfSpender Whether the dispute is resolved in favor of the spender
     */
    function resolveDispute(uint256 disputeId, bool inFavorOfSpender) 
        external 
        onlyOwner 
    {
        require(disputeId < disputeCounter, "Invalid dispute ID");
        require(!disputes[disputeId].resolved, "Dispute already resolved");
        
        Dispute storage dispute = disputes[disputeId];
        dispute.resolved = true;
        dispute.inFavorOfSpender = inFavorOfSpender;
        
        if (inFavorOfSpender) {
            address earner = dispute.earner;
            reputationData[earner].resolvedDisputes++;
            
            // Apply penalty using the new updateReputation function
            // This will automatically apply decay and then subtract the penalty
            updateReputation(earner, false, 1e18); // Use 1 ETH tier for dispute penalty
        }
    }
    
    /**
     * @dev Get current reputation data for an earner (with decay applied)
     * @param earner Address of the earner
     * @return score Current reputation score (after decay)
     * @return lastUpdatedTimestamp Timestamp of last update
     * @return totalDisputes Total number of disputes
     * @return resolvedDisputes Number of resolved disputes
     * @return isActive Whether earner is active
     */
    function getReputationData(address earner) 
        external 
        view 
        returns (
            uint256 score,
            uint256 lastUpdatedTimestamp,
            uint256 totalDisputes,
            uint256 resolvedDisputes,
            bool isActive
        )
    {
        ReputationData memory data = reputationData[earner];
        return (
            _applyDecayAndGetCurrentScore(earner),
            data.lastUpdatedTimestamp,
            data.totalDisputes,
            data.resolvedDisputes,
            data.isActive
        );
    }
    
    /**
     * @dev Get dispute information
     * @param disputeId ID of the dispute
     * @return spender Address of the spender
     * @return earner Address of the earner
     * @return leaseId ID of the lease
     * @return timestamp Timestamp of dispute
     * @return reason Reason for dispute
     * @return resolved Whether dispute is resolved
     * @return inFavorOfSpender Whether resolved in favor of spender
     */
    function getDispute(uint256 disputeId) 
        external 
        view 
        returns (
            address spender,
            address earner,
            uint256 leaseId,
            uint256 timestamp,
            string memory reason,
            bool resolved,
            bool inFavorOfSpender
        )
    {
        require(disputeId < disputeCounter, "Invalid dispute ID");
        Dispute memory dispute = disputes[disputeId];
        return (
            dispute.spender,
            dispute.earner,
            dispute.leaseId,
            dispute.timestamp,
            dispute.reason,
            dispute.resolved,
            dispute.inFavorOfSpender
        );
    }
    
    /**
     * @dev Deactivate an earner (only owner)
     * @param earner Address of the earner to deactivate
     */
    function deactivateEarner(address earner) 
        external 
        onlyOwner 
        onlyValidAddress(earner)
    {
        reputationData[earner].isActive = false;
    }
    
    /**
     * @dev Reactivate an earner (only owner)
     * @param earner Address of the earner to reactivate
     */
    function reactivateEarner(address earner) 
        external 
        onlyOwner 
        onlyValidAddress(earner)
    {
        reputationData[earner].isActive = true;
    }
    
    /**
     * @dev Get total number of disputes
     * @return Total dispute count
     */
    function getDisputeCount() external view returns (uint256) {
        return disputeCounter;
    }
}
