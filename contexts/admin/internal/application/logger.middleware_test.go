package application_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/go-arrower/arrower"
	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/stretchr/testify/assert"
)

type exampleCommand struct{}

func TestLoggedCommand(t *testing.T) {
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
		assert.Contains(t, buf.String(), `command=application_test.exampleCommand`)
		assert.Contains(t, buf.String(), `msg="command executed successfully"`)
	})

	t.Run("failed command", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		h := arrower.NewFilteredLogger(buf)

		cmd := application.Logged(h.Logger, func(context.Context, exampleCommand) (string, error) {
			return "", errors.New("some-error")
		})

		_, _ = cmd(context.Background(), exampleCommand{})

		assert.Contains(t, buf.String(), `msg="executing command"`)
		assert.Contains(t, buf.String(), `command=application_test.exampleCommand`)
		assert.Contains(t, buf.String(), `msg="failed to execute command"`)
		assert.Contains(t, buf.String(), `error=some-error`)
	})
}
