package application_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-arrower/arrower"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
)

func TestLogged(t *testing.T) {
	t.Parallel()

	t.Run("successful command", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		h := arrower.NewFilteredLogger(buf)

		cmd := application.Logged(h.Logger, func(context.Context, exampleCommand) (string, error) {
			return "", nil
		})

		_, _ = cmd(context.Background(), exampleCommand{})

		assert.Contains(t, buf.String(), `msg="executing command"`)
		assert.Contains(t, buf.String(), `command=exampleCommand`)
		assert.Contains(t, buf.String(), `msg="command executed successfully"`)
	})

	t.Run("failed command", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		h := arrower.NewFilteredLogger(buf)

		cmd := application.Logged(h.Logger, func(context.Context, exampleCommand) (string, error) {
			return "", errUseCaseFails
		})

		_, _ = cmd(context.Background(), exampleCommand{})

		assert.Contains(t, buf.String(), `msg="executing command"`)
		assert.Contains(t, buf.String(), `command=exampleCommand`)
		assert.Contains(t, buf.String(), `msg="failed to execute command"`)
		assert.Contains(t, buf.String(), `error=some-error`)
	})
}

func TestLoggedU(t *testing.T) {
	t.Parallel()

	t.Run("successful command", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		h := arrower.NewFilteredLogger(buf)

		cmd := application.LoggedU(h.Logger, func(context.Context, exampleCommand) error {
			return nil
		})

		_ = cmd(context.Background(), exampleCommand{})

		assert.Contains(t, buf.String(), `msg="executing command"`)
		assert.Contains(t, buf.String(), `command=exampleCommand`)
		assert.Contains(t, buf.String(), `msg="command executed successfully"`)
	})

	t.Run("failed command", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		h := arrower.NewFilteredLogger(buf)

		cmd := application.LoggedU(h.Logger, func(context.Context, exampleCommand) error {
			return errUseCaseFails
		})

		_ = cmd(context.Background(), exampleCommand{})

		assert.Contains(t, buf.String(), `msg="executing command"`)
		assert.Contains(t, buf.String(), `command=exampleCommand`)
		assert.Contains(t, buf.String(), `msg="failed to execute command"`)
		assert.Contains(t, buf.String(), `error=some-error`)
	})
}
