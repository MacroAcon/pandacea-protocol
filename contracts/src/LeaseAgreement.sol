// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "./Reputation.sol";
import "./PGT.sol";

/**
 * @title ILeaseAgreement
 * @dev As per SDD v1.1, this is the initial interface for the LeaseAgreement contract.
 * It defines the core functions and events for creating and managing data leases.
 * [cite: 697-707]
 */
interface ILeaseAgreement {
    event LeaseCreated(bytes32 indexed leaseId, address indexed spender, address indexed earner, uint256 price);
    event LeaseApproved(bytes32 indexed leaseId);
    event LeaseExecuted(bytes32 indexed leaseId);
    event LeaseFinalized(bytes32 indexed leaseId, address indexed earner, uint256 reputationReward);
    event DisputeRaised(bytes32 indexed leaseId, address indexed spender, address indexed earner, string reason, uint256 stakeAmount);
    event DisputeResolved(bytes32 indexed leaseId, bool isDisputeValid, uint256 stakeAmount);
    event DisputeStakeRateUpdated(uint256 oldRate, uint256 newRate);

    function createLease(address earner, bytes32 dataProductId, uint256 maxPrice) external payable;
    function approveLease(bytes32 leaseId) external;
    function executeLease(bytes32 leaseId) external;
    function finalizeLease(bytes32 leaseId) external;
    function raiseDispute(bytes32 leaseId, string calldata reason) external;
    function resolveDispute(bytes32 leaseId, bool isDisputeValid) external;
    function setDisputeStakeRate(uint256 newRate) external;
    function getRequiredStake(bytes32 leaseId) external view returns (uint256);
}

/**
 * @title LeaseAgreement
 * @dev Implementation of the ILeaseAgreement interface with Dynamic Minimum Pricing (DMP) logic.
 * This contract manages data lease agreements between spenders and earners.
 */
contract LeaseAgreement is ILeaseAgreement, ReentrancyGuard, Ownable {
    
    // Dynamic Minimum Pricing (DMP) - minimum price required for lease creation
    uint256 public constant MIN_PRICE = 0.001 ether; // 0.001 ETH minimum price
    
    // Dispute window - time after execution before lease can be finalized
    uint256 public constant DISPUTE_WINDOW = 7 days;
    
    // Dispute stake rate - percentage of lease value required as stake (e.g., 10 = 10%)
    uint256 public disputeStakeRate;
    
    // Contract references
    Reputation public reputationContract;
    PGT public pgtToken;
    
    // DAO treasury address for stake forfeiture
    address public daoTreasury;
    
    // Lease structure to store lease data
    struct Lease {
        address spender;
        address earner;
        bytes32 dataProductId;
        uint256 price;
        uint256 maxPrice;
        bool isApproved;
        bool isExecuted;
        bool isDisputed;
        bool isFinalized;
        uint256 createdAt;
        uint256 executedAt;
        uint256 disputeId; // Reference to dispute in Reputation contract
        uint256 stakeAmount; // PGT tokens staked for dispute
    }
    
    // Mapping from leaseId to Lease struct
    mapping(bytes32 => Lease) public leases;
    
    // Mapping to track if a leaseId exists
    mapping(bytes32 => bool) public leaseExists;
    
    // Counter for generating unique lease IDs
    uint256 private _leaseCounter;
    
    constructor(address _reputationContract, address _pgtToken, address _daoTreasury) Ownable(msg.sender) {
        require(_reputationContract != address(0), "Invalid reputation contract address");
        require(_pgtToken != address(0), "Invalid PGT token address");
        require(_daoTreasury != address(0), "Invalid DAO treasury address");
        
        reputationContract = Reputation(_reputationContract);
        pgtToken = PGT(_pgtToken);
        daoTreasury = _daoTreasury;
        disputeStakeRate = 10; // Initialize to 10% stake rate
    }
    
    /**
     * @dev Creates a new lease agreement with Dynamic Minimum Pricing (DMP) validation.
     * @param earner The address of the data earner
     * @param dataProductId The unique identifier for the data product
     * @param maxPrice The maximum price the spender is willing to pay
     */
    function createLease(
        address earner, 
        bytes32 dataProductId, 
        uint256 maxPrice
    ) external payable override nonReentrant {
        // DMP Logic: Ensure msg.value is greater than or equal to MIN_PRICE
        require(msg.value >= MIN_PRICE, "LeaseAgreement: Insufficient payment - below minimum price");
        require(msg.value <= maxPrice, "LeaseAgreement: Payment exceeds maximum price");
        require(earner != address(0), "LeaseAgreement: Invalid earner address");
        require(earner != msg.sender, "LeaseAgreement: Spender cannot be earner");
        
        // Generate unique lease ID
        bytes32 leaseId = keccak256(abi.encodePacked(
            msg.sender,
            earner,
            dataProductId,
            block.timestamp,
            _leaseCounter
        ));
        _leaseCounter++;
        
        // Create lease record
        leases[leaseId] = Lease({
            spender: msg.sender,
            earner: earner,
            dataProductId: dataProductId,
            price: msg.value,
            maxPrice: maxPrice,
            isApproved: false,
            isExecuted: false,
            isDisputed: false,
            isFinalized: false,
            createdAt: block.timestamp,
            executedAt: 0,
            disputeId: 0,
            stakeAmount: 0
        });
        
        leaseExists[leaseId] = true;
        
        // Emit event
        emit LeaseCreated(leaseId, msg.sender, earner, msg.value);
        
        // TODO: Implement data product validation logic
        // TODO: Implement escrow mechanism for payment
    }
    
    /**
     * @dev Approves a lease by the designated earner.
     * @param leaseId The unique identifier of the lease to approve
     */
    function approveLease(bytes32 leaseId) external override nonReentrant {
        require(leaseExists[leaseId], "LeaseAgreement: Lease does not exist");
        require(msg.sender == leases[leaseId].earner, "LeaseAgreement: Only designated earner can approve");
        require(!leases[leaseId].isApproved, "LeaseAgreement: Lease already approved");
        require(!leases[leaseId].isDisputed, "LeaseAgreement: Cannot approve disputed lease");
        
        leases[leaseId].isApproved = true;
        
        emit LeaseApproved(leaseId);
        
        // TODO: Implement data transfer initiation
        // TODO: Implement payment escrow release conditions
    }
    
    /**
     * @dev Executes a lease after approval.
     * @param leaseId The unique identifier of the lease to execute
     */
    function executeLease(bytes32 leaseId) external override nonReentrant {
        require(leaseExists[leaseId], "LeaseAgreement: Lease does not exist");
        require(leases[leaseId].isApproved, "LeaseAgreement: Lease must be approved before execution");
        require(!leases[leaseId].isExecuted, "LeaseAgreement: Lease already executed");
        require(!leases[leaseId].isDisputed, "LeaseAgreement: Cannot execute disputed lease");
        
        // TODO: Implement access control for execution
        // TODO: Implement data access verification
        // TODO: Implement payment release logic
        
        leases[leaseId].isExecuted = true;
        leases[leaseId].executedAt = block.timestamp;
        
        emit LeaseExecuted(leaseId);
    }
    
    /**
     * @dev Finalizes a lease and rewards the earner with positive reputation.
     * @param leaseId The unique identifier of the lease to finalize
     */
    function finalizeLease(bytes32 leaseId) external override nonReentrant {
        require(leaseExists[leaseId], "LeaseAgreement: Lease does not exist");
        require(msg.sender == leases[leaseId].spender, "LeaseAgreement: Only spender can finalize lease");
        require(leases[leaseId].isExecuted, "LeaseAgreement: Lease must be executed before finalization");
        require(!leases[leaseId].isDisputed, "LeaseAgreement: Cannot finalize disputed lease");
        require(!leases[leaseId].isFinalized, "LeaseAgreement: Lease already finalized");
        
        Lease storage lease = leases[leaseId];
        
        // Check if dispute window has passed
        require(
            block.timestamp >= lease.executedAt + DISPUTE_WINDOW,
            "LeaseAgreement: Dispute window has not passed"
        );
        
        // Mark lease as finalized
        lease.isFinalized = true;
        
        // Reward earner with positive reputation
        reputationContract.updateReputation(lease.earner, true, lease.price);
        
        // Get the reputation change for the event
        uint256 reputationReward = _calculateReputationReward(lease.price);
        
        emit LeaseFinalized(leaseId, lease.earner, reputationReward);
    }
    
    /**
     * @dev Internal function to calculate reputation reward based on lease value
     * @param leaseValue Value of the lease in wei
     * @return Reputation points to be awarded
     */
    function _calculateReputationReward(uint256 leaseValue) internal pure returns (uint256) {
        // Tier 1: < 1 ETH (1e18 wei) = +25 points
        if (leaseValue < 1e18) {
            return 25;
        }
        // Tier 2: 1 ETH <= leaseValue < 10 ETH = +50 points
        else if (leaseValue < 10e18) {
            return 50;
        }
        // Tier 3: leaseValue >= 10 ETH = +100 points
        else {
            return 100;
        }
    }
    
    /**
     * @dev Calculates the required stake amount for a given lease based on the dispute stake rate.
     * @param leaseId The unique identifier of the lease
     * @return The required stake amount in PGT tokens
     */
    function getRequiredStake(bytes32 leaseId) external view override returns (uint256) {
        require(leaseExists[leaseId], "LeaseAgreement: Lease does not exist");
        Lease storage lease = leases[leaseId];
        return (lease.price * disputeStakeRate) / 100;
    }
    
    /**
     * @dev Raises a stake-based dispute for a lease with dynamic stake calculation.
     * @param leaseId The unique identifier of the lease to dispute
     * @param reason The reason for the dispute
     */
    function raiseDispute(bytes32 leaseId, string calldata reason) external override nonReentrant {
        require(leaseExists[leaseId], "LeaseAgreement: Lease does not exist");
        require(
            msg.sender == leases[leaseId].spender || msg.sender == leases[leaseId].earner,
            "LeaseAgreement: Only spender or earner can raise dispute"
        );
        require(!leases[leaseId].isDisputed, "LeaseAgreement: Dispute already raised");
        require(!leases[leaseId].isFinalized, "LeaseAgreement: Cannot dispute finalized lease");
        require(bytes(reason).length > 0, "LeaseAgreement: Dispute reason cannot be empty");
        
        Lease storage lease = leases[leaseId];
        
        // Calculate required stake based on lease value and dispute stake rate
        uint256 requiredStake = (lease.price * disputeStakeRate) / 100;
        require(requiredStake > 0, "LeaseAgreement: Calculated stake amount must be greater than 0");
        
        lease.isDisputed = true;
        lease.stakeAmount = requiredStake;
        
        // If spender is raising dispute against earner, handle staking and reputation
        if (msg.sender == lease.spender) {
            // Verify spender has approved this contract to spend their PGT tokens
            require(
                pgtToken.allowance(msg.sender, address(this)) >= requiredStake,
                "LeaseAgreement: Insufficient PGT allowance"
            );
            
            // Transfer PGT tokens from spender to this contract
            require(
                pgtToken.transferFrom(msg.sender, address(this), requiredStake),
                "LeaseAgreement: PGT transfer failed"
            );
            
            // Convert leaseId to uint256 for reputation contract
            uint256 leaseIdUint = uint256(leaseId);
            
            // Raise dispute in reputation contract
            reputationContract.raiseDispute(lease.earner, leaseIdUint, reason);
            
            // Get the dispute ID from reputation contract
            uint256 disputeCount = reputationContract.getDisputeCount();
            lease.disputeId = disputeCount - 1; // disputeCount is incremented after raising dispute
        }
        
        emit DisputeRaised(leaseId, lease.spender, lease.earner, reason, requiredStake);
    }
    
    /**
     * @dev Resolves a dispute and handles stake distribution based on validity.
     * @param leaseId The unique identifier of the lease
     * @param isDisputeValid Whether the dispute is valid (true) or invalid (false)
     */
    function resolveDispute(bytes32 leaseId, bool isDisputeValid) external override onlyOwner {
        require(leaseExists[leaseId], "LeaseAgreement: Lease does not exist");
        require(leases[leaseId].isDisputed, "LeaseAgreement: Lease is not disputed");
        require(leases[leaseId].stakeAmount > 0, "LeaseAgreement: No stake found for dispute");
        
        Lease storage lease = leases[leaseId];
        uint256 stakeAmount = lease.stakeAmount;
        
        if (isDisputeValid) {
            // Dispute is valid: penalize earner and return stake to spender
            reputationContract.updateReputation(lease.earner, false, lease.price);
            
            // Return stake to spender
            require(
                pgtToken.transfer(lease.spender, stakeAmount),
                "LeaseAgreement: Failed to return stake to spender"
            );
        } else {
            // Dispute is invalid: no reputation penalty, stake is forfeited
            // 50% to earner, 50% to DAO treasury
            uint256 earnerShare = stakeAmount / 2;
            uint256 treasuryShare = stakeAmount - earnerShare; // Handle odd amounts
            
            require(
                pgtToken.transfer(lease.earner, earnerShare),
                "LeaseAgreement: Failed to transfer stake to earner"
            );
            
            require(
                pgtToken.transfer(daoTreasury, treasuryShare),
                "LeaseAgreement: Failed to transfer stake to DAO treasury"
            );
        }
        
        // Clear the stake amount
        lease.stakeAmount = 0;
        
        emit DisputeResolved(leaseId, isDisputeValid, stakeAmount);
    }
    
    /**
     * @dev Allows the DAO to update the dispute stake rate.
     * @param newRate The new dispute stake rate (e.g., 10 for 10%)
     */
    function setDisputeStakeRate(uint256 newRate) external override onlyOwner {
        require(newRate > 0 && newRate <= 100, "LeaseAgreement: Stake rate must be between 1 and 100");
        uint256 oldRate = disputeStakeRate;
        disputeStakeRate = newRate;
        emit DisputeStakeRateUpdated(oldRate, newRate);
    }
    
    /**
     * @dev Returns the details of a specific lease.
     * @param leaseId The unique identifier of the lease
     * @return The complete lease structure
     */
    function getLease(bytes32 leaseId) external view returns (Lease memory) {
        require(leaseExists[leaseId], "LeaseAgreement: Lease does not exist");
        return leases[leaseId];
    }
    
    /**
     * @dev Allows the contract owner to update the minimum price (DMP).
     * @param newMinPrice The new minimum price in wei
     */
    function updateMinPrice(uint256 newMinPrice) external onlyOwner {
        // TODO: Implement minimum price update logic with proper validation
        // TODO: Implement price update events
        // TODO: Implement governance mechanism for price updates
    }
    
    /**
     * @dev Get dispute information for a lease from the reputation contract.
     * @param leaseId The unique identifier of the lease
     * @return disputeId The dispute ID in the reputation contract
     * @return spender Address of the spender who raised the dispute
     * @return earner Address of the earner being disputed
     * @return leaseIdUint The lease ID as uint256
     * @return timestamp Timestamp when dispute was raised
     * @return reason Reason for the dispute
     * @return resolved Whether the dispute is resolved
     * @return inFavorOfSpender Whether the dispute was resolved in favor of spender
     */
    function getDisputeInfo(bytes32 leaseId) external view returns (
        uint256 disputeId,
        address spender,
        address earner,
        uint256 leaseIdUint,
        uint256 timestamp,
        string memory reason,
        bool resolved,
        bool inFavorOfSpender
    ) {
        require(leaseExists[leaseId], "LeaseAgreement: Lease does not exist");
        require(leases[leaseId].isDisputed, "LeaseAgreement: Lease is not disputed");
        
        disputeId = leases[leaseId].disputeId;
        return reputationContract.getDispute(disputeId);
    }
    
    /**
     * @dev Update the reputation contract reference (only owner).
     * @param newReputationContract Address of the new reputation contract
     */
    function updateReputationContract(address newReputationContract) external onlyOwner {
        require(newReputationContract != address(0), "Invalid reputation contract address");
        reputationContract = Reputation(newReputationContract);
    }
    
    /**
     * @dev Update the PGT token contract reference (only owner).
     * @param newPgtToken Address of the new PGT token contract
     */
    function updatePgtToken(address newPgtToken) external onlyOwner {
        require(newPgtToken != address(0), "Invalid PGT token address");
        pgtToken = PGT(newPgtToken);
    }
    
    /**
     * @dev Update the DAO treasury address (only owner).
     * @param newDaoTreasury Address of the new DAO treasury
     */
    function updateDaoTreasury(address newDaoTreasury) external onlyOwner {
        require(newDaoTreasury != address(0), "Invalid DAO treasury address");
        daoTreasury = newDaoTreasury;
    }
    
    /**
     * @dev Emergency function to pause contract operations (only owner).
     */
    function emergencyPause() external onlyOwner {
        // TODO: Implement emergency pause functionality
        // TODO: Implement pause state management
        // TODO: Implement unpause functionality
    }
}