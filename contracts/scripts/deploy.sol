// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "../src/Reputation.sol";
import "../src/LeaseAgreement.sol";
import "../src/PGT.sol";

/**
 * @title Deployment Script
 * @dev Script to deploy Reputation and LeaseAgreement contracts
 */
contract DeployScript is Script {
    
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        address deployer = vm.addr(deployerPrivateKey);
        
        vm.startBroadcast(deployerPrivateKey);
        
        console.log("Deploying contracts with deployer:", deployer);
        
        // Deploy PGT token contract first
        console.log("Deploying PGT token contract...");
        PGT pgtToken = new PGT();
        console.log("PGT token contract deployed at:", address(pgtToken));
        
        // Deploy Reputation contract
        console.log("Deploying Reputation contract...");
        Reputation reputation = new Reputation();
        console.log("Reputation contract deployed at:", address(reputation));
        
        // Deploy LeaseAgreement contract with all contract addresses
        console.log("Deploying LeaseAgreement contract...");
        LeaseAgreement leaseAgreement = new LeaseAgreement(
            address(reputation),
            address(pgtToken),
            deployer // Using deployer as DAO treasury for testing
        );
        console.log("LeaseAgreement contract deployed at:", address(leaseAgreement));
        
        // Initialize some test reputation scores (optional)
        console.log("Initializing test reputation scores...");
        reputation.initializeReputation(0x1234567890123456789012345678901234567890, 800);
        reputation.initializeReputation(0x2345678901234567890123456789012345678901, 750);
        reputation.initializeReputation(0x3456789012345678901234567890123456789012, 900);
        
        // Mint some PGT tokens for testing
        console.log("Minting test PGT tokens...");
        pgtToken.mint(0x1234567890123456789012345678901234567890, 1000e18); // 1000 PGT
        pgtToken.mint(0x2345678901234567890123456789012345678901, 1000e18); // 1000 PGT
        pgtToken.mint(0x3456789012345678901234567890123456789012, 1000e18); // 1000 PGT
        
        vm.stopBroadcast();
        
        console.log("Deployment completed successfully!");
        console.log("PGT token contract:", address(pgtToken));
        console.log("Reputation contract:", address(reputation));
        console.log("LeaseAgreement contract:", address(leaseAgreement));
    }
}
