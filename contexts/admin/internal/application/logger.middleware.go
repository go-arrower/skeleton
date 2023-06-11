package application

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/exp/slog"
)

type DecoratorFunc[in, out any] interface {
	func(context.Context, in) (out, error)
}

type DecoratorFuncUnary[in any] interface {
	func(context.Context, in) error
}

// Logged wraps an application function / command with debug logs.
func Logged[in, out any, F DecoratorFunc[in, out]](logger *slog.Logger, next F) F { //nolint:ireturn
	return func(ctx context.Context, in in) (out, error) {
		cmdName := commandName(in)

		logger.DebugCtx(ctx, "executing command",
			slog.String("command", cmdName),
		)

		result, err := next(ctx, in)

		if err == nil {
			logger.DebugCtx(ctx, "command executed successfully",
				slog.String("command", cmdName))
		} else {
			logger.DebugCtx(ctx, "failed to execute command",
				slog.String("command", cmdName),
				slog.String("error", err.Error()),
			)
		}

		return result, err
	}
}

// LoggedU is like Logged but for functions only returning errors, e.g. jobs.
func LoggedU[in any, F DecoratorFuncUnary[in]](logger *slog.Logger, next F) F { //nolint:ireturn
	return func(ctx context.Context, in in) error {
		cmdName := commandName(in)

		logger.DebugCtx(ctx, "executing command",
			slog.String("command", cmdName),
		)

		err := next(ctx, in)

		if err == nil {
			logger.DebugCtx(ctx, "command executed successfully",
				slog.String("command", cmdName))
		} else {
			logger.DebugCtx(ctx, "failed to execute command",
				slog.String("command", cmdName),
				slog.String("error", err.Error()),
			)
		}

		return err
	}
}

// commandName extracts a printable name from cmd in the format of: functionName.
//
// functionName 								=> strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
// functionname 								=> strings.ToLower(strings.Split(fmt.Sprintf("%T", cmd), ".")[1])
// packageName.functionName 					=> fmt.Sprintf("%T", cmd)
// github.com/go-arrower/skeleton/.../package	=> fmt.Sprintln(reflect.TypeOf(cmd).PkgPath())
// functionName is used, the other examples are for inspiration.
func commandName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
}
