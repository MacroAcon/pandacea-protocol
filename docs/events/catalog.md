# Pandacea Protocol Event Catalog (v1.0)

This document catalogs all events emitted by Pandacea Protocol smart contracts. Events are used by indexers to track protocol state and provide real-time updates to applications.

## Event Format

All events follow the standard Ethereum event format:
- `event EventName(indexed type1 param1, type2 param2, ...)`
- Indexed parameters (up to 3) are searchable by indexers
- Non-indexed parameters are included in the event data but not searchable

## LeaseAgreement Contract Events

### LeaseCreated
**Contract**: `LeaseAgreement.sol`  
**Event Signature**: `LeaseCreated(bytes32 indexed leaseId, address indexed spender, address indexed earner, uint256 price)`

**Description**: Emitted when a new lease agreement is created between a spender and earner.

**Parameters**:
- `leaseId` (indexed): Unique identifier for the lease (bytes32)
- `spender` (indexed): Address of the data spender (address)
- `earner` (indexed): Address of the data earner (address)
- `price`: Price of the lease in wei (uint256)

**Topics**:
- Topic 0: `0x...` (event signature hash)
- Topic 1: `leaseId` (bytes32)
- Topic 2: `spender` (address)
- Topic 3: `earner` (address)

**Data**:
- `price` (uint256)

### LeaseApproved
**Contract**: `LeaseAgreement.sol`  
**Event Signature**: `LeaseApproved(bytes32 indexed leaseId)`

**Description**: Emitted when a lease is approved by the earner.

**Parameters**:
- `leaseId` (indexed): Unique identifier for the lease (bytes32)

**Topics**:
- Topic 0: `0x...` (event signature hash)
- Topic 1: `leaseId` (bytes32)

### LeaseExecuted
**Contract**: `LeaseAgreement.sol`  
**Event Signature**: `LeaseExecuted(bytes32 indexed leaseId)`

**Description**: Emitted when a lease is executed and computation begins.

**Parameters**:
- `leaseId` (indexed): Unique identifier for the lease (bytes32)

**Topics**:
- Topic 0: `0x...` (event signature hash)
- Topic 1: `leaseId` (bytes32)

### LeaseFinalized
**Contract**: `LeaseAgreement.sol`  
**Event Signature**: `LeaseFinalized(bytes32 indexed leaseId, address indexed earner, uint256 reputationReward)`

**Description**: Emitted when a lease is finalized and reputation rewards are distributed.

**Parameters**:
- `leaseId` (indexed): Unique identifier for the lease (bytes32)
- `earner` (indexed): Address of the earner receiving reputation reward (address)
- `reputationReward`: Amount of reputation points awarded (uint256)

**Topics**:
- Topic 0: `0x...` (event signature hash)
- Topic 1: `leaseId` (bytes32)
- Topic 2: `earner` (address)

**Data**:
- `reputationReward` (uint256)

### DisputeRaised
**Contract**: `LeaseAgreement.sol`  
**Event Signature**: `DisputeRaised(bytes32 indexed leaseId, address indexed spender, address indexed earner, string reason, uint256 stakeAmount)`

**Description**: Emitted when a dispute is raised against a lease.

**Parameters**:
- `leaseId` (indexed): Unique identifier for the lease (bytes32)
- `spender` (indexed): Address of the spender raising the dispute (address)
- `earner` (indexed): Address of the earner being disputed (address)
- `reason`: Reason for the dispute (string)
- `stakeAmount`: Amount of PGT tokens staked for the dispute (uint256)

**Topics**:
- Topic 0: `0x...` (event signature hash)
- Topic 1: `leaseId` (bytes32)
- Topic 2: `spender` (address)
- Topic 3: `earner` (address)

**Data**:
- `reason` (string)
- `stakeAmount` (uint256)

### DisputeResolved
**Contract**: `LeaseAgreement.sol`  
**Event Signature**: `DisputeResolved(bytes32 indexed leaseId, bool isDisputeValid, uint256 stakeAmount)`

**Description**: Emitted when a dispute is resolved.

**Parameters**:
- `leaseId` (indexed): Unique identifier for the lease (bytes32)
- `isDisputeValid`: Whether the dispute was found valid (bool)
- `stakeAmount`: Amount of PGT tokens distributed (uint256)

**Topics**:
- Topic 0: `0x...` (event signature hash)
- Topic 1: `leaseId` (bytes32)

**Data**:
- `isDisputeValid` (bool)
- `stakeAmount` (uint256)

### DisputeStakeRateUpdated
**Contract**: `LeaseAgreement.sol`  
**Event Signature**: `DisputeStakeRateUpdated(uint256 oldRate, uint256 newRate)`

**Description**: Emitted when the dispute stake rate is updated by the contract owner.

**Parameters**:
- `oldRate`: Previous dispute stake rate (uint256)
- `newRate`: New dispute stake rate (uint256)

**Topics**:
- Topic 0: `0x...` (event signature hash)

**Data**:
- `oldRate` (uint256)
- `newRate` (uint256)

## Reputation Contract Events

### ReputationAwarded
**Contract**: `Reputation.sol`  
**Event Signature**: `ReputationAwarded(address indexed user, uint256 amount, string reason)`

**Description**: Emitted when reputation points are awarded to a user.

**Parameters**:
- `user` (indexed): Address of the user receiving reputation (address)
- `amount`: Amount of reputation points awarded (uint256)
- `reason`: Reason for the reputation award (string)

**Topics**:
- Topic 0: `0x...` (event signature hash)
- Topic 1: `user` (address)

**Data**:
- `amount` (uint256)
- `reason` (string)

### ReputationDecayed
**Contract**: `Reputation.sol`  
**Event Signature**: `ReputationDecayed(address indexed user, uint256 amount, uint256 decayRate)`

**Description**: Emitted when reputation points decay for a user.

**Parameters**:
- `user` (indexed): Address of the user losing reputation (address)
- `amount`: Amount of reputation points decayed (uint256)
- `decayRate`: Rate at which reputation decayed (uint256)

**Topics**:
- Topic 0: `0x...` (event signature hash)
- Topic 1: `user` (address)

**Data**:
- `amount` (uint256)
- `decayRate` (uint256)

### ReputationThresholdUpdated
**Contract**: `Reputation.sol`  
**Event Signature**: `ReputationThresholdUpdated(uint256 oldThreshold, uint256 newThreshold)`

**Description**: Emitted when the reputation threshold is updated.

**Parameters**:
- `oldThreshold`: Previous reputation threshold (uint256)
- `newThreshold`: New reputation threshold (uint256)

**Topics**:
- Topic 0: `0x...` (event signature hash)

**Data**:
- `oldThreshold` (uint256)
- `newThreshold` (uint256)

## PGT Token Contract Events

### Transfer
**Contract**: `PGT.sol`  
**Event Signature**: `Transfer(address indexed from, address indexed to, uint256 value)`

**Description**: Standard ERC20 transfer event.

**Parameters**:
- `from` (indexed): Address sending tokens (address)
- `to` (indexed): Address receiving tokens (address)
- `value`: Amount of tokens transferred (uint256)

**Topics**:
- Topic 0: `0x...` (event signature hash)
- Topic 1: `from` (address)
- Topic 2: `to` (address)

**Data**:
- `value` (uint256)

### Approval
**Contract**: `PGT.sol`  
**Event Signature**: `Approval(address indexed owner, address indexed spender, uint256 value)`

**Description**: Standard ERC20 approval event.

**Parameters**:
- `owner` (indexed): Address granting approval (address)
- `spender` (indexed): Address receiving approval (address)
- `value`: Amount of tokens approved (uint256)

**Topics**:
- Topic 0: `0x...` (event signature hash)
- Topic 1: `owner` (address)
- Topic 2: `spender` (address)

**Data**:
- `value` (uint256)

### Staked
**Contract**: `PGT.sol`  
**Event Signature**: `Staked(address indexed user, uint256 amount, bytes32 indexed leaseId)`

**Description**: Emitted when PGT tokens are staked for a lease.

**Parameters**:
- `user` (indexed): Address of the user staking tokens (address)
- `amount`: Amount of tokens staked (uint256)
- `leaseId` (indexed): ID of the lease being staked for (bytes32)

**Topics**:
- Topic 0: `0x...` (event signature hash)
- Topic 1: `user` (address)
- Topic 2: `leaseId` (bytes32)

**Data**:
- `amount` (uint256)

### Unstaked
**Contract**: `PGT.sol`  
**Event Signature**: `Unstaked(address indexed user, uint256 amount, bytes32 indexed leaseId)`

**Description**: Emitted when PGT tokens are unstaked from a lease.

**Parameters**:
- `user` (indexed): Address of the user unstaking tokens (address)
- `amount`: Amount of tokens unstaked (uint256)
- `leaseId` (indexed): ID of the lease being unstaked from (bytes32)

**Topics**:
- Topic 0: `0x...` (event signature hash)
- Topic 1: `user` (address)
- Topic 2: `leaseId` (bytes32)

**Data**:
- `amount` (uint256)

## Event Indexing Guidelines

### Recommended Indexes
1. **Lease Events**: Index by `leaseId` for tracking individual lease lifecycles
2. **User Events**: Index by user address for tracking user activity
3. **Dispute Events**: Index by `leaseId` and user addresses for dispute tracking
4. **Reputation Events**: Index by user address for reputation tracking

### Filtering Strategies
- Use indexed parameters for efficient filtering
- Combine multiple indexed parameters for complex queries
- Use non-indexed parameters for additional data retrieval

### Performance Considerations
- Limit the number of indexed parameters per event (max 3)
- Use appropriate data types for indexed parameters
- Consider event frequency when designing indexes

## Version History

- **v1.0**: Initial event catalog for Sprint 1 production readiness
- All events are frozen and backward compatible
- Future versions will add new events without modifying existing ones
