package application_test

import (
	"context"
	"net"
	"time"

	"github.com/go-arrower/skeleton/contexts/auth"

	"github.com/go-arrower/arrower/setting"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
)

const (
	validUserLogin       = "0@test.com"
	notVerifiedUserLogin = "1@test.com"
	blockedUserLogin     = "2@test.com"
	newUserLogin         = "99@test.com"

	strongPassword     = "R^&npAL2iu&M6S"                                               //nolint:gosec // gosec is right, but it's testdata
	strongPasswordHash = "$2a$10$T7Bq1sNmHoGlGJUsHoF1A.S3oy.P3iLT6MoVXi6WvNdq1jbE.TnZy" // hash of strongPassword

	sessionKey = "session-key"
	userAgent  = "arrower/1"
	ip         = "127.0.0.1"
)

const (
	user0Login            = "0@test.com"
	userIDZero            = user.ID("00000000-0000-0000-0000-000000000000")
	userNotVerifiedUserID = user.ID("00000000-0000-0000-0000-000000000001")
	userBlockedUserID     = user.ID("00000000-0000-0000-0000-000000000002")
)

var (
	ctx = context.Background()

	userVerified = user.User{
		ID:           userIDZero,
		Login:        user0Login,
		PasswordHash: user.PasswordHash(strongPasswordHash),
		Verified:     user.BoolFlag{}.SetTrue(),
		Sessions: []user.Session{{
			ID:        sessionKey,
			CreatedAt: time.Now().UTC(),
			ExpiresAt: time.Now().UTC().Add(time.Hour),
			Device:    user.Device{},
		}},
	}
	userNotVerified = user.User{
		ID:           userNotVerifiedUserID,
		Login:        user0Login,
		PasswordHash: user.PasswordHash(strongPasswordHash),
		Verified:     user.BoolFlag{}.SetFalse(),
	}
	userBlocked = user.User{
		ID:           userBlockedUserID,
		Login:        user0Login,
		PasswordHash: user.PasswordHash(strongPasswordHash),
		Blocked:      user.BoolFlag{}.SetTrue(),
	}

	resolvedIP = user.ResolvedIP{
		IP:          net.ParseIP(ip),
		Country:     "-",
		CountryCode: "-",
		Region:      "-",
		City:        "-",
	}
)

func registrator(repo user.Repository) *user.RegistrationService {
	settings := setting.NewInMemorySettings()
	settings.Save(ctx, auth.SettingAllowRegistration, setting.NewValue(true))

	return user.NewRegistrationService(settings, repo)
}

func authentificator() *user.AuthenticationService {
	settings := setting.NewInMemorySettings()
	settings.Save(ctx, auth.SettingAllowLogin, setting.NewValue(true))

	return user.NewAuthenticationService(settings)
}
