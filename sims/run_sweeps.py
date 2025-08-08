#!/usr/bin/env python3
"""
Adversarial Economic Simulation Runner

This script runs parameter sweeps across the economic simulation model
to test the robustness of the Pandacea Protocol against various attack vectors.
"""

import argparse
import itertools
import os
import sys
import yaml
import pandas as pd
import numpy as np
from pathlib import Path
from tqdm import tqdm
import random
from typing import Dict, List, Tuple

# Add the engine directory to the path
sys.path.append(str(Path(__file__).parent / "engine"))

from model import EconomicModel, SimulationResult, AgentType


def load_config(config_path: str = "sims/config/params.yaml") -> Dict:
    """Load simulation configuration from YAML file."""
    with open(config_path, 'r') as f:
        config = yaml.safe_load(f)
    return config


def create_output_directories(config: Dict):
    """Create output directories if they don't exist."""
    results_dir = Path(config['output']['results_dir'])
    plots_dir = Path(config['output']['plots_dir'])
    
    results_dir.mkdir(parents=True, exist_ok=True)
    plots_dir.mkdir(parents=True, exist_ok=True)


def run_single_simulation(
    stake_level: float,
    reputation_decay: float,
    collusion_size: int,
    sybil_cost: float,
    config: Dict,
    run_id: int
) -> Dict:
    """Run a single simulation with given parameters."""
    
    # Set random seed for reproducibility
    random.seed(config['seed'] + run_id)
    np.random.seed(config['seed'] + run_id)
    
    # Create economic model
    model = EconomicModel(config)
    
    # Initialize agents
    agent_distribution = {
        'total_agents': config['network']['total_agents'],
        'honest_percentage': config['network']['honest_percentage'],
        'colluder_percentage': config['network']['colluder_percentage'],
        'griefer_percentage': config['network']['griefer_percentage'],
        'hoarder_percentage': config['network']['hoarder_percentage'],
        'collusion_size': collusion_size
    }
    
    model.initialize_agents([stake_level], agent_distribution)
    
    # Update config with current parameters
    model.config['reputation_decay'] = reputation_decay
    model.config['attacks']['sybil_cost'] = sybil_cost
    
    # Run simulation epochs
    epochs = config['simulation']['epochs']
    warmup_epochs = config['simulation']['warmup_epochs']
    
    for epoch in range(epochs):
        epoch_result = model.run_epoch()
        
        # Skip warmup epochs in analysis
        if epoch < warmup_epochs:
            continue
    
    # Get final results
    result = model.get_simulation_result()
    
    # Return results with parameter information
    return {
        'run_id': run_id,
        'stake_level': stake_level,
        'reputation_decay': reputation_decay,
        'collusion_size': collusion_size,
        'sybil_cost': sybil_cost,
        'honest_share_of_revenue': result.honest_share_of_revenue,
        'expected_loss_for_honest': result.expected_loss_for_honest,
        'liveness_score': result.liveness_score,
        'collusion_detection_rate': result.collusion_detection_rate,
        'griefing_effectiveness': result.griefing_effectiveness,
        'hoarding_influence': result.hoarding_influence,
        'total_transactions': result.total_transactions,
        'successful_attacks': result.successful_attacks,
        'failed_attacks': result.failed_attacks,
        'total_revenue': model.total_revenue,
        'total_stake': model.total_stake,
        'dispute_resolutions': model.dispute_resolutions,
        'collusion_detections': model.collusion_detections
    }


def run_parameter_sweep(config: Dict, smoke_test: bool = False) -> pd.DataFrame:
    """Run parameter sweep across all combinations."""
    
    # Get parameter ranges
    if smoke_test:
        # Use minimal parameters for smoke test
        stake_levels = [config['stake_levels'][0]]
        reputation_decays = [config['reputation_decay'][0]]
        collusion_sizes = [config['collusion_size'][0]]
        sybil_costs = [config['sybil_cost'][0]]
        runs_per_point = 1
    else:
        stake_levels = config['stake_levels']
        reputation_decays = config['reputation_decay']
        collusion_sizes = config['collusion_size']
        sybil_costs = config['sybil_cost']
        runs_per_point = config['runs_per_point']
    
    # Generate all parameter combinations
    param_combinations = list(itertools.product(
        stake_levels, reputation_decays, collusion_sizes, sybil_costs
    ))
    
    print(f"Running {len(param_combinations)} parameter combinations")
    print(f"Total simulations: {len(param_combinations) * runs_per_point}")
    
    # Run simulations
    results = []
    
    for i, (stake_level, reputation_decay, collusion_size, sybil_cost) in enumerate(param_combinations):
        print(f"\nParameter combination {i+1}/{len(param_combinations)}:")
        print(f"  Stake level: {stake_level}")
        print(f"  Reputation decay: {reputation_decay}")
        print(f"  Collusion size: {collusion_size}")
        print(f"  Sybil cost: {sybil_cost}")
        
        # Run multiple times for this parameter combination
        for run_id in tqdm(range(runs_per_point), desc=f"Runs for combination {i+1}"):
            try:
                result = run_single_simulation(
                    stake_level, reputation_decay, collusion_size, sybil_cost,
                    config, run_id
                )
                results.append(result)
            except Exception as e:
                print(f"Error in simulation {run_id}: {e}")
                continue
    
    return pd.DataFrame(results)


def save_results(df: pd.DataFrame, config: Dict):
    """Save simulation results to CSV files."""
    results_dir = Path(config['output']['results_dir'])
    
    # Save raw results
    results_file = results_dir / "sweep_results.csv"
    df.to_csv(results_file, index=False)
    print(f"Raw results saved to: {results_file}")
    
    # Calculate and save summary statistics
    summary_stats = calculate_summary_statistics(df)
    summary_file = results_dir / "summary_stats.csv"
    summary_stats.to_csv(summary_file, index=False)
    print(f"Summary statistics saved to: {summary_file}")
    
    # Calculate and save parameter sensitivity
    sensitivity_stats = calculate_parameter_sensitivity(df)
    sensitivity_file = results_dir / "parameter_sensitivity.csv"
    sensitivity_stats.to_csv(sensitivity_file, index=False)
    print(f"Parameter sensitivity saved to: {sensitivity_file}")


def calculate_summary_statistics(df: pd.DataFrame) -> pd.DataFrame:
    """Calculate summary statistics across all runs."""
    summary_stats = []
    
    # Group by parameter combinations
    param_cols = ['stake_level', 'reputation_decay', 'collusion_size', 'sybil_cost']
    
    for name, group in df.groupby(param_cols):
        stats = {
            'stake_level': name[0],
            'reputation_decay': name[1],
            'collusion_size': name[2],
            'sybil_cost': name[3],
            'num_runs': len(group),
            'mean_honest_share': group['honest_share_of_revenue'].mean(),
            'std_honest_share': group['honest_share_of_revenue'].std(),
            'mean_expected_loss': group['expected_loss_for_honest'].mean(),
            'std_expected_loss': group['expected_loss_for_honest'].std(),
            'mean_liveness': group['liveness_score'].mean(),
            'std_liveness': group['liveness_score'].std(),
            'mean_collusion_detection': group['collusion_detection_rate'].mean(),
            'std_collusion_detection': group['collusion_detection_rate'].std(),
            'mean_griefing_effectiveness': group['griefing_effectiveness'].mean(),
            'std_griefing_effectiveness': group['griefing_effectiveness'].std(),
            'mean_hoarding_influence': group['hoarding_influence'].mean(),
            'std_hoarding_influence': group['hoarding_influence'].std(),
            'total_successful_attacks': group['successful_attacks'].sum(),
            'total_failed_attacks': group['failed_attacks'].sum(),
            'attack_success_rate': group['successful_attacks'].sum() / max(1, group['successful_attacks'].sum() + group['failed_attacks'].sum())
        }
        summary_stats.append(stats)
    
    return pd.DataFrame(summary_stats)


def calculate_parameter_sensitivity(df: pd.DataFrame) -> pd.DataFrame:
    """Calculate parameter sensitivity analysis."""
    sensitivity_stats = []
    
    # Calculate sensitivity for each parameter
    for param in ['stake_level', 'reputation_decay', 'collusion_size', 'sybil_cost']:
        for metric in ['honest_share_of_revenue', 'expected_loss_for_honest', 'liveness_score']:
            # Group by parameter value and calculate mean
            grouped = df.groupby(param)[metric].mean()
            
            # Calculate sensitivity (change in metric per unit change in parameter)
            if len(grouped) > 1:
                param_values = grouped.index.values
                metric_values = grouped.values
                
                # Calculate sensitivity as slope of linear fit
                if len(param_values) > 1:
                    slope = np.polyfit(param_values, metric_values, 1)[0]
                    
                    sensitivity_stats.append({
                        'parameter': param,
                        'metric': metric,
                        'sensitivity': slope,
                        'min_value': param_values.min(),
                        'max_value': param_values.max(),
                        'min_metric': metric_values.min(),
                        'max_metric': metric_values.max(),
                        'metric_range': metric_values.max() - metric_values.min()
                    })
    
    return pd.DataFrame(sensitivity_stats)


def main():
    """Main function to run parameter sweeps."""
    parser = argparse.ArgumentParser(description='Run adversarial economic simulations')
    parser.add_argument('--config', default='sims/config/params.yaml', help='Configuration file path')
    parser.add_argument('--smoke-test', action='store_true', help='Run smoke test with minimal parameters')
    parser.add_argument('--stake-levels', nargs='+', type=float, help='Custom stake levels to test')
    parser.add_argument('--runs-per-point', type=int, help='Number of runs per parameter combination')
    parser.add_argument('--parallel', action='store_true', help='Enable parallel processing')
    
    args = parser.parse_args()
    
    # Load configuration
    config = load_config(args.config)
    
    # Override parameters if provided
    if args.stake_levels:
        config['stake_levels'] = args.stake_levels
    if args.runs_per_point:
        config['runs_per_point'] = args.runs_per_point
    
    # Create output directories
    create_output_directories(config)
    
    print("Starting adversarial economic simulation...")
    print(f"Configuration: {args.config}")
    print(f"Smoke test: {args.smoke_test}")
    
    # Run parameter sweep
    results_df = run_parameter_sweep(config, smoke_test=args.smoke_test)
    
    # Save results
    save_results(results_df, config)
    
    # Print summary
    print("\n" + "="*50)
    print("SIMULATION SUMMARY")
    print("="*50)
    print(f"Total parameter combinations: {len(results_df.groupby(['stake_level', 'reputation_decay', 'collusion_size', 'sybil_cost']))}")
    print(f"Total simulation runs: {len(results_df)}")
    print(f"Average honest share of revenue: {results_df['honest_share_of_revenue'].mean():.3f}")
    print(f"Average expected loss for honest: {results_df['expected_loss_for_honest'].mean():.3f}")
    print(f"Average liveness score: {results_df['liveness_score'].mean():.3f}")
    print(f"Total successful attacks: {results_df['successful_attacks'].sum()}")
    print(f"Total failed attacks: {results_df['failed_attacks'].sum()}")
    print(f"Attack success rate: {results_df['successful_attacks'].sum() / max(1, results_df['successful_attacks'].sum() + results_df['failed_attacks'].sum()):.3f}")
    
    print("\nResults saved to sims/out/")
    print("Run 'make sims-report' to generate visualizations")


if __name__ == "__main__":
    main()
