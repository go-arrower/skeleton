package testdata

import "github.com/go-arrower/skeleton/contexts/auth/internal/application/user"

const (
	ValidUserLogin       = "0@test.com"
	NotVerifiedUserLogin = "1@test.com"
	BlockedUserLogin     = "2@test.com"
	NewUserLogin         = "99@test.com"

	StrongPassword     = "R^&npAL2iu&M6S"                                               //nolint:gosec,lll // gosec is right, but it's testdata
	StrongPasswordHash = "$2a$10$T7Bq1sNmHoGlGJUsHoF1A.S3oy.P3iLT6MoVXi6WvNdq1jbE.TnZy" // hash of StrongPassword

	SessionKey = "session-key"
	UserAgent  = "arrower/1"
	IP         = "127.0.0.1"
)

const (
	UserIDZero            = user.ID("00000000-0000-0000-0000-000000000000")
	UserNotVerifiedUserID = user.ID("00000000-0000-0000-0000-000000000001")
	UserBlockedUserID     = user.ID("00000000-0000-0000-0000-000000000002")
)
