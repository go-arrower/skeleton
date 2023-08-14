package user_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
)

func TestNewPasswordHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testName string
		pw       string
		err      error
	}{
		{
			"empty pw",
			"",
			nil,
		},
		{
			"pw",
			"some-pw",
			nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			_, err := user.NewPasswordHash(tt.pw)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestNewStrongPasswordHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testName string
		password string
	}{
		{
			"too short",
			"123456",
		},
		{
			"missing lower case letter",
			"1234567890",
		},
		{
			"missing upper case letter",
			"123456abc",
		},
		{
			"missing number",
			"abcdefghi",
		},
		{
			"missing special character",
			"123456abCD",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			_, err := user.NewStrongPasswordHash(tt.password)
			assert.Error(t, err)
			assert.ErrorIs(t, err, user.ErrPasswordTooWeak)
		})
	}
}

func TestName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testName string
		fn       string
		ln       string
		dn       string
		expFN    string
		expLN    string
		expDN    string
	}{
		{
			"empty name",
			"",
			"",
			"",
			"",
			"",
			"",
		},
		{
			"full name",
			"Arrower",
			"Project",
			"Arrower Project",
			"Arrower",
			"Project",
			"Arrower Project",
		},
		{
			"sanitise name",
			" Arrower",
			"Project ",
			" Arrower Project ",
			"Arrower",
			"Project",
			"Arrower Project",
		},
		{
			"automatic capitalise",
			"arrower",
			"project",
			"arrower project",
			"Arrower",
			"Project",
			"Arrower Project",
		},
		{
			"build display name",
			"arrower",
			"project",
			"",
			"Arrower",
			"Project",
			"Arrower Project",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			name := user.NewName(tt.fn, tt.ln, tt.dn)
			assert.Equal(t, tt.expFN, name.FirstName())
			assert.Equal(t, tt.expLN, name.LastName())
			assert.Equal(t, tt.expDN, name.DisplayName())
		})
	}
}

func TestNewBirthday(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testName string
		day      user.Day
		month    user.Month
		year     user.Year
		expected error
	}{
		{
			"",
			1,
			1,
			2000,
			nil,
		},
		{
			"invalid day",
			0,
			1,
			2000,
			user.ErrInvalidBirthday,
		},
		{
			"invalid month",
			1,
			0,
			2000,
			user.ErrInvalidBirthday,
		},
		{
			"too old",
			1,
			1,
			1000,
			user.ErrInvalidBirthday,
		},
		{
			"invalid day",
			32,
			1,
			2000,
			user.ErrInvalidBirthday,
		},
		{
			"invalid month",
			1,
			13,
			2000,
			user.ErrInvalidBirthday,
		},
		{
			"in the future",
			1,
			1,
			3000,
			user.ErrInvalidBirthday,
		},
		{
			"",
			29,
			2,
			2020,
			nil,
		},
		{
			"invalid date",
			31,
			4,
			2020,
			user.ErrInvalidBirthday,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			_, got := user.NewBirthday(tt.day, tt.month, tt.year)
			assert.ErrorIs(t, got, tt.expected)
		})
	}
}

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
