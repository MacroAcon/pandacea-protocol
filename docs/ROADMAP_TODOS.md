# Pandacea Protocol - Roadmap TODOs

This document tracks TODO items that have been removed from the codebase in accordance with the project's Engineering Handbook, which specifies using GitHub Issues for task management.

## Purpose

- **No TODO comments in code**: All TODO items should be converted to GitHub Issues for proper tracking
- **Preserve context**: This file maintains a record of removed TODOs with their original context
- **Reference for development**: Developers can reference this file when creating GitHub Issues

## Removed TODO Items

### Policy Engine TODOs

**File**: `agent-backend/internal/policy/policy.go`  
**Line**: 91-97  
**Date Removed**: July 20, 2025  
**Original Content**:
```
// TODO: Future policy evaluation features to implement:
// - Reputation system integration
// - Data product availability checks
// - Rate limiting and abuse prevention
// - Compliance with regulatory requirements
// - Economic sustainability checks
```

**GitHub Issue**: [Create Issue](https://github.com/pandacea/pandacea-protocol/issues/new)  
**Suggested Labels**: `enhancement`, `policy-engine`, `future-feature`

**Description**: The policy engine currently implements basic Dynamic Minimum Pricing (DMP) validation. Future enhancements should include:
- Integration with a reputation system for trust scoring
- Validation of data product availability before lease approval
- Rate limiting mechanisms to prevent abuse
- Regulatory compliance checks (GDPR, CCPA, etc.)
- Economic sustainability validation beyond minimum pricing

---

## Guidelines for Future TODOs

1. **Create GitHub Issues**: Instead of adding TODO comments, create GitHub Issues with appropriate labels
2. **Use Descriptive Titles**: Issue titles should clearly describe the feature or fix needed
3. **Add Context**: Include relevant code references, requirements, and acceptance criteria
4. **Assign Labels**: Use appropriate labels for categorization (e.g., `bug`, `enhancement`, `documentation`)
5. **Link to Milestones**: Associate issues with project milestones when applicable

## Issue Templates

### Feature Request Template
```
## Feature Description
[Describe the feature or enhancement]

## Current Behavior
[Describe what currently happens]

## Desired Behavior
[Describe what should happen]

## Technical Requirements
- [ ] Requirement 1
- [ ] Requirement 2

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2

## Related Files
- `path/to/file.go`
- `path/to/file.py`

## Additional Context
[Any additional information]
```

### Bug Report Template
```
## Bug Description
[Describe the bug]

## Steps to Reproduce
1. Step 1
2. Step 2
3. Step 3

## Expected Behavior
[What should happen]

## Actual Behavior
[What actually happens]

## Environment
- OS: [e.g., Windows 10]
- Version: [e.g., v1.0.0]
- Component: [e.g., Agent Backend]

## Additional Context
[Any additional information]
```

---

*This document should be updated whenever TODO comments are removed from the codebase.* 