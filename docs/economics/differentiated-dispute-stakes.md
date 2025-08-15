# Differentiated Dispute Stakes Implementation

## Overview

This document describes the implementation of Differentiated Dispute Stakes for the Pandacea Protocol. The system enhances economic security by scaling dispute stakes proportionally with lease values, rather than using fixed amounts.

## Key Changes

### 1. Smart Contract Modifications (`contracts/src/LeaseAgreement.sol`)

#### New State Variables
- `disputeStakeRate`: Percentage of lease value required as stake (default: 10%)
- `DisputeStakeRateUpdated` event for tracking rate changes

#### Modified Functions

**`raiseDispute(bytes32 leaseId, string calldata reason)`**
- **Removed**: `stakeAmount` parameter
- **Added**: Dynamic stake calculation based on lease value and dispute stake rate
- **Formula**: `requiredStake = (lease.price * disputeStakeRate) / 100`

**`getRequiredStake(bytes32 leaseId)`** (New)
- Returns the calculated stake amount for a given lease
- Useful for frontend applications to show users the required stake before disputing

**`setDisputeStakeRate(uint256 newRate)`** (New)
- Allows DAO to adjust the stake rate (1-100%)
- Emits `DisputeStakeRateUpdated` event

### 2. Go Backend Changes (`agent-backend/internal/api/server.go`)

#### Modified Structs
```go
// Before
type DisputeRequest struct {
    Reason      string `json:"reason"`
    StakeAmount string `json:"stakeAmount"`
}

// After
type DisputeRequest struct {
    Reason string `json:"reason"`
}
```

#### Updated Handler
- Removed stake amount validation
- Updated logging to reflect dynamic stake calculation
- Added TODO comments for blockchain integration

### 3. Python SDK Changes (`builder-sdk/pandacea_sdk/client.py`)

#### New Methods
**`get_required_stake(lease_id: str) -> int`**
- Calls the smart contract's `getRequiredStake` function
- Returns the calculated stake amount in wei

#### Modified Methods
**`raise_dispute(lease_id: str, reason: str) -> str`**
- **Removed**: `stake_amount` parameter
- **Added**: Automatic stake calculation using `get_required_stake`
- **Added**: PGT approval for the calculated amount
- **Updated**: API payload to remove stake amount

### 4. Integration Tests (`integration/test_dispute_system.py`)

#### New Test Scenarios
1. **Low-Value Lease Dispute**: Tests 0.5 ETH lease with small stake
2. **High-Value Lease Dispute**: Tests 20 ETH lease with large stake
3. **Stake Rate Change**: Tests DAO's ability to change stake rates
4. **Invalid Dispute**: Tests stake forfeiture for invalid disputes
5. **Successful Lease**: Tests positive reputation rewards

#### Key Assertions
- Stake calculations match expected values (10% of lease value)
- Different lease values result in proportionally different stakes
- Stake rate changes affect all future calculations

## Economic Model

### Stake Calculation
```
Required Stake = (Lease Value × Dispute Stake Rate) / 100
```

### Examples
| Lease Value | Stake Rate | Required Stake |
|-------------|------------|----------------|
| 0.5 ETH     | 10%        | 0.05 ETH       |
| 1 ETH       | 10%        | 0.1 ETH        |
| 5 ETH       | 10%        | 0.5 ETH        |
| 20 ETH      | 10%        | 2 ETH          |

### Stake Rate Changes
If DAO changes rate from 10% to 20%:
| Lease Value | Old Stake (10%) | New Stake (20%) | Increase |
|-------------|-----------------|-----------------|----------|
| 0.5 ETH     | 0.05 ETH        | 0.1 ETH         | 2x       |
| 1 ETH       | 0.1 ETH         | 0.2 ETH         | 2x       |
| 5 ETH       | 0.5 ETH         | 1 ETH           | 2x       |
| 20 ETH      | 2 ETH           | 4 ETH           | 2x       |

## Usage Examples

### 1. Check Required Stake Before Disputing
```python
from pandacea_sdk.client import PandaceaClient

client = PandaceaClient(base_url="http://localhost:8080")
required_stake = client.get_required_stake(lease_id)
print(f"You need {required_stake} PGT tokens to dispute this lease")
```

### 2. Raise Dispute with Dynamic Stake
```python
# Old way (removed)
# dispute_id = client.raise_dispute(lease_id, reason, stake_amount)

# New way
dispute_id = client.raise_dispute(lease_id, "Data quality issues")
```

### 3. Go Backend API Call
```bash
# Old way (removed)
# curl -X POST /api/v1/leases/{leaseId}/dispute \
#   -d '{"reason": "Data quality issues", "stakeAmount": "1000000000000000000"}'

# New way
curl -X POST /api/v1/leases/{leaseId}/dispute \
  -d '{"reason": "Data quality issues"}'
```

## Benefits

### 1. Economic Security
- High-value leases require proportionally higher stakes
- Prevents frivolous disputes on expensive data
- Maintains accessibility for low-value disputes

### 2. DAO Governance
- Stake rates can be adjusted based on market conditions
- Allows fine-tuning of the dispute economy
- Transparent rate changes with event logging

### 3. User Experience
- Simplified API (no need to calculate stakes manually)
- Clear stake requirements before disputing
- Consistent economic model across all lease values

### 4. Scalability
- System automatically scales with lease values
- No need for manual stake amount calculations
- Future-proof for different market conditions

## Testing

### Running Integration Tests
```bash
cd integration
python test_dispute_system.py
```

### Running Example
```bash
cd builder-sdk/examples
python dynamic_staking_example.py
```

### Test Coverage
- ✅ Low-value lease disputes (0.5 ETH)
- ✅ High-value lease disputes (20 ETH)
- ✅ Stake rate changes (10% → 20%)
- ✅ Invalid dispute handling
- ✅ Successful lease completion
- ✅ Dynamic stake calculations
- ✅ API compatibility

## Migration Guide

### For Developers
1. **Update API calls**: Remove `stakeAmount` parameter from dispute requests
2. **Use new helper method**: Call `get_required_stake()` to show users stake requirements
3. **Update frontend**: Display calculated stakes instead of fixed amounts
4. **Test thoroughly**: Verify stake calculations match expected values

### For Users
1. **No action required**: The system automatically calculates required stakes
2. **Check requirements**: Use `get_required_stake()` to see dispute costs before proceeding
3. **Monitor changes**: Stake rates may be adjusted by the DAO over time

## Security Considerations

### 1. Rate Limiting
- Stake rate changes are restricted to contract owner (DAO)
- Rate must be between 1% and 100%
- Changes are logged with events

### 2. Stake Validation
- Calculated stakes must be greater than 0
- PGT allowance is verified before dispute creation
- Stake amounts are locked in the contract during disputes

### 3. Economic Attacks
- High-value leases require proportionally higher stakes
- Prevents spam disputes on expensive data
- Maintains economic balance across lease values

## Future Enhancements

### 1. Dynamic Rate Adjustment
- Automatic rate adjustment based on dispute frequency
- Market-driven stake rate optimization
- Time-based rate changes

### 2. Tiered Stake Rates
- Different rates for different lease value ranges
- Progressive stake scaling
- Risk-based rate calculation

### 3. Stake Insurance
- Optional stake insurance for high-value disputes
- Risk mitigation for legitimate disputes
- Community-funded dispute protection

## Conclusion

The Differentiated Dispute Stakes implementation successfully enhances the Pandacea Protocol's economic security by:

1. **Scaling dispute costs** with lease values
2. **Preventing frivolous disputes** on high-value data
3. **Maintaining accessibility** for low-value disputes
4. **Providing DAO governance** over stake rates
5. **Simplifying the user experience** with automatic calculations

The implementation is backward-compatible for new disputes while providing a more robust economic model for the protocol's long-term sustainability.
