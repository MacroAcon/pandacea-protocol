"""
Adversarial Economic Simulation Engine

This module implements the core simulation model for testing the Pandacea Protocol's
economic security against various attack vectors and malicious actors.
"""

import numpy as np
import pandas as pd
from abc import ABC, abstractmethod
from typing import Dict, List, Tuple, Optional
from dataclasses import dataclass
import random
from enum import Enum


class AgentType(Enum):
    """Types of agents in the simulation."""
    HONEST = "honest"
    COLLUDER = "colluder"
    GRIEFER = "griefer"
    HOARDER = "hoarder"


@dataclass
class AgentState:
    """State of an individual agent."""
    agent_id: int
    agent_type: AgentType
    stake: float
    reputation: float
    balance: float
    total_revenue: float
    total_losses: float
    disputes_won: int
    disputes_lost: int
    is_active: bool = True


@dataclass
class SimulationResult:
    """Results from a single simulation run."""
    honest_share_of_revenue: float
    expected_loss_for_honest: float
    stake_at_risk_curves: Dict[float, float]
    liveness_score: float
    collusion_detection_rate: float
    griefing_effectiveness: float
    hoarding_influence: float
    total_transactions: int
    successful_attacks: int
    failed_attacks: int


class Agent(ABC):
    """Abstract base class for all agents in the simulation."""
    
    def __init__(self, agent_id: int, initial_stake: float, initial_reputation: float = 1.0):
        self.agent_id = agent_id
        self.stake = initial_stake
        self.reputation = initial_reputation
        self.balance = 0.0
        self.total_revenue = 0.0
        self.total_losses = 0.0
        self.disputes_won = 0
        self.disputes_lost = 0
        self.is_active = True
        
    @abstractmethod
    def decide_action(self, epoch: int, network_state: Dict) -> Dict:
        """Decide what action to take in the current epoch."""
        pass
    
    @abstractmethod
    def update_state(self, reward: float, penalty: float, reputation_change: float):
        """Update agent state based on outcomes."""
        pass
    
    def get_stake_at_risk(self) -> float:
        """Calculate the stake at risk for this agent."""
        return self.stake * (1 - self.reputation)
    
    def is_solvent(self) -> bool:
        """Check if the agent has sufficient stake to continue participating."""
        return self.stake >= 0.1  # Minimum stake requirement


class HonestEarner(Agent):
    """Legitimate data contributors who follow protocol rules."""
    
    def __init__(self, agent_id: int, initial_stake: float, initial_reputation: float = 1.0):
        super().__init__(agent_id, initial_stake, initial_reputation)
        self.agent_type = AgentType.HONEST
    
    def decide_action(self, epoch: int, network_state: Dict) -> Dict:
        """Honest agents always contribute data and follow protocol rules."""
        return {
            'action': 'contribute',
            'stake_committed': min(self.stake * 0.1, 1.0),  # Conservative stake usage
            'quality': random.uniform(0.8, 1.0),  # High quality contributions
            'collude': False,
            'grief': False,
            'hoard': False
        }
    
    def update_state(self, reward: float, penalty: float, reputation_change: float):
        """Update honest agent state."""
        self.balance += reward - penalty
        self.total_revenue += reward
        self.total_losses += penalty
        self.reputation = max(0.0, min(1.0, self.reputation + reputation_change))
        
        if reward > 0:
            # Honest agents may increase their stake with profits
            self.stake += reward * 0.1


class Colluder(Agent):
    """Malicious actors who coordinate to manipulate the system."""
    
    def __init__(self, agent_id: int, initial_stake: float, collusion_group: List[int]):
        super().__init__(agent_id, initial_stake, initial_reputation=0.8)
        self.agent_type = AgentType.COLLUDER
        self.collusion_group = collusion_group
        self.collusion_coordination = 0.0
    
    def decide_action(self, epoch: int, network_state: Dict) -> Dict:
        """Colluders coordinate attacks with their group."""
        # Coordinate with other colluders
        if self.collusion_group:
            self.collusion_coordination = min(1.0, self.collusion_coordination + 0.1)
        
        # Decide whether to collude based on coordination level
        will_collude = random.random() < self.collusion_coordination
        
        return {
            'action': 'contribute' if not will_collude else 'collude',
            'stake_committed': self.stake * 0.3 if will_collude else self.stake * 0.1,
            'quality': random.uniform(0.3, 0.7) if will_collude else random.uniform(0.6, 0.9),
            'collude': will_collude,
            'grief': False,
            'hoard': False,
            'collusion_group': self.collusion_group if will_collude else []
        }
    
    def update_state(self, reward: float, penalty: float, reputation_change: float):
        """Update colluder state."""
        self.balance += reward - penalty
        self.total_revenue += reward
        self.total_losses += penalty
        self.reputation = max(0.0, min(1.0, self.reputation + reputation_change))
        
        # Colluders may lose coordination if detected
        if penalty > reward:
            self.collusion_coordination = max(0.0, self.collusion_coordination - 0.2)


class Griefer(Agent):
    """Actors who attempt to disrupt the system without direct profit."""
    
    def __init__(self, agent_id: int, initial_stake: float):
        super().__init__(agent_id, initial_stake, initial_reputation=0.6)
        self.agent_type = AgentType.GRIEFER
        self.griefing_intensity = random.uniform(0.5, 1.0)
    
    def decide_action(self, epoch: int, network_state: Dict) -> Dict:
        """Griefers attempt to disrupt the system."""
        # Griefing probability increases with system stability
        grief_prob = self.griefing_intensity * (1 - network_state.get('liveness', 0.5))
        
        will_grief = random.random() < grief_prob
        
        return {
            'action': 'contribute' if not will_grief else 'grief',
            'stake_committed': self.stake * 0.5 if will_grief else self.stake * 0.1,
            'quality': random.uniform(0.1, 0.4) if will_grief else random.uniform(0.5, 0.8),
            'collude': False,
            'grief': will_grief,
            'hoard': False
        }
    
    def update_state(self, reward: float, penalty: float, reputation_change: float):
        """Update griefer state."""
        self.balance += reward - penalty
        self.total_revenue += reward
        self.total_losses += penalty
        self.reputation = max(0.0, min(1.0, self.reputation + reputation_change))
        
        # Griefing intensity may change based on effectiveness
        if penalty > reward:
            self.griefing_intensity = min(1.0, self.griefing_intensity + 0.1)
        else:
            self.griefing_intensity = max(0.1, self.griefing_intensity - 0.05)


class Hoarder(Agent):
    """Actors who accumulate resources to gain disproportionate influence."""
    
    def __init__(self, agent_id: int, initial_stake: float):
        super().__init__(agent_id, initial_stake, initial_reputation=0.9)
        self.agent_type = AgentType.HOARDER
        self.hoarding_target = initial_stake * 5  # Target 5x initial stake
        self.hoarding_strategy = 'accumulate'
    
    def decide_action(self, epoch: int, network_state: Dict) -> Dict:
        """Hoarders accumulate resources strategically."""
        # Switch to influence mode if target reached
        if self.stake >= self.hoarding_target:
            self.hoarding_strategy = 'influence'
        
        if self.hoarding_strategy == 'accumulate':
            # Conservative participation to accumulate stake
            return {
                'action': 'contribute',
                'stake_committed': self.stake * 0.05,  # Very conservative
                'quality': random.uniform(0.7, 0.9),  # Good quality to avoid penalties
                'collude': False,
                'grief': False,
                'hoard': True
            }
        else:
            # Use accumulated influence
            return {
                'action': 'influence',
                'stake_committed': self.stake * 0.3,
                'quality': random.uniform(0.6, 0.8),
                'collude': False,
                'grief': False,
                'hoard': True
            }
    
    def update_state(self, reward: float, penalty: float, reputation_change: float):
        """Update hoarder state."""
        self.balance += reward - penalty
        self.total_revenue += reward
        self.total_losses += penalty
        self.reputation = max(0.0, min(1.0, self.reputation + reputation_change))
        
        # Reinvest most profits into stake
        if reward > penalty:
            self.stake += (reward - penalty) * 0.8


class EconomicModel:
    """Core economic model for the Pandacea Protocol simulation."""
    
    def __init__(self, config: Dict):
        self.config = config
        self.agents: List[Agent] = []
        self.epoch = 0
        self.total_revenue = 0.0
        self.total_stake = 0.0
        self.dispute_resolutions = 0
        self.collusion_detections = 0
        
    def initialize_agents(self, stake_levels: List[float], agent_distribution: Dict):
        """Initialize the agent population."""
        agent_id = 0
        
        # Create honest agents
        num_honest = int(agent_distribution['total_agents'] * agent_distribution['honest_percentage'])
        for i in range(num_honest):
            stake = random.choice(stake_levels)
            agent = HonestEarner(agent_id, stake)
            self.agents.append(agent)
            agent_id += 1
        
        # Create colluding agents
        num_colluders = int(agent_distribution['total_agents'] * agent_distribution['colluder_percentage'])
        collusion_groups = self._create_collusion_groups(num_colluders, agent_distribution.get('collusion_size', 5))
        for i, group in enumerate(collusion_groups):
            stake = random.choice(stake_levels)
            agent = Colluder(agent_id, stake, group)
            self.agents.append(agent)
            agent_id += 1
        
        # Create griefing agents
        num_griefers = int(agent_distribution['total_agents'] * agent_distribution['griefer_percentage'])
        for i in range(num_griefers):
            stake = random.choice(stake_levels)
            agent = Griefer(agent_id, stake)
            self.agents.append(agent)
            agent_id += 1
        
        # Create hoarding agents
        num_hoarders = int(agent_distribution['total_agents'] * agent_distribution['hoarder_percentage'])
        for i in range(num_hoarders):
            stake = random.choice(stake_levels)
            agent = Hoarder(agent_id, stake)
            self.agents.append(agent)
            agent_id += 1
        
        self._update_total_stake()
    
    def _create_collusion_groups(self, num_colluders: int, group_size: int) -> List[List[int]]:
        """Create collusion groups for coordinated attacks."""
        groups = []
        colluder_ids = list(range(num_colluders))
        random.shuffle(colluder_ids)
        
        for i in range(0, num_colluders, group_size):
            group = colluder_ids[i:i + group_size]
            if len(group) >= 2:  # Minimum group size
                groups.append(group)
        
        return groups
    
    def _update_total_stake(self):
        """Update total stake in the system."""
        self.total_stake = sum(agent.stake for agent in self.agents if agent.is_active)
    
    def run_epoch(self) -> Dict:
        """Run a single epoch of the simulation."""
        self.epoch += 1
        
        # Get network state for agent decisions
        network_state = self._get_network_state()
        
        # Collect actions from all agents
        actions = []
        for agent in self.agents:
            if agent.is_active and agent.is_solvent():
                action = agent.decide_action(self.epoch, network_state)
                action['agent_id'] = agent.agent_id
                action['agent_type'] = agent.agent_type
                actions.append(action)
        
        # Process actions and calculate outcomes
        outcomes = self._process_actions(actions)
        
        # Update agent states
        for outcome in outcomes:
            agent = self.agents[outcome['agent_id']]
            agent.update_state(
                outcome['reward'],
                outcome['penalty'],
                outcome['reputation_change']
            )
        
        # Apply reputation decay
        self._apply_reputation_decay()
        
        # Update system state
        self._update_total_stake()
        
        return {
            'epoch': self.epoch,
            'total_revenue': self.total_revenue,
            'total_stake': self.total_stake,
            'active_agents': len([a for a in self.agents if a.is_active]),
            'outcomes': outcomes
        }
    
    def _get_network_state(self) -> Dict:
        """Get current network state for agent decision making."""
        active_agents = [a for a in self.agents if a.is_active]
        
        return {
            'total_agents': len(active_agents),
            'total_stake': self.total_stake,
            'average_reputation': np.mean([a.reputation for a in active_agents]),
            'liveness': len(active_agents) / len(self.agents),
            'epoch': self.epoch
        }
    
    def _process_actions(self, actions: List[Dict]) -> List[Dict]:
        """Process agent actions and calculate outcomes."""
        outcomes = []
        
        for action in actions:
            agent = self.agents[action['agent_id']]
            
            # Base reward calculation
            base_reward = self.config['economic']['base_reward']
            quality_multiplier = action['quality']
            reputation_multiplier = 1 + (agent.reputation - 0.5) * self.config['economic']['reputation_multiplier']
            
            reward = base_reward * quality_multiplier * reputation_multiplier
            
            # Penalty calculation
            penalty = 0.0
            reputation_change = 0.0
            
            # Handle different action types
            if action['collude']:
                # Collusion detection and penalties
                detection_prob = self.config['attacks']['collusion_detection_prob']
                if random.random() < detection_prob:
                    penalty = reward * self.config['attacks']['collusion_penalty_multiplier']
                    reputation_change = -0.3
                    self.collusion_detections += 1
                else:
                    # Undetected collusion may still succeed
                    reward *= 1.5  # Collusion bonus
            
            elif action['grief']:
                # Griefing attacks
                griefing_cost = self.config['attacks']['griefing_cost']
                griefing_effectiveness = self.config['attacks']['griefing_effectiveness']
                
                penalty = griefing_cost
                reputation_change = -0.2
                
                # Griefing may reduce system liveness
                if random.random() < griefing_effectiveness:
                    # Successful griefing attack
                    pass
            
            elif action['hoard']:
                # Hoarding behavior
                hoarding_cost = self.config['attacks']['hoarding_cost']
                hoarding_efficiency = self.config['attacks']['hoarding_efficiency']
                
                penalty = hoarding_cost
                
                # Hoarding may increase influence
                if random.random() < hoarding_efficiency:
                    reputation_change = 0.1
            
            # Dispute resolution (simplified)
            if random.random() < 0.1:  # 10% chance of dispute
                dispute_cost = self.config['economic']['dispute_cost']
                if random.random() < agent.reputation:  # Higher reputation = more likely to win
                    agent.disputes_won += 1
                    reward += dispute_cost
                else:
                    agent.disputes_lost += 1
                    penalty += self.config['economic']['dispute_penalty']
                    reputation_change -= 0.1
                
                self.dispute_resolutions += 1
            
            # Update total revenue
            self.total_revenue += reward
            
            outcomes.append({
                'agent_id': action['agent_id'],
                'agent_type': action['agent_type'],
                'reward': reward,
                'penalty': penalty,
                'reputation_change': reputation_change,
                'action': action['action']
            })
        
        return outcomes
    
    def _apply_reputation_decay(self):
        """Apply reputation decay to all agents."""
        decay_rate = self.config.get('reputation_decay', 0.01)
        
        for agent in self.agents:
            if agent.is_active:
                agent.reputation = max(0.0, agent.reputation * (1 - decay_rate))
    
    def get_simulation_result(self) -> SimulationResult:
        """Calculate final simulation results."""
        active_agents = [a for a in self.agents if a.is_active]
        honest_agents = [a for a in active_agents if a.agent_type == AgentType.HONEST]
        
        # Calculate honest share of revenue
        total_honest_revenue = sum(a.total_revenue for a in honest_agents)
        honest_share = total_honest_revenue / max(1, self.total_revenue)
        
        # Calculate expected loss for honest agents
        if honest_agents:
            expected_loss = np.mean([a.total_losses for a in honest_agents])
        else:
            expected_loss = 0.0
        
        # Calculate stake-at-risk curves
        stake_levels = [0.5, 1.0, 2.0, 5.0, 10.0]
        stake_at_risk = {}
        for level in stake_levels:
            agents_at_level = [a for a in active_agents if abs(a.stake - level) < 0.1]
            if agents_at_level:
                avg_risk = np.mean([a.get_stake_at_risk() for a in agents_at_level])
                stake_at_risk[level] = avg_risk
        
        # Calculate liveness score
        liveness_score = len(active_agents) / len(self.agents)
        
        # Calculate attack effectiveness metrics
        colluder_agents = [a for a in active_agents if a.agent_type == AgentType.COLLUDER]
        griefer_agents = [a for a in active_agents if a.agent_type == AgentType.GRIEFER]
        hoarder_agents = [a for a in active_agents if a.agent_type == AgentType.HOARDER]
        
        collusion_detection_rate = self.collusion_detections / max(1, len(colluder_agents))
        griefing_effectiveness = np.mean([a.griefing_intensity for a in griefer_agents]) if griefer_agents else 0.0
        hoarding_influence = np.mean([a.stake / self.total_stake for a in hoarder_agents]) if hoarder_agents else 0.0
        
        return SimulationResult(
            honest_share_of_revenue=honest_share,
            expected_loss_for_honest=expected_loss,
            stake_at_risk_curves=stake_at_risk,
            liveness_score=liveness_score,
            collusion_detection_rate=collusion_detection_rate,
            griefing_effectiveness=griefing_effectiveness,
            hoarding_influence=hoarding_influence,
            total_transactions=self.epoch * self.config['simulation']['transactions_per_epoch'],
            successful_attacks=self.collusion_detections,
            failed_attacks=len(colluder_agents) - self.collusion_detections
        )
