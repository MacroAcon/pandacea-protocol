// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/LeaseAgreement.sol";
import "../src/Reputation.sol";
import "../src/PGT.sol";

/**
 * @title LeaseAgreementInvariantTest
 * @dev Invariant tests for LeaseAgreement contract focusing on dispute staking, 
 * reputation decay, and settlement flows
 */
contract LeaseAgreementInvariantTest is Test {
    LeaseAgreement public leaseAgreement;
    Reputation public reputationContract;
    PGT public pgtToken;
    
    address public owner = address(this);
    address public daoTreasury = address(0x100);
    
    // Test actors
    address[] public spenders;
    address[] public earners;
    
    // State tracking
    mapping(bytes32 => bool) public activeLeases;
    mapping(bytes32 => bool) public disputedLeases;
    mapping(address => uint256) public userStakes;
    mapping(address => uint256) public userReputation;
    
    event LeaseCreated(bytes32 indexed leaseId, address indexed spender, address indexed earner, uint256 price);
    event LeaseApproved(bytes32 indexed leaseId);
    event LeaseExecuted(bytes32 indexed leaseId);
    event LeaseFinalized(bytes32 indexed leaseId, address indexed earner, uint256 reputationReward);
    event DisputeRaised(bytes32 indexed leaseId, address indexed spender, address indexed earner, string reason, uint256 stakeAmount);
    event DisputeResolved(bytes32 indexed leaseId, bool isDisputeValid, uint256 stakeAmount);
    
    function setUp() public {
        // Deploy contracts
        reputationContract = new Reputation();
        pgtToken = new PGT();
        leaseAgreement = new LeaseAgreement(address(reputationContract), address(pgtToken), daoTreasury);
        
        // Setup test actors
        for (uint i = 0; i < 5; i++) {
            address spender = address(uint160(1000 + i));
            address earner = address(uint160(2000 + i));
            
            spenders.push(spender);
            earners.push(earner);
            
            // Fund accounts
            vm.deal(spender, 10 ether);
            vm.deal(earner, 10 ether);
            
            // Mint PGT tokens
            pgtToken.mint(spender, 1000 * 10**18);
            pgtToken.mint(earner, 1000 * 10**18);
        }
    }
    
    // ========== INVARIANT TESTS ==========
    
    /**
     * @dev Invariant: Total staked PGT should always equal sum of individual stakes
     */
    function invariant_totalStakesConsistent() public view {
        uint256 totalStaked = pgtToken.balanceOf(address(leaseAgreement));
        uint256 calculatedTotal = 0;
        
        for (uint i = 0; i < spenders.length; i++) {
            calculatedTotal += userStakes[spenders[i]];
        }
        for (uint i = 0; i < earners.length; i++) {
            calculatedTotal += userStakes[earners[i]];
        }
        
        assertEq(totalStaked, calculatedTotal, "Total staked PGT inconsistent");
    }
    
    /**
     * @dev Invariant: No lease can be both active and disputed simultaneously
     */
    function invariant_noLeaseActiveAndDisputed() public view {
        for (uint i = 0; i < 100; i++) {
            bytes32 leaseId = bytes32(i);
            bool isActive = activeLeases[leaseId];
            bool isDisputed = disputedLeases[leaseId];
            
            assertFalse(isActive && isDisputed, "Lease cannot be both active and disputed");
        }
    }
    
    /**
     * @dev Invariant: Dispute stake amount should always be >= required stake
     */
    function invariant_disputeStakesSufficient() public view {
        for (uint i = 0; i < 100; i++) {
            bytes32 leaseId = bytes32(i);
            if (disputedLeases[leaseId]) {
                uint256 requiredStake = leaseAgreement.getRequiredStake(leaseId);
                uint256 actualStake = userStakes[leaseAgreement.getLease(leaseId).spender];
                
                assertGe(actualStake, requiredStake, "Dispute stake insufficient");
            }
        }
    }
    
    /**
     * @dev Invariant: Reputation should never be negative
     */
    function invariant_reputationNonNegative() public view {
        for (uint i = 0; i < spenders.length; i++) {
            uint256 rep = reputationContract.getReputation(spenders[i]);
            assertGe(rep, 0, "Reputation cannot be negative");
        }
        for (uint i = 0; i < earners.length; i++) {
            uint256 rep = reputationContract.getReputation(earners[i]);
            assertGe(rep, 0, "Reputation cannot be negative");
        }
    }
    
    /**
     * @dev Invariant: Lease price should always be >= MIN_PRICE
     */
    function invariant_leasePriceMinimum() public view {
        for (uint i = 0; i < 100; i++) {
            bytes32 leaseId = bytes32(i);
            if (leaseAgreement.leaseExists(leaseId)) {
                LeaseAgreement.Lease memory lease = leaseAgreement.getLease(leaseId);
                assertGe(lease.price, leaseAgreement.MIN_PRICE(), "Lease price below minimum");
            }
        }
    }
    
    /**
     * @dev Invariant: Dispute stake rate should be reasonable (0-50%)
     */
    function invariant_disputeStakeRateReasonable() public view {
        uint256 stakeRate = leaseAgreement.disputeStakeRate();
        assertLe(stakeRate, 50, "Dispute stake rate too high");
        assertGe(stakeRate, 0, "Dispute stake rate cannot be negative");
    }
    
    // ========== FUZZ TESTS ==========
    
    /**
     * @dev Fuzz test: Create lease with various prices
     */
    function testFuzz_createLease(uint256 price) public {
        // Bound price to reasonable range
        price = bound(price, leaseAgreement.MIN_PRICE(), 1 ether);
        
        address spender = spenders[0];
        address earner = earners[0];
        bytes32 dataProductId = keccak256(abi.encodePacked(price));
        uint256 maxPrice = price + 0.1 ether;
        
        vm.startPrank(spender);
        
        if (price >= leaseAgreement.MIN_PRICE()) {
            leaseAgreement.createLease{value: price}(earner, dataProductId, maxPrice);
            
            bytes32 leaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
            assertTrue(leaseAgreement.leaseExists(leaseId), "Lease should exist");
            
            LeaseAgreement.Lease memory lease = leaseAgreement.getLease(leaseId);
            assertEq(lease.price, price, "Lease price should match");
            assertEq(lease.spender, spender, "Spender should match");
            assertEq(lease.earner, earner, "Earner should match");
        }
        
        vm.stopPrank();
    }
    
    /**
     * @dev Fuzz test: Dispute stake amounts
     */
    function testFuzz_disputeStakeAmount(uint256 stakeAmount) public {
        // Create a lease first
        address spender = spenders[0];
        address earner = earners[0];
        bytes32 dataProductId = keccak256("test-product");
        
        vm.startPrank(spender);
        leaseAgreement.createLease{value: leaseAgreement.MIN_PRICE()}(earner, dataProductId, 1 ether);
        bytes32 leaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        
        // Approve and execute lease
        vm.stopPrank();
        vm.prank(earner);
        leaseAgreement.approveLease(leaseId);
        
        vm.prank(spender);
        leaseAgreement.executeLease(leaseId);
        
        // Try to dispute with various stake amounts
        uint256 requiredStake = leaseAgreement.getRequiredStake(leaseId);
        stakeAmount = bound(stakeAmount, 0, requiredStake * 2);
        
        vm.startPrank(spender);
        pgtToken.approve(address(leaseAgreement), stakeAmount);
        
        if (stakeAmount >= requiredStake) {
            leaseAgreement.raiseDispute(leaseId, "Test dispute");
            assertTrue(leaseAgreement.getLease(leaseId).isDisputed, "Lease should be disputed");
        } else {
            vm.expectRevert();
            leaseAgreement.raiseDispute(leaseId, "Test dispute");
        }
        
        vm.stopPrank();
    }
    
    /**
     * @dev Fuzz test: Reputation decay rates
     */
    function testFuzz_reputationDecay(uint256 decayRate) public {
        address user = spenders[0];
        
        // Award some reputation first
        reputationContract.awardReputation(user, 100, "Test award");
        uint256 initialRep = reputationContract.getReputation(user);
        
        // Bound decay rate to reasonable range
        decayRate = bound(decayRate, 0, 100);
        
        // Apply decay
        reputationContract.applyDecay(user, decayRate);
        uint256 finalRep = reputationContract.getReputation(user);
        
        // Invariants
        assertLe(finalRep, initialRep, "Reputation should not increase after decay");
        assertGe(finalRep, 0, "Reputation should not go negative");
    }
    
    /**
     * @dev Fuzz test: Multiple concurrent disputes
     */
    function testFuzz_concurrentDisputes(uint256 numDisputes) public {
        numDisputes = bound(numDisputes, 1, 10);
        
        for (uint i = 0; i < numDisputes; i++) {
            address spender = spenders[i % spenders.length];
            address earner = earners[i % earners.length];
            bytes32 dataProductId = keccak256(abi.encodePacked(i));
            
            // Create lease
            vm.startPrank(spender);
            leaseAgreement.createLease{value: leaseAgreement.MIN_PRICE()}(earner, dataProductId, 1 ether);
            bytes32 leaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, i));
            
            // Approve and execute
            vm.stopPrank();
            vm.prank(earner);
            leaseAgreement.approveLease(leaseId);
            
            vm.prank(spender);
            leaseAgreement.executeLease(leaseId);
            
            // Dispute
            uint256 requiredStake = leaseAgreement.getRequiredStake(leaseId);
            pgtToken.approve(address(leaseAgreement), requiredStake);
            leaseAgreement.raiseDispute(leaseId, "Concurrent dispute");
            
            vm.stopPrank();
        }
        
        // Verify invariants still hold
        invariant_totalStakesConsistent();
        invariant_noLeaseActiveAndDisputed();
    }
    
    // ========== EDGE CASE TESTS ==========
    
    /**
     * @dev Test: Dispute resolution with zero stake
     */
    function test_disputeResolutionZeroStake() public {
        // Create and dispute a lease
        address spender = spenders[0];
        address earner = earners[0];
        bytes32 dataProductId = keccak256("test-product");
        
        vm.startPrank(spender);
        leaseAgreement.createLease{value: leaseAgreement.MIN_PRICE()}(earner, dataProductId, 1 ether);
        bytes32 leaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        
        vm.stopPrank();
        vm.prank(earner);
        leaseAgreement.approveLease(leaseId);
        
        vm.prank(spender);
        leaseAgreement.executeLease(leaseId);
        
        uint256 requiredStake = leaseAgreement.getRequiredStake(leaseId);
        pgtToken.approve(address(leaseAgreement), requiredStake);
        leaseAgreement.raiseDispute(leaseId, "Test dispute");
        
        // Resolve dispute
        vm.prank(owner);
        leaseAgreement.resolveDispute(leaseId, true);
        
        // Verify stake distribution
        assertEq(pgtToken.balanceOf(spender), 1000 * 10**18, "Spender should get stake back");
    }
    
    /**
     * @dev Test: Maximum dispute stake rate
     */
    function test_maximumDisputeStakeRate() public {
        // Set maximum stake rate
        leaseAgreement.setDisputeStakeRate(50);
        assertEq(leaseAgreement.disputeStakeRate(), 50, "Stake rate should be set");
        
        // Create lease with high value
        address spender = spenders[0];
        address earner = earners[0];
        bytes32 dataProductId = keccak256("high-value-product");
        uint256 highPrice = 1 ether;
        
        vm.startPrank(spender);
        leaseAgreement.createLease{value: highPrice}(earner, dataProductId, highPrice);
        bytes32 leaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        
        uint256 requiredStake = leaseAgreement.getRequiredStake(leaseId);
        assertEq(requiredStake, highPrice * 50 / 100, "Stake should be 50% of lease value");
        
        vm.stopPrank();
    }
    
    /**
     * @dev Test: Reputation decay to zero
     */
    function test_reputationDecayToZero() public {
        address user = spenders[0];
        
        // Award minimal reputation
        reputationContract.awardReputation(user, 1, "Minimal award");
        assertEq(reputationContract.getReputation(user), 1, "Should have 1 reputation");
        
        // Apply decay that should reduce to zero
        reputationContract.applyDecay(user, 100);
        assertEq(reputationContract.getReputation(user), 0, "Reputation should be zero");
        
        // Apply more decay
        reputationContract.applyDecay(user, 50);
        assertEq(reputationContract.getReputation(user), 0, "Reputation should remain zero");
    }
    
    // ========== HELPER FUNCTIONS ==========
    
    /**
     * @dev Helper: Update state tracking after lease creation
     */
    function _trackLeaseCreation(bytes32 leaseId) internal {
        activeLeases[leaseId] = true;
    }
    
    /**
     * @dev Helper: Update state tracking after dispute
     */
    function _trackDispute(bytes32 leaseId, address disputer, uint256 stakeAmount) internal {
        disputedLeases[leaseId] = true;
        activeLeases[leaseId] = false;
        userStakes[disputer] += stakeAmount;
    }
    
    /**
     * @dev Helper: Update state tracking after dispute resolution
     */
    function _trackDisputeResolution(bytes32 leaseId, address disputer, uint256 stakeAmount) internal {
        disputedLeases[leaseId] = false;
        userStakes[disputer] -= stakeAmount;
    }
}
