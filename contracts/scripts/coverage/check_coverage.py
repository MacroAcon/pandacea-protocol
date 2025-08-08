#!/usr/bin/env python3
"""
Foundry Coverage Checker
Enforces minimum coverage thresholds for the Pandacea Protocol contracts.
"""

import subprocess
import sys
import re
from typing import Dict, Tuple

# Coverage thresholds
LINE_THRESHOLD = 95
BRANCH_THRESHOLD = 90

def run_forge_coverage() -> str:
    """Run forge coverage and return the summary output."""
    try:
        result = subprocess.run(
            ["forge", "coverage", "--report", "summary"],
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout
    except subprocess.CalledProcessError as e:
        print(f"❌ Error running forge coverage: {e}")
        print(f"stderr: {e.stderr}")
        sys.exit(1)
    except FileNotFoundError:
        print("❌ forge command not found. Please install Foundry first.")
        sys.exit(1)

def parse_coverage_summary(output: str) -> Dict[str, float]:
    """Parse the coverage summary output and extract percentages."""
    coverage = {}
    
    # Look for the summary table
    lines = output.split('\n')
    for line in lines:
        # Match lines like "Total | 95.2% | 90.1% | 85.3%"
        if 'Total' in line and '%' in line:
            parts = line.split('|')
            if len(parts) >= 4:
                try:
                    # Extract percentages (remove % and convert to float)
                    statements = float(parts[1].strip().replace('%', ''))
                    branches = float(parts[2].strip().replace('%', ''))
                    functions = float(parts[3].strip().replace('%', ''))
                    
                    coverage = {
                        'statements': statements,
                        'branches': branches,
                        'functions': functions
                    }
                    break
                except ValueError:
                    continue
    
    return coverage

def check_thresholds(coverage: Dict[str, float]) -> Tuple[bool, str]:
    """Check if coverage meets thresholds and return status with message."""
    if not coverage:
        return False, "Could not parse coverage data"
    
    issues = []
    
    # Check line coverage (statements)
    if coverage.get('statements', 0) < LINE_THRESHOLD:
        issues.append(f"Lines: {coverage.get('statements', 0):.1f}% < {LINE_THRESHOLD}%")
    
    # Check branch coverage
    if coverage.get('branches', 0) < BRANCH_THRESHOLD:
        issues.append(f"Branches: {coverage.get('branches', 0):.1f}% < {BRANCH_THRESHOLD}%")
    
    if issues:
        return False, "; ".join(issues)
    
    return True, "All thresholds met"

def print_coverage_table(coverage: Dict[str, float]):
    """Print a formatted coverage table."""
    print("\n📊 Coverage Summary:")
    print("=" * 50)
    print(f"{'Metric':<12} {'Current':<10} {'Threshold':<10} {'Status':<8}")
    print("-" * 50)
    
    if coverage:
        # Lines (statements)
        current_lines = coverage.get('statements', 0)
        status = "✅" if current_lines >= LINE_THRESHOLD else "❌"
        print(f"{'Lines':<12} {current_lines:<10.1f}% {LINE_THRESHOLD:<10}% {status:<8}")
        
        # Branches
        current_branches = coverage.get('branches', 0)
        status = "✅" if current_branches >= BRANCH_THRESHOLD else "❌"
        print(f"{'Branches':<12} {current_branches:<10.1f}% {BRANCH_THRESHOLD:<10}% {status:<8}")
        
        # Functions (informational)
        current_functions = coverage.get('functions', 0)
        print(f"{'Functions':<12} {current_functions:<10.1f}% {'N/A':<10} {'ℹ️':<8}")
    else:
        print("❌ No coverage data available")
    
    print("=" * 50)

def main():
    """Main function to run coverage check."""
    print("🔍 Running Foundry coverage check...")
    
    # Run forge coverage
    output = run_forge_coverage()
    
    # Parse coverage data
    coverage = parse_coverage_summary(output)
    
    # Print coverage table
    print_coverage_table(coverage)
    
    # Check thresholds
    passed, message = check_thresholds(coverage)
    
    if passed:
        print(f"\n✅ Coverage check passed: {message}")
        sys.exit(0)
    else:
        print(f"\n❌ Coverage check failed: {message}")
        print(f"\n💡 To improve coverage:")
        print(f"   - Add more test cases to cover missing lines")
        print(f"   - Add tests for conditional branches")
        print(f"   - Run 'forge coverage --report lcov' for detailed report")
        sys.exit(2)

if __name__ == "__main__":
    main()
