package infrastructure_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/go-arrower/arrower/alog"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/shared/infrastructure"
)

func TestSecret_String(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		secret string
	}{
		"empty":      {""},
		"whitespace": {" "},
		"secret":     {"this-should-be-masked"},
	}

	for name, tc := range tests {
		tc := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			s := infrastructure.Secret(tc.secret)
			fmt.Fprintln(buf, s)

			// t.Log(s)
			// t.Log(buf.String())
			assert.Equal(t, "******\n", buf.String())

			buf.Reset()
			logger := alog.NewTest(buf)
			logger.Info("msg", slog.Any("secret", s))

			// t.Log(buf.String())
			assert.Contains(t, buf.String(), "******")
			if notEmpty := strings.Trim(tc.secret, " ") != ""; notEmpty {
				assert.NotContains(t, buf.String(), tc.secret)
			}
		})
	}
}
