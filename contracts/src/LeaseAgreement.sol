// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

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

    function createLease(address earner, bytes32 dataProductId, uint256 maxPrice) external payable;
    function approveLease(bytes32 leaseId) external;
    function executeLease(bytes32 leaseId) external;
    function raiseDispute(bytes32 leaseId, string calldata reason) external;
}

/**
 * @title LeaseAgreement
 * @dev Implementation of the ILeaseAgreement interface with Dynamic Minimum Pricing (DMP) logic.
 * This contract manages data lease agreements between spenders and earners.
 */
contract LeaseAgreement is ILeaseAgreement, ReentrancyGuard, Ownable {
    
    // Dynamic Minimum Pricing (DMP) - minimum price required for lease creation
    uint256 public constant MIN_PRICE = 0.001 ether; // 0.001 ETH minimum price
    
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
        uint256 createdAt;
    }
    
    // Mapping from leaseId to Lease struct
    mapping(bytes32 => Lease) public leases;
    
    // Mapping to track if a leaseId exists
    mapping(bytes32 => bool) public leaseExists;
    
    // Counter for generating unique lease IDs
    uint256 private _leaseCounter;
    
    constructor() Ownable(msg.sender) {}
    
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
            createdAt: block.timestamp
        });
        
        leaseExists[leaseId] = true;
        
        // Emit event
        emit LeaseCreated(leaseId, msg.sender, earner, msg.value);
        
        // TODO: Implement data product validation logic
        // TODO: Implement reputation system integration
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
        
        emit LeaseExecuted(leaseId);
    }
    
    /**
     * @dev Raises a dispute for a lease.
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
        require(bytes(reason).length > 0, "LeaseAgreement: Dispute reason cannot be empty");
        
        // TODO: Implement dispute resolution mechanism
        // TODO: Implement arbitration system
        // TODO: Implement refund logic for disputed leases
        
        leases[leaseId].isDisputed = true;
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
     * @dev Emergency function to pause contract operations (only owner).
     */
    function emergencyPause() external onlyOwner {
        // TODO: Implement emergency pause functionality
        // TODO: Implement pause state management
        // TODO: Implement unpause functionality
    }
}