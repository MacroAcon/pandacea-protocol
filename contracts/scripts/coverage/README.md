# Foundry Coverage Checker

This directory contains tools for enforcing minimum coverage thresholds for the Pandacea Protocol smart contracts.

## Coverage Thresholds

- **Lines**: ≥95% (statements)
- **Branches**: ≥90% (conditional branches)

## Usage

### Manual Coverage Check

```bash
# From the contracts directory
forge coverage --report summary
python scripts/coverage/check_coverage.py
```

### Generate Detailed Coverage Report

```bash
# Generate LCOV format for detailed analysis
forge coverage --report lcov

# Generate HTML report (if lcov-genhtml is available)
genhtml lcov.info --output-directory coverage-html
```

### Makefile Integration

The coverage checker is integrated into the main Makefile:

```bash
# Run tests with verbose output
make contracts-test

# Run coverage check with thresholds
make contracts-coverage

# Run full verification (tests + coverage + other checks)
make verify
```

## Understanding Coverage

- **Lines**: Percentage of executable lines of code that are executed during tests
- **Branches**: Percentage of conditional branches (if/else, switch cases) that are tested
- **Functions**: Percentage of functions that are called during tests

## Improving Coverage

1. **Add test cases** for untested code paths
2. **Test edge cases** and error conditions
3. **Use fuzz testing** to discover unexpected code paths
4. **Review coverage reports** to identify gaps

## CI Integration

The coverage checker runs automatically in CI pipelines and will fail builds if thresholds are not met. This ensures code quality and prevents regressions.

## Troubleshooting

If coverage checks fail:

1. Run `forge coverage --report summary` to see current coverage
2. Review the detailed LCOV report to identify uncovered lines
3. Add test cases for missing coverage
4. Re-run the coverage check

For Windows users, ensure Foundry is properly installed and accessible in your PATH.
