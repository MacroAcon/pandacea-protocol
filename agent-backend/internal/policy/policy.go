package policy

import (
	"context"
	"log/slog"

	"github.com/shopspring/decimal"
)

// Request represents a lease request to be evaluated
type Request struct {
	ProductID string `json:"productId"`
	MaxPrice  string `json:"maxPrice"`
	Duration  string `json:"duration"`
}

// EvaluationResult represents the result of a policy evaluation
type EvaluationResult struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// Engine represents the policy evaluation engine
type Engine struct {
	logger   *slog.Logger
	minPrice decimal.Decimal
}

// NewEngine creates a new policy engine
func NewEngine(logger *slog.Logger, minPriceStr string) (*Engine, error) {
	minPrice, err := decimal.NewFromString(minPriceStr)
	if err != nil {
		return nil, err
	}

	return &Engine{
		logger:   logger,
		minPrice: minPrice,
	}, nil
}

// EvaluateRequest evaluates a lease request according to the Guiding Principles
// Implements Dynamic Minimum Pricing (DMP) validation
func (e *Engine) EvaluateRequest(ctx context.Context, req *Request) *EvaluationResult {
	e.logger.Info("policy evaluation started",
		"product_id", req.ProductID,
		"max_price", req.MaxPrice,
		"duration", req.Duration,
		"min_price", e.minPrice.String(),
	)

	// Parse the request's max price
	requestPrice, err := decimal.NewFromString(req.MaxPrice)
	if err != nil {
		result := &EvaluationResult{
			Allowed: false,
			Reason:  "Invalid price format",
		}
		e.logger.Info("policy evaluation completed",
			"allowed", result.Allowed,
			"reason", result.Reason,
		)
		return result
	}

	// Check if the price meets the minimum requirement (DMP validation)
	if requestPrice.LessThan(e.minPrice) {
		result := &EvaluationResult{
			Allowed: false,
			Reason:  "Proposed maxPrice is below the dynamic minimum price.",
		}
		e.logger.Info("policy evaluation completed",
			"allowed", result.Allowed,
			"reason", result.Reason,
		)
		return result
	}

	result := &EvaluationResult{
		Allowed: true,
		Reason:  "Policy evaluation passed - price meets minimum requirement",
	}

	e.logger.Info("policy evaluation completed",
		"allowed", result.Allowed,
		"reason", result.Reason,
	)

	return result
}
