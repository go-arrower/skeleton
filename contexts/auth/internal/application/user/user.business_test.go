package user_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
)

func TestVerifiedFlag_IsVerified(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testName string
		verified user.VerifiedFlag
		expected bool
	}{
		{
			"empty time",
			user.VerifiedFlag(time.Time{}),
			false,
		},
		{
			"verified",
			user.VerifiedFlag(time.Now()),
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.verified.IsVerified())
		})
	}
}
