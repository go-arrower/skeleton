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

func TestBlockedFlag_IsBlocked(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testName string
		blocked  user.BlockedFlag
		expected bool
	}{
		{
			"empty time",
			user.BlockedFlag(time.Time{}),
			false,
		},
		{
			"blocked",
			user.BlockedFlag(time.Now()),
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.blocked.IsBlocked())
		})
	}
}

func TestSuperUserFlag_IsSuperuser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testName  string
		superuser user.SuperUserFlag
		expected  bool
	}{
		{
			"empty time",
			user.SuperUserFlag(time.Time{}),
			false,
		},
		{
			"superuser",
			user.SuperUserFlag(time.Now()),
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.superuser.IsSuperuser())
		})
	}
}

func TestDevice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testName     string
		device       user.Device
		expectedName string
		expectedOS   string
	}{
		{
			"",
			user.NewDevice("Mozilla/5.0 (Linux; Android 4.3; GT-I9300 Build/JSS15J) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.125 Mobile Safari/537.36"),
			"Chrome v59.0.3071.125",
			"Android v4.3",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expectedName, tt.device.Name())
			assert.Equal(t, tt.expectedOS, tt.device.OS())
		})
	}
}
