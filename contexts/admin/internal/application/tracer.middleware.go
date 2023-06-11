package application

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Traced wraps an application function / command with trace information.
func Traced[in, out any, F DecoratorFunc[in, out]](traceProvider trace.TracerProvider, next F) F { //nolint:ireturn
	tracer := traceProvider.Tracer("arrower.application") // trace.WithInstrumentationVersion("0.0.0"),

	return func(ctx context.Context, in in) (out, error) {
		cmdName := commandName(in)

		newCtx, span := tracer.Start(ctx, "usecase",
			trace.WithAttributes(attribute.String("command", cmdName)),
		)
		defer span.End()

		r, err := next(newCtx, in)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}

		return r, err
	}
}
