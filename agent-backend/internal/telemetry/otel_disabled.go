package telemetry

import (
	"context"
	"log/slog"
)

// Init is a no-op telemetry initializer used when the 'otel' build tag is not set.
func Init(ctx context.Context, logger *slog.Logger) (func(context.Context) error, error) {
	return func(context.Context) error { return nil }, nil
}
