# On-Chain Reputation and Dispute Resolution System

## Overview

This document describes the implementation of the on-chain reputation and dispute resolution system for the Pandacea Protocol. The system allows Spenders to penalize Earners for providing low-quality data through a decentralized dispute mechanism.

## Architecture

### Smart Contracts

#### 1. Reputation.sol

The core reputation management contract that tracks Earners' reputation scores and handles dispute resolution.

**Key Features:**
- Reputation score tracking (0-1000 scale)
- Dispute creation and resolution
- Reputation decay mechanism (1 point per day)
- Dispute cooldown periods (7 days)
- Automatic reputation penalties for valid disputes

**Key Functions:**
- `initializeReputation(address earner, uint256 initialScore)` - Initialize reputation for new earners
- `updateReputation(address earner, uint256 newScore, string reason)` - Update reputation scores
- `raiseDispute(address earner, uint256 leaseId, string reason)` - Create a new dispute
- `resolveDispute(uint256 disputeId, bool inFavorOfSpender)` - Resolve disputes
- `applyDecay(address earner)` - Apply time-based reputation decay
- `getReputationData(address earner)` - Get comprehensive reputation data

**Events:**
- `ReputationUpdated` - Emitted when reputation is updated
- `DisputeRaised` - Emitted when a dispute is created
- `ReputationDecay` - Emitted when reputation decays

#### 2. LeaseAgreement.sol (Modified)

Enhanced lease agreement contract that integrates with the reputation system.

**Key Modifications:**
- Added `Reputation` contract reference
- Enhanced `Lease` struct with `disputeId` field
- Modified `raiseDispute()` function to integrate with reputation system
- Added `getDisputeInfo()` function to retrieve dispute information
- Added `updateReputationContract()` function for contract upgrades

**Integration Points:**
- When a Spender raises a dispute, it automatically calls the Reputation contract
- Dispute IDs are tracked and linked between contracts
- Lease status is updated to reflect dispute state

### Backend Integration

#### Go Agent Backend

**New API Endpoint:**
```
POST /api/v1/leases/{leaseId}/dispute
```

**Request Body:**
```json
{
  "reason": "Data quality issues: Incomplete or inaccurate data provided"
}
```

**Response:**
```json
{
  "disputeId": "dispute_lease123_1640995200",
  "status": "pending"
}
```

**Implementation Details:**
- Added `DisputeRequest` and `DisputeResponse` structs
- Implemented `handleRaiseDispute()` handler function
- Added route to the API router
- Integrated with signature verification middleware

#### Python SDK

**New Method:**
```python
def raise_dispute(self, lease_id: str, reason: str) -> str:
    """
    Raise a dispute against an earner for a specific lease.
    
    Args:
        lease_id: The on-chain ID of the lease to dispute.
        reason: The reason for the dispute.
        
    Returns:
        The dispute ID for tracking the dispute.
    """
```

**Features:**
- Full integration with existing authentication system
- Proper error handling and validation
- Consistent API response parsing
- Support for all existing exception types

## Deployment

### Contract Deployment

**Deployment Script:**
```solidity
// contracts/scripts/deploy.sol
contract DeployScript is Script {
    function run() external {
        // Deploy Reputation contract first
        Reputation reputation = new Reputation();
        
        // Deploy LeaseAgreement contract with Reputation address
        LeaseAgreement leaseAgreement = new LeaseAgreement(address(reputation));
        
        // Initialize test reputation scores
        reputation.initializeReputation(0x1234..., 800);
    }
}
```

**Deployment Commands:**
```bash
# Set environment variables
export PRIVATE_KEY="your_private_key"
export RPC_URL="your_rpc_url"

# Deploy contracts
forge script scripts/deploy.sol --rpc-url $RPC_URL --broadcast
```

### Configuration

**Environment Variables:**
```bash
# Required for dispute system
CONTRACT_ADDRESS="deployed_lease_agreement_address"
RPC_URL="polygon_rpc_url"
SPENDER_PRIVATE_KEY="path_to_spender_key"
EARNER_PRIVATE_KEY="path_to_earner_key"
```

## Usage Examples

### 1. Raising a Dispute via SDK

```python
from pandacea_sdk.client import PandaceaClient

# Initialize client
client = PandaceaClient(
    base_url="http://localhost:8080",
    private_key_path="spender_key.pem"
)

# Raise a dispute
dispute_id = client.raise_dispute(
    lease_id="0x1234567890abcdef...",
    reason="Data quality issues: Incomplete or inaccurate data provided"
)

print(f"Dispute raised with ID: {dispute_id}")
```

### 2. Direct Smart Contract Interaction

```javascript
// Using ethers.js
const reputationContract = new ethers.Contract(reputationAddress, reputationABI, signer);

// Raise a dispute
const tx = await reputationContract.raiseDispute(
    earnerAddress,
    leaseId,
    "Data quality issues"
);
await tx.wait();

// Get dispute information
const dispute = await reputationContract.getDispute(0);
console.log("Dispute:", dispute);
```

### 3. Checking Reputation Data

```javascript
// Get reputation data for an earner
const reputationData = await reputationContract.getReputationData(earnerAddress);
console.log("Reputation score:", reputationData.score);
console.log("Total disputes:", reputationData.totalDisputes);
console.log("Resolved disputes:", reputationData.resolvedDisputes);
```

## Testing

### Integration Test

**File:** `integration/test_dispute_system.py`

**Test Coverage:**
1. Lease creation and approval
2. Dispute raising and verification
3. Reputation penalty calculation
4. Lease status verification
5. Dispute cooldown testing

**Running the Test:**
```bash
# Set up environment
export AGENT_URL="http://localhost:8080"
export SPENDER_PRIVATE_KEY="spender_key.pem"
export EARNER_PRIVATE_KEY="earner_key.pem"
export CONTRACT_ADDRESS="0x..."
export RPC_URL="http://127.0.0.1:8545"

# Run the test
python integration/test_dispute_system.py
```

### Unit Tests

**Contract Tests:**
```bash
# Run Foundry tests
forge test --match-contract Reputation
forge test --match-contract LeaseAgreement
```

## Security Considerations

### Access Control
- Only contract owner can initialize and update reputation scores
- Only contract owner can resolve disputes
- Earners cannot dispute themselves
- Dispute cooldown prevents spam disputes

### Reputation Protection
- Reputation scores are bounded (0-1000)
- Decay mechanism prevents score inflation
- Dispute penalties are capped
- Active status can be revoked by owner

### Gas Optimization
- Efficient dispute storage and retrieval
- Minimal state changes during dispute creation
- Batch operations for reputation updates

## Economic Model

### Reputation Scoring
- **Initial Score:** 800 (configurable)
- **Maximum Score:** 1000
- **Minimum Score:** 0
- **Decay Rate:** 1 point per day
- **Dispute Penalty:** 50 points

### Dispute Economics
- **Cooldown Period:** 7 days between disputes
- **Resolution Time:** Configurable by governance
- **Penalty Application:** Automatic on valid disputes
- **Score Recovery:** Through positive interactions

## Governance and Upgrades

### Contract Upgrades
- Reputation contract can be upgraded via `updateReputationContract()`
- Lease agreement maintains backward compatibility
- Dispute history is preserved across upgrades

### Parameter Updates
- Reputation decay rate can be adjusted
- Dispute penalties can be modified
- Cooldown periods can be updated
- All changes require owner approval

## Future Enhancements

### Planned Features
1. **Automated Dispute Resolution:** AI-powered dispute assessment
2. **Reputation Recovery Mechanisms:** Ways for earners to improve scores
3. **Dispute Arbitration:** Multi-party dispute resolution
4. **Reputation Transfer:** Cross-chain reputation portability
5. **Advanced Analytics:** Reputation trend analysis and insights

### Integration Opportunities
1. **Oracle Integration:** External data for dispute validation
2. **Cross-Chain Reputation:** Reputation sharing across networks
3. **DeFi Integration:** Reputation-based lending and insurance
4. **DAO Governance:** Community-driven dispute resolution

## Troubleshooting

### Common Issues

**1. Dispute Creation Fails**
- Check if earner is active
- Verify cooldown period has passed
- Ensure lease exists and is valid

**2. Reputation Not Updated**
- Verify contract owner permissions
- Check for transaction failures
- Ensure proper contract addresses

**3. API Errors**
- Validate request signatures
- Check authentication headers
- Verify lease ID format

### Debug Commands

```bash
# Check contract state
cast call $REPUTATION_ADDRESS "getReputationData(address)" $EARNER_ADDRESS

# Check dispute count
cast call $REPUTATION_ADDRESS "getDisputeCount()"

# Verify lease status
cast call $LEASE_ADDRESS "getLease(bytes32)" $LEASE_ID
```

## Conclusion

The on-chain reputation and dispute resolution system provides a robust foundation for maintaining data quality in the Pandacea Protocol. Through transparent dispute mechanisms, automatic reputation penalties, and comprehensive tracking, the system incentivizes high-quality data provision while protecting both Earners and Spenders.

The implementation is designed to be scalable, secure, and upgradeable, ensuring the protocol can evolve with the needs of the data economy while maintaining trust and transparency.
