package telemetry

import (
	"context"

	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"go.opentelemetry.io/otel/trace"
)

func TracerFromContext(ctx context.Context) trace.Tracer {
	tracer, ok := ctx.Value(options.TracerKey).(trace.Tracer)
	if ok {
		return tracer
	}
	// TODO: Dangerous, this can cause a panic
	return nil
}
