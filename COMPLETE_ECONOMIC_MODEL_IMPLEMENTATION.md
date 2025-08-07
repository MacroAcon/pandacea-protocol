# Complete Economic Model Implementation

## Overview

This document describes the implementation of the complete economic model for the Pandacea Protocol, which includes:

1. **Automated "Just-in-Time" Reputation Decay**: Gas-efficient reputation decay that applies automatically during score updates
2. **Positive Reputation Rewards**: Mechanism to reward honest earners for successful, undisputed leases
3. **Stake-Based Disputes**: Economic disincentives for frivolous disputes
4. **Tiered Reputation System**: Reputation impact scaling with lease value

## Architecture

### Smart Contracts

#### 1. Reputation.sol - Automated Decay

**Key Changes:**
- Modified `ReputationData` struct to include `lastUpdatedTimestamp`
- Removed manual `decayReputation()` function
- Implemented `_applyDecayAndGetCurrentScore()` internal function
- Updated `updateReputation()` to apply decay before new score changes
- Added `getReputation()` view function with automatic decay calculation

**Core Functions:**

```solidity
function updateReputation(address earner, bool successfulLease, uint256 leaseValue) external onlyOwner
```

This function now:
1. Applies "just-in-time" decay based on time since last update
2. Calculates reputation change using tiered system
3. Updates score and timestamp
4. Emits `ReputationUpdated` event

```solidity
function getReputation(address earner) external view returns (uint256)
```

This view function returns the current reputation score with decay automatically applied.

**Decay Calculation:**
- Decay rate: 1 point per day
- Applied automatically when reputation is accessed or updated
- Formula: `decayAmount = daysPassed * REPUTATION_DECAY_RATE`

#### 2. LeaseAgreement.sol - Positive Rewards

**Key Changes:**
- Added `finalizeLease()` function for positive reputation rewards
- Added `DISPUTE_WINDOW` constant (7 days)
- Updated `Lease` struct with `isFinalized` and `executedAt` fields
- Added `LeaseFinalized` event
- Updated `raiseDispute()` to prevent disputes on finalized leases

**Core Functions:**

```solidity
function finalizeLease(bytes32 leaseId) external nonReentrant
```

This function:
1. Verifies only spender can call it
2. Checks lease is executed and not disputed
3. Ensures dispute window has passed (7 days)
4. Calls `Reputation.updateReputation(earner, true, lease.price)`
5. Marks lease as finalized
6. Emits `LeaseFinalized` event

**Access Control:**
- Only the spender can finalize a lease
- Lease must be executed but not disputed
- Dispute window must have passed

#### 3. PGT.sol - Governance Token

**Purpose:** ERC20 token used for staking in disputes

**Key Functions:**
- `mint()`: Owner-only function for testing
- `burn()`: User can burn their own tokens
- `burnFrom()`: Owner can burn user tokens

### Tiered Reputation System

**Reputation Impact by Lease Value:**

| Lease Value | Reputation Change |
|-------------|-------------------|
| < 1 ETH     | ±25 points       |
| 1-10 ETH    | ±50 points       |
| ≥ 10 ETH    | ±100 points      |

**Implementation:**
```solidity
function calculateReputationChange(uint256 leaseValue) internal pure returns (uint256) {
    if (leaseValue < 1e18) { return 25; }
    else if (leaseValue < 10e18) { return 50; }
    else { return 100; }
}
```

## Off-Chain Integration

### Python SDK

**New Method:**
```python
def finalize_lease(self, lease_id: str) -> str:
```

This method:
1. Converts lease_id to bytes32 format
2. Builds and signs the `finalizeLease` transaction
3. Sends transaction to blockchain
4. Returns transaction hash

**Usage:**
```python
# After lease execution and dispute window
tx_hash = client.finalize_lease(lease_id)
print(f"Lease finalized: {tx_hash}")
```

### Go Agent Backend

**API Endpoints:**
- `POST /api/v1/leases/{leaseId}/dispute` - Updated to include `stakeAmount`
- Smart contract interactions handled through Web3 integration

## Integration Testing

### Test Scenarios

#### 1. Successful Lease and Reward Test
- Creates and executes a lease
- Simulates time passing (dispute window)
- Calls `finalizeLease()`
- Verifies positive reputation reward

#### 2. Automated Decay Test
- Checks initial reputation
- Simulates time passing (30 days)
- Executes and finalizes new lease
- Verifies decay + reward calculation

#### 3. Finalize Disputed Lease Test
- Raises dispute on lease
- Attempts to call `finalizeLease()`
- Verifies transaction reverts

#### 4. Stake-Based Dispute Tests
- Valid dispute with stake return
- Invalid dispute with stake forfeiture
- Insufficient stake approval

### Running Tests

```bash
cd integration
python test_dispute_system.py
```

**Required Environment Variables:**
- `SPENDER_PRIVATE_KEY`
- `EARNER_PRIVATE_KEY`
- `CONTRACT_ADDRESS`
- `PGT_TOKEN_ADDRESS`
- `RPC_URL` (optional, defaults to localhost)

## Economic Model Benefits

### 1. Automated Decay
- **Gas Efficiency**: No need for periodic decay calls
- **Automatic**: Decay applied when reputation is accessed
- **Fair**: All users experience same decay rate
- **Transparent**: Decay calculation is deterministic

### 2. Positive Rewards
- **Incentivizes Quality**: Earners rewarded for successful work
- **Balanced**: Rewards scale with lease value
- **Dispute-Protected**: Only undisputed leases can be finalized
- **Time-Gated**: Dispute window prevents premature finalization

### 3. Stake-Based Disputes
- **Economic Disincentive**: Frivolous disputes cost tokens
- **Fair Distribution**: Invalid dispute stakes split between earner and DAO
- **Valid Protection**: Valid disputes return stake to spender
- **Reputation Integration**: Disputes affect earner reputation

### 4. Tiered System
- **Value-Aligned**: Higher-value leases have greater reputation impact
- **Risk-Adjusted**: Reflects the economic importance of transactions
- **Scalable**: System works for small and large transactions

## Security Considerations

### 1. Access Control
- Only spender can finalize their own leases
- Only owner can resolve disputes
- Only owner can update reputation

### 2. State Validation
- Lease must be executed before finalization
- Dispute window must pass before finalization
- Cannot dispute finalized leases

### 3. Economic Security
- Stake amounts must be approved before dispute
- Dispute resolution is owner-controlled
- Reputation changes are bounded (0-1000)

### 4. Reentrancy Protection
- All state-changing functions use `nonReentrant` modifier
- External calls made after state updates

## Deployment

### Contract Deployment Order
1. `PGT.sol` - Governance token
2. `Reputation.sol` - Reputation system
3. `LeaseAgreement.sol` - Lease management (with PGT and Reputation addresses)

### Configuration
- Set DAO treasury address in LeaseAgreement constructor
- Mint initial PGT tokens for testing
- Configure dispute window duration (7 days)
- Set reputation decay rate (1 point/day)

## Future Enhancements

### 1. Governance Integration
- DAO voting for dispute resolution
- Community-driven reputation adjustments
- Configurable parameters through governance

### 2. Advanced Reputation
- Reputation multipliers for consistent performance
- Specialized reputation categories
- Cross-chain reputation portability

### 3. Economic Optimizations
- Dynamic stake amounts based on lease value
- Automated dispute resolution for clear cases
- Reputation-based pricing models

## Conclusion

The complete economic model provides a robust foundation for the Pandacea Protocol's reputation and dispute resolution system. The automated decay mechanism ensures gas efficiency while the positive rewards create strong incentives for quality service. The stake-based dispute system provides economic disincentives for frivolous disputes while protecting legitimate concerns.

The tiered reputation system aligns reputation impact with economic value, creating a fair and scalable system that works for both small and large transactions. Together, these mechanisms create a self-sustaining economic ecosystem that rewards honest behavior and penalizes misconduct.
