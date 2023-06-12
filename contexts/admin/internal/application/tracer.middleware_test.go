package application_test

import (
	"context"
	"testing"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
)

func TestTraced(t *testing.T) {
	t.Parallel()

	t.Run("successful command", func(t *testing.T) {
		t.Parallel()

		cmd := application.Traced(newFakeTracer(t), func(context.Context, exampleCommand) (string, error) {
			return "", nil
		})

		_, _ = cmd(context.Background(), exampleCommand{})
	})

	t.Run("failed command", func(t *testing.T) {
		t.Parallel()

		cmd := application.Traced(newFakeTracer(t), func(context.Context, exampleCommand) (string, error) {
			return "", errUseCaseFails
		})

		_, _ = cmd(context.Background(), exampleCommand{})
	})
}
