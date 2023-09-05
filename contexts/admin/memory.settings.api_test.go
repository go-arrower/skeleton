package admin_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/admin"
)

func TestSettingKey_Context(t *testing.T) {
	t.Parallel()

	tests := []struct {
		context  string
		key      string
		expected string
	}{
		{
			"",
			"",
			"",
		},
		{
			"context",
			"",
			"context",
		},
		{
			"context",
			"some_key",
			"context",
		},
		{
			"context.subcontext",
			"some.key",
			"context",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%s->%s", tt.context, tt.key), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, admin.NewSettingsKey(tt.context, tt.key).Context())
		})
	}
}

func TestNewSettingsValue(t *testing.T) {
	t.Parallel()

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		v := admin.NewSettingsValue("some")
		assert.Equal(t, admin.SettingValue("some"), v)
		assert.Equal(t, "some", v.String())
	})

	t.Run("bool", func(t *testing.T) {
		t.Parallel()

		v := admin.NewSettingsValue(true)
		assert.Equal(t, true, v.Bool())
	})

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		val := admin.NewSettingsValue(1337)
		assert.Equal(t, 1337, val.Int())

		val = admin.NewSettingsValue(1337)
		assert.Equal(t, int64(1337), val.Int64())

		val = admin.NewSettingsValue(int32(1337))
		assert.Equal(t, 1337, val.Int())
	})
}
