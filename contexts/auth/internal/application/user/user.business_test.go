package user_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
)

func TestUser_IsVerified(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testName string
		user     user.User
		expected bool
	}{
		{
			"empty time",
			user.User{Verified: user.BoolFlag{}},
			false,
		},
		{
			"user",
			user.User{Verified: user.BoolFlag(time.Now().UTC())},
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.user.IsVerified())
		})
	}
}

func TestUser_IsBlocked(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testName string
		user     user.User
		expected bool
	}{
		{
			"empty time",
			user.User{Blocked: user.BoolFlag{}},
			false,
		},
		{
			"user",
			user.User{Blocked: user.BoolFlag(time.Now().UTC())},
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.user.IsBlocked())
		})
	}
}

func TestUser_Block(t *testing.T) {
	t.Parallel()

	user := user.User{}
	assert.False(t, user.IsBlocked())

	user.Block()
	assert.True(t, user.IsBlocked())

	blockedAt := user.Blocked.At()
	user.Block()

	assert.Equal(t, blockedAt, user.Blocked.At(), "if user is blocked, new calls to block will not update the time")
}

func TestUser_Unblock(t *testing.T) {
	t.Parallel()

	user := user.User{}
	assert.False(t, user.IsBlocked())

	user.Unblock()
	assert.False(t, user.IsBlocked(), "no change on already unblocked user")

	user.Block()
	user.Unblock()
	assert.False(t, user.IsBlocked())
}

func TestUser_IsSuperuser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testName string
		user     user.User
		expected bool
	}{
		{
			"empty time",
			user.User{SuperUser: user.BoolFlag{}},
			false,
		},
		{
			"superuser",
			user.User{SuperUser: user.BoolFlag(time.Now().UTC())},
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.user.IsSuperuser())
		})
	}
}

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

func TestBoolFlag(t *testing.T) {
	t.Parallel()

	flag := user.BoolFlag{}
	assert.False(t, flag.IsTrue())
	assert.True(t, flag.IsFalse())
	assert.Empty(t, flag.At())

	flag = user.BoolFlag(time.Now().UTC())
	assert.True(t, flag.IsTrue())
	assert.False(t, flag.IsFalse())
	assert.NotEmpty(t, flag.At())
}

func TestBoolFlag_SetTrue(t *testing.T) {
	t.Parallel()

	t.Run("set true", func(t *testing.T) {
		t.Parallel()

		flag := user.BoolFlag{}
		assert.False(t, flag.IsTrue())

		flag = flag.SetTrue()
		assert.True(t, flag.IsTrue())
	})

	t.Run("if flag was true, time does not change", func(t *testing.T) {
		t.Parallel()

		flag := user.BoolFlag{}
		assert.False(t, flag.IsTrue())

		flag = flag.SetTrue()
		assert.True(t, flag.IsTrue())
		trueAt := flag.At()

		flag = flag.SetTrue()
		assert.True(t, flag.IsTrue())
		assert.Equal(t, trueAt, flag.At(), "second call does not change the time")
	})
}

func TestBoolFlag_SetFalse(t *testing.T) {
	t.Parallel()

	t.Run("set false", func(t *testing.T) {
		t.Parallel()

		flag := user.BoolFlag(time.Now().UTC())
		assert.True(t, flag.IsTrue())

		flag = flag.SetFalse()
		assert.True(t, flag.IsFalse())

		flag = flag.SetFalse()
		assert.True(t, flag.IsFalse(), "subsequent calls stay false")
	})
}
