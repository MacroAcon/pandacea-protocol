# Adversarial Economic Simulations

This directory contains adversarial economic simulations for the Pandacea Protocol, designed to test the robustness of the economic model against various attack vectors and malicious actors.

## Windows Quickstart

1. **Open PowerShell**
2. **Navigate to the project directory:**
   ```powershell
   cd C:\Users\thnxt\Documents\pandacea-protocol
   ```
3. **Create and activate virtual environment:**
   ```powershell
   python -m venv .venv
   .\.venv\Scripts\Activate.ps1
   ```
4. **Install simulation dependencies:**
   ```powershell
   pip install -r sims/requirements.txt
   ```
5. **Run simulations:**
   ```powershell
   make sims-run
   ```

## Simulation Overview

The adversarial simulation framework models various economic scenarios with different types of agents:

### Agent Types

- **HonestEarner**: Legitimate data contributors who follow protocol rules
- **Colluder**: Malicious actors who coordinate to manipulate the system
- **Griefer**: Actors who attempt to disrupt the system without direct profit
- **Hoarder**: Actors who accumulate resources to gain disproportionate influence

### Economic Mechanisms

- **Dispute Staking**: Differentiated stakes based on lease value
- **Reputation Decay**: Aggressive decay to prevent reputation farming
- **Stake-at-Risk**: Economic disincentives for malicious behavior

### Key Performance Indicators (KPIs)

- **Honest Share of Revenue**: Percentage of revenue earned by honest actors
- **Expected Loss for Honest**: Average losses suffered by honest participants
- **Stake-at-Risk Curves**: Relationship between stake and risk of loss
- **Liveness**: System's ability to continue operating under attack

## Configuration

Simulation parameters are configured in `sims/config/params.yaml`:

```yaml
stake_levels: [0.5, 1.0, 2.0]      # Different stake levels to test
reputation_decay: [0.01, 0.05, 0.1] # Reputation decay rates
collusion_size: [5, 20, 100]        # Number of colluding actors
sybil_cost: [0.1, 0.5, 1.0]        # Cost of creating Sybil identities
runs_per_point: 20                  # Number of runs per parameter combination
seed: 1337                          # Random seed for reproducibility
```

## Running Simulations

### Full Parameter Sweep

```powershell
# Run complete parameter sweep
make sims-run
```

This will:
1. Load parameters from `sims/config/params.yaml`
2. Execute grid sweep across all parameter combinations
3. Save results to CSV files in `sims/out/`
4. Generate summary statistics

### Custom Parameters

```powershell
# Run with custom parameters
python sims/run_sweeps.py --stake-levels 1.0 2.0 --runs-per-point 10
```

### Smoke Test

```powershell
# Quick smoke test with minimal parameters
python sims/run_sweeps.py --smoke-test
```

## Generating Reports

### HTML Report

```powershell
# Generate HTML report with plots
make sims-report
```

This creates `docs/economics/sim_report.html` with:
- Interactive plots and visualizations
- Statistical analysis of results
- Parameter sensitivity analysis
- Economic insights and recommendations

### Jupyter Notebook

```powershell
# Open Jupyter notebook for interactive analysis
jupyter notebook sims/plots/report.ipynb
```

## Output Files

### CSV Results

Simulation results are saved in `sims/out/`:

- `sweep_results.csv`: Raw simulation data
- `summary_stats.csv`: Aggregated statistics
- `parameter_sensitivity.csv`: Sensitivity analysis results

### Plots and Visualizations

Generated plots are saved in `docs/economics/assets/`:

- `honest_dominance_heatmap.png`: Heatmap of honest actor dominance
- `griefing_cost_vs_stake.png`: Relationship between griefing cost and stake
- `sensitivity_curves.png`: Parameter sensitivity analysis
- `revenue_distribution.png`: Distribution of revenue across agent types

## Analysis and Interpretation

### Key Metrics

1. **Honest Dominance**: Percentage of total revenue earned by honest actors
2. **Attack Cost**: Economic cost required to execute successful attacks
3. **System Resilience**: Ability to maintain operation under attack
4. **Parameter Sensitivity**: Which parameters most affect system security

### Economic Insights

The simulations help answer questions like:

- What stake levels provide adequate economic security?
- How does reputation decay affect attack profitability?
- What is the optimal balance between security and usability?
- How do different attack vectors compare in effectiveness?

## Customizing Simulations

### Adding New Agent Types

1. Create a new agent class in `sims/engine/model.py`
2. Implement the required interface methods
3. Add the agent type to the simulation configuration
4. Update the analysis and plotting code

### Adding New Economic Mechanisms

1. Extend the economic model in `sims/engine/model.py`
2. Add new parameters to `sims/config/params.yaml`
3. Update the simulation loop to include new mechanisms
4. Add relevant KPIs and analysis

### Custom Analysis

1. Modify `sims/plots/report.ipynb` for custom analysis
2. Add new plotting functions for specific insights
3. Export custom visualizations to `docs/economics/assets/`

## Troubleshooting

### Common Issues

1. **Memory Usage**: Large parameter sweeps can be memory-intensive
   - Reduce `runs_per_point` or parameter ranges
   - Use `--smoke-test` for quick validation

2. **Long Runtime**: Full sweeps can take hours
   - Use parallel processing with `--parallel`
   - Run subsets of parameters separately

3. **Missing Dependencies**: Ensure all requirements are installed
   ```powershell
   pip install -r sims/requirements.txt
   ```

### Performance Optimization

- Use `--parallel` flag for multi-core processing
- Reduce parameter ranges for faster iteration
- Use `--cache-results` to avoid recomputing identical scenarios

## Contributing

When adding new simulations:

1. **Documentation**: Update this README with new features
2. **Testing**: Add unit tests for new agent types and mechanisms
3. **Validation**: Ensure results are reproducible and statistically sound
4. **Analysis**: Provide clear interpretation of new results

## References

- [Economic Model Documentation](../docs/economics/simulation_report.md)
- [Protocol Whitepaper](../docs/Pandacea%20Protocol%20Technical%20Whitepaper%20v5.1.pdf)
- [Security Implementation](../docs/security/implementation.md)
