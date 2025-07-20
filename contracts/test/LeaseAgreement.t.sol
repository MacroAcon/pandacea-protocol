// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/LeaseAgreement.sol";

/**
 * @title LeaseAgreementTest
 * @dev Comprehensive unit tests for the LeaseAgreement contract
 */
contract LeaseAgreementTest is Test {
    LeaseAgreement public leaseAgreement;
    
    address public spender = address(0x1);
    address public earner = address(0x2);
    address public nonEarner = address(0x3);
    
    bytes32 public dataProductId = keccak256("test-data-product");
    uint256 public maxPrice = 0.01 ether;
    uint256 public minPrice = 0.001 ether; // MIN_PRICE from contract
    
    event LeaseCreated(bytes32 indexed leaseId, address indexed spender, address indexed earner, uint256 price);
    event LeaseApproved(bytes32 indexed leaseId);
    
    function setUp() public {
        leaseAgreement = new LeaseAgreement();
        
        // Fund test accounts
        vm.deal(spender, 10 ether);
        vm.deal(earner, 10 ether);
        vm.deal(nonEarner, 10 ether);
    }
    
    // ========== SUCCESS CASES ==========
    
    function testCreateLeaseWithExactMinPrice() public {
        // Test: A user can successfully create a lease by sending exactly the MIN_PRICE
        vm.startPrank(spender);
        
        vm.expectEmit(true, true, true, true);
        emit LeaseCreated(bytes32(0), spender, earner, minPrice);
        
        leaseAgreement.createLease{value: minPrice}(earner, dataProductId, maxPrice);
        
        vm.stopPrank();
        
        // Verify lease was created
        bytes32 expectedLeaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        assertTrue(leaseAgreement.leaseExists(expectedLeaseId));
        
        LeaseAgreement.Lease memory lease = leaseAgreement.getLease(expectedLeaseId);
        assertEq(lease.spender, spender);
        assertEq(lease.earner, earner);
        assertEq(lease.dataProductId, dataProductId);
        assertEq(lease.price, minPrice);
        assertEq(lease.maxPrice, maxPrice);
        assertFalse(lease.isApproved);
        assertFalse(lease.isExecuted);
        assertFalse(lease.isDisputed);
    }
    
    function testCreateLeaseWithMoreThanMinPrice() public {
        // Test: A user can successfully create a lease by sending more than MIN_PRICE
        uint256 payment = 0.005 ether; // More than min price
        
        vm.startPrank(spender);
        
        vm.expectEmit(true, true, true, true);
        emit LeaseCreated(bytes32(0), spender, earner, payment);
        
        leaseAgreement.createLease{value: payment}(earner, dataProductId, maxPrice);
        
        vm.stopPrank();
        
        // Verify lease was created with correct price
        bytes32 expectedLeaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        LeaseAgreement.Lease memory lease = leaseAgreement.getLease(expectedLeaseId);
        assertEq(lease.price, payment);
    }
    
    function testLeaseCreatedEventEmittedCorrectly() public {
        // Test: The LeaseCreated event is correctly emitted with all parameters
        vm.startPrank(spender);
        
        vm.expectEmit(true, true, true, true);
        emit LeaseCreated(bytes32(0), spender, earner, minPrice);
        
        leaseAgreement.createLease{value: minPrice}(earner, dataProductId, maxPrice);
        
        vm.stopPrank();
    }
    
    function testEarnerCanApproveLease() public {
        // Test: The designated earner of a lease can successfully approve it
        // First create a lease
        vm.prank(spender);
        leaseAgreement.createLease{value: minPrice}(earner, dataProductId, maxPrice);
        
        bytes32 leaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        
        // Then approve it as the earner
        vm.startPrank(earner);
        
        vm.expectEmit(true, false, false, false);
        emit LeaseApproved(leaseId);
        
        leaseAgreement.approveLease(leaseId);
        
        vm.stopPrank();
        
        // Verify lease is approved
        LeaseAgreement.Lease memory lease = leaseAgreement.getLease(leaseId);
        assertTrue(lease.isApproved);
    }
    
    function testMultipleLeasesCanBeCreated() public {
        // Test: Multiple leases can be created with different IDs
        vm.startPrank(spender);
        
        leaseAgreement.createLease{value: minPrice}(earner, dataProductId, maxPrice);
        leaseAgreement.createLease{value: minPrice}(earner, keccak256("different-product"), maxPrice);
        
        vm.stopPrank();
        
        // Verify both leases exist
        bytes32 leaseId1 = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        bytes32 leaseId2 = keccak256(abi.encodePacked(spender, earner, keccak256("different-product"), block.timestamp, 1));
        
        assertTrue(leaseAgreement.leaseExists(leaseId1));
        assertTrue(leaseAgreement.leaseExists(leaseId2));
    }
    
    // ========== FAILURE CASES ==========
    
    function testCreateLeaseRevertsWhenPaymentBelowMinPrice() public {
        // Test: The createLease function MUST revert if the msg.value sent is less than MIN_PRICE
        uint256 insufficientPayment = 0.0005 ether; // Less than min price
        
        vm.startPrank(spender);
        
        vm.expectRevert("LeaseAgreement: Insufficient payment - below minimum price");
        leaseAgreement.createLease{value: insufficientPayment}(earner, dataProductId, maxPrice);
        
        vm.stopPrank();
    }
    
    function testCreateLeaseRevertsWhenPaymentExceedsMaxPrice() public {
        // Test: The createLease function reverts when payment exceeds maxPrice
        uint256 excessivePayment = 0.02 ether; // More than maxPrice (0.01 ether)
        
        vm.startPrank(spender);
        
        vm.expectRevert("LeaseAgreement: Payment exceeds maximum price");
        leaseAgreement.createLease{value: excessivePayment}(earner, dataProductId, maxPrice);
        
        vm.stopPrank();
    }
    
    function testCreateLeaseRevertsWithZeroAddressEarner() public {
        // Test: The createLease function reverts with zero address earner
        vm.startPrank(spender);
        
        vm.expectRevert("LeaseAgreement: Invalid earner address");
        leaseAgreement.createLease{value: minPrice}(address(0), dataProductId, maxPrice);
        
        vm.stopPrank();
    }
    
    function testCreateLeaseRevertsWhenSpenderIsEarner() public {
        // Test: The createLease function reverts when spender is the same as earner
        vm.startPrank(spender);
        
        vm.expectRevert("LeaseAgreement: Spender cannot be earner");
        leaseAgreement.createLease{value: minPrice}(spender, dataProductId, maxPrice);
        
        vm.stopPrank();
    }
    
    function testApproveLeaseRevertsWhenCalledByNonEarner() public {
        // Test: The approveLease function MUST revert if it is called by any address other than the lease's designated earner
        // First create a lease
        vm.prank(spender);
        leaseAgreement.createLease{value: minPrice}(earner, dataProductId, maxPrice);
        
        bytes32 leaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        
        // Try to approve as non-earner
        vm.startPrank(nonEarner);
        
        vm.expectRevert("LeaseAgreement: Only designated earner can approve");
        leaseAgreement.approveLease(leaseId);
        
        vm.stopPrank();
    }
    
    function testApproveLeaseRevertsWhenLeaseDoesNotExist() public {
        // Test: The approveLease function MUST revert if it is called for a leaseId that does not exist
        bytes32 nonExistentLeaseId = keccak256("non-existent-lease");
        
        vm.startPrank(earner);
        
        vm.expectRevert("LeaseAgreement: Lease does not exist");
        leaseAgreement.approveLease(nonExistentLeaseId);
        
        vm.stopPrank();
    }
    
    function testApproveLeaseRevertsWhenAlreadyApproved() public {
        // Test: The approveLease function reverts when lease is already approved
        // First create and approve a lease
        vm.prank(spender);
        leaseAgreement.createLease{value: minPrice}(earner, dataProductId, maxPrice);
        
        bytes32 leaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        
        vm.prank(earner);
        leaseAgreement.approveLease(leaseId);
        
        // Try to approve again
        vm.startPrank(earner);
        
        vm.expectRevert("LeaseAgreement: Lease already approved");
        leaseAgreement.approveLease(leaseId);
        
        vm.stopPrank();
    }
    
    function testApproveLeaseRevertsWhenLeaseIsDisputed() public {
        // Test: The approveLease function reverts when lease is disputed
        // First create a lease
        vm.prank(spender);
        leaseAgreement.createLease{value: minPrice}(earner, dataProductId, maxPrice);
        
        bytes32 leaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        
        // Raise a dispute
        vm.prank(spender);
        leaseAgreement.raiseDispute(leaseId, "Test dispute");
        
        // Try to approve the disputed lease
        vm.startPrank(earner);
        
        vm.expectRevert("LeaseAgreement: Cannot approve disputed lease");
        leaseAgreement.approveLease(leaseId);
        
        vm.stopPrank();
    }
    
    function testExecuteLeaseRevertsWhenNotApproved() public {
        // Test: The executeLease function reverts when lease is not approved
        // First create a lease without approving it
        vm.prank(spender);
        leaseAgreement.createLease{value: minPrice}(earner, dataProductId, maxPrice);
        
        bytes32 leaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        
        // Try to execute without approval
        vm.expectRevert("LeaseAgreement: Lease must be approved before execution");
        leaseAgreement.executeLease(leaseId);
    }
    
    function testExecuteLeaseRevertsWhenLeaseDoesNotExist() public {
        // Test: The executeLease function reverts when lease does not exist
        bytes32 nonExistentLeaseId = keccak256("non-existent-lease");
        
        vm.expectRevert("LeaseAgreement: Lease does not exist");
        leaseAgreement.executeLease(nonExistentLeaseId);
    }
    
    function testRaiseDisputeRevertWhenLeaseDoesNotExist() public {
        // Test: The raiseDispute function reverts when lease does not exist
        bytes32 nonExistentLeaseId = keccak256("non-existent-lease");
        
        vm.expectRevert("LeaseAgreement: Lease does not exist");
        leaseAgreement.raiseDispute(nonExistentLeaseId, "Test dispute");
    }
    
    function testRaiseDisputeRevertWhenCalledByUnauthorized() public {
        // Test: The raiseDispute function reverts when called by unauthorized address
        // First create a lease
        vm.prank(spender);
        leaseAgreement.createLease{value: minPrice}(earner, dataProductId, maxPrice);
        
        bytes32 leaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        
        // Try to raise dispute as unauthorized address
        vm.startPrank(nonEarner);
        
        vm.expectRevert("LeaseAgreement: Only spender or earner can raise dispute");
        leaseAgreement.raiseDispute(leaseId, "Test dispute");
        
        vm.stopPrank();
    }
    
    function testRaiseDisputeRevertWithEmptyReason() public {
        // Test: The raiseDispute function reverts with empty reason
        // First create a lease
        vm.prank(spender);
        leaseAgreement.createLease{value: minPrice}(earner, dataProductId, maxPrice);
        
        bytes32 leaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        
        // Try to raise dispute with empty reason
        vm.expectRevert("LeaseAgreement: Dispute reason cannot be empty");
        leaseAgreement.raiseDispute(leaseId, "");
    }
    
    function testGetLeaseRevertsWhenLeaseDoesNotExist() public {
        // Test: The getLease function reverts when lease does not exist
        bytes32 nonExistentLeaseId = keccak256("non-existent-lease");
        
        vm.expectRevert("LeaseAgreement: Lease does not exist");
        leaseAgreement.getLease(nonExistentLeaseId);
    }
    
    // ========== EDGE CASES ==========
    
    function testMinPriceConstantIsCorrect() public {
        // Test: The MIN_PRICE constant is set correctly
        assertEq(leaseAgreement.MIN_PRICE(), 0.001 ether);
    }
    
    function testLeaseCounterIncrements() public {
        // Test: Lease counter increments for each new lease
        vm.startPrank(spender);
        
        leaseAgreement.createLease{value: minPrice}(earner, dataProductId, maxPrice);
        leaseAgreement.createLease{value: minPrice}(earner, keccak256("product-2"), maxPrice);
        leaseAgreement.createLease{value: minPrice}(earner, keccak256("product-3"), maxPrice);
        
        vm.stopPrank();
        
        // Verify all leases exist with different IDs
        bytes32 leaseId1 = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        bytes32 leaseId2 = keccak256(abi.encodePacked(spender, earner, keccak256("product-2"), block.timestamp, 1));
        bytes32 leaseId3 = keccak256(abi.encodePacked(spender, earner, keccak256("product-3"), block.timestamp, 2));
        
        assertTrue(leaseAgreement.leaseExists(leaseId1));
        assertTrue(leaseAgreement.leaseExists(leaseId2));
        assertTrue(leaseAgreement.leaseExists(leaseId3));
    }
    
    function testSpenderCanRaiseDispute() public {
        // Test: Spender can raise dispute
        vm.prank(spender);
        leaseAgreement.createLease{value: minPrice}(earner, dataProductId, maxPrice);
        
        bytes32 leaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        
        vm.prank(spender);
        leaseAgreement.raiseDispute(leaseId, "Spender dispute");
        
        LeaseAgreement.Lease memory lease = leaseAgreement.getLease(leaseId);
        assertTrue(lease.isDisputed);
    }
    
    function testEarnerCanRaiseDispute() public {
        // Test: Earner can raise dispute
        vm.prank(spender);
        leaseAgreement.createLease{value: minPrice}(earner, dataProductId, maxPrice);
        
        bytes32 leaseId = keccak256(abi.encodePacked(spender, earner, dataProductId, block.timestamp, 0));
        
        vm.prank(earner);
        leaseAgreement.raiseDispute(leaseId, "Earner dispute");
        
        LeaseAgreement.Lease memory lease = leaseAgreement.getLease(leaseId);
        assertTrue(lease.isDisputed);
    }
} 