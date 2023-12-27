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
		{
			"",
			"some_key",
			"default",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%s->%s", tt.context, tt.key), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, admin.NewSettingKey(tt.context, tt.key).Context())
		})
	}
}

func TestNewSettingsValue(t *testing.T) {
	t.Parallel()

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		v := admin.NewSettingValue("some")
		assert.Equal(t, admin.SettingValue("some"), v)
		assert.Equal(t, "some", v.String())
	})

	t.Run("bool", func(t *testing.T) {
		t.Parallel()

		v := admin.NewSettingValue(true)
		assert.Equal(t, true, v.Bool())
		assert.Equal(t, "true", v.String())
	})

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		val := admin.NewSettingValue(1337)
		assert.Equal(t, 1337, val.Int())

		val = admin.NewSettingValue(1337)
		assert.Equal(t, int64(1337), val.Int64())

		val = admin.NewSettingValue(int32(1337))
		assert.Equal(t, 1337, val.Int())
		assert.Equal(t, "1337", val.String())
	})
}