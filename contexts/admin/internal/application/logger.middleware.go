package application

import (
	"context"
	"fmt"

	"golang.org/x/exp/slog"
)

type DDecoratorFunc[in, out any] interface {
	func(context.Context, in) (out, error)
}

type DDecoratorFuncUnary[in, out any] interface {
	func(context.Context, in) error
}

func Logged[in, out any, F DDecoratorFunc[in, out]](logger *slog.Logger, next F) F {
	return func(ctx context.Context, in in) (out, error) {
		cmdName := commandName(in)

		logger.LogAttrs(ctx, slog.LevelDebug, "executing command",
			slog.String("command", cmdName),
		)

		r, err := next(ctx, in)

		if err == nil {
			logger.LogAttrs(ctx, slog.LevelDebug, "command executed successfully",
				slog.String("command", cmdName))
		} else {
			logger.LogAttrs(ctx, slog.LevelDebug, "failed to execute command",
				slog.String("command", cmdName),
				slog.String("error", err.Error()),
			)
		}

		return r, err
	}
}

func LoggedCommandUnary[in, out any, F DDecoratorFuncUnary[in, out]](logger *slog.Logger, next F) F {
	return func(ctx context.Context, in in) error {
		return next(ctx, in)
	}
}

func commandName(cmd any) string {
	// fmt.Println(reflect.TypeOf(cmd).PkgPath())
	// return strings.ToLower(strings.Split(fmt.Sprintf("%T", cmd), ".")[1])
	return fmt.Sprintf("%T", cmd)
}
