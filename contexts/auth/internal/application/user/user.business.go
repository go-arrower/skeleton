package user

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/mileusna/useragent"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/language"
)

var (
	ErrPasswordTooWeak = errors.New("password too weak")
	ErrInvalidBirthday = errors.New("invalid birthday")
)

// User represents a user of the software, that can perform all the auth functionalities.
type User struct { //nolint:govet // fieldalignment less important than grouping of fields.
	ID           ID
	Login        Login // UserName / email, or phone, or nickname, or whatever the developer wants to have as a login
	PasswordHash PasswordHash
	RegisteredAt time.Time

	FirstName         string
	LastName          string
	Name              string // DisplayName
	Birthday          Birthday
	Locale            Locale
	TimeZone          TimeZone
	ProfilePictureURL URL
	// a helper for simple stuff, if you have a complicated profile => do it in your Context, as it's the better place
	Profile  Profile  // limit the length of keys & values // { plan: 'silver', team_id: 'a111' }
	Profile2 Profile2 // limit the length of keys & values // { plan: 'silver', team_id: 'a111' }
	// email, phone???

	Verified  VerifiedFlag
	Blocked   BlockedFlag
	SuperUser SuperUserFlag

	Sessions []Session
}

func NewID() ID {
	return ID(uuid.NewString())
}

// ID is the primary identifier of a User.
type ID string

type Login string

// NewPasswordHash returns a PasswordHash. Consider NewStrongPasswordHash instead.
func NewPasswordHash(password string) (PasswordHash, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return PasswordHash(hash), err
}

// NewStrongPasswordHash returns a PasswordHash or an error, if the password is too weak.
func NewStrongPasswordHash(password string) (PasswordHash, error) {
	if isWeakPassword(password) {
		return "", ErrPasswordTooWeak
	}

	return NewPasswordHash(password)
}

var (
	upperCase   = regexp.MustCompile("[A-Z]")
	lowerCase   = regexp.MustCompile("[a-z]")
	number      = regexp.MustCompile("[0-9]")
	specialChar = regexp.MustCompile("[!@#$%^&*]")
)

// isWeakPassword required the password to be:
// - 8 characters or longer
// - contain at least one lower case letter
// - contain at least one upper case letter
// - contain at least one number
// - contain at least one special character.
func isWeakPassword(password string) bool {
	minPasswordLength := 8
	if len(password) < minPasswordLength {
		return true
	}

	matchRules := []*regexp.Regexp{upperCase, lowerCase, number, specialChar}
	mPW := []byte(password)

	for _, r := range matchRules {
		if !r.Match(mPW) {
			return true
		}
	}

	return false
}

type PasswordHash string

func (pw PasswordHash) Matches(checkPW string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(pw), []byte(checkPW)); err == nil {
		return true
	}

	return false
}

// String prevents a hash to exponentially leak by masking it in functions like fmt.
func (pw PasswordHash) String() string { return "xxxxxx" }

func NewBirthday(day Day, month Month, year Year) (Birthday, error) {
	if day < 1 || day > 31 {
		return Birthday{}, ErrInvalidBirthday
	}

	if month < 1 || month > 12 {
		return Birthday{}, ErrInvalidBirthday
	}

	const maxAge = 150 * 356 * 24 * time.Hour // 150 years
	isTooOld := int(year) < time.Now().UTC().Add(-maxAge).Year()
	isInTheFuture := int(year) > time.Now().UTC().Year()

	if isTooOld || isInTheFuture {
		return Birthday{}, ErrInvalidBirthday
	}

	_, err := time.Parse(time.DateOnly, fmt.Sprintf("%d-%02d-%02d", year, month, day))
	if err != nil {
		return Birthday{}, ErrInvalidBirthday
	}

	return Birthday{day: day, month: month, year: year}, nil
}

type (
	Day   uint8
	Month uint8
	Year  uint16

	Birthday struct {
		day   Day
		month Month
		year  Year
	}
)

func (b Birthday) String() string { return "" }

// func (b Birthday) Format(layout string) string { return "" }
// func (b Birthday) Format(loc Locale) string { return "" }

type Locale language.Tag

type TimeZone string

func NewURL(url string) (URL, error) { return "", nil }

type URL string

type (
	ProfileKey   string
	ProfileValue string

	Profile  map[ProfileKey]ProfileValue
	Profile2 map[string]*string
)

type VerifiedFlag time.Time

func (t VerifiedFlag) IsVerified() bool {
	return time.Time(t) != time.Time{}
}

func (t VerifiedFlag) At() time.Time { return time.Time(t) }

type BlockedFlag time.Time

func (t BlockedFlag) IsBlocked() bool {
	return time.Time(t) != time.Time{}
}

func (t BlockedFlag) At() time.Time { return time.Time(t) }

type SuperUserFlag time.Time

func (t SuperUserFlag) IsSuperuser() bool {
	return time.Time(t) != time.Time{}
}

func (t SuperUserFlag) At() time.Time { return time.Time(t) }

type (
	VerificationToken string

	Registered struct { //nolint:govet // fieldalignment less important than grouping of fields.
		ID         ID
		RecordedAt time.Time
	}
)

func NewDevice(userAgent string) Device {
	return Device{userAgent: userAgent}
}

// Device contains human friendly information about the device the user is using.
type Device struct {
	userAgent string
}

func (d Device) Name() string {
	ua := useragent.Parse(d.userAgent)

	return fmt.Sprintf("%s v%s", ua.Name, ua.Version)
}

func (d Device) OS() string {
	ua := useragent.Parse(d.userAgent)

	return fmt.Sprintf("%s v%s", ua.OS, ua.OSVersion)
}

type Session struct {
	ID        string
	CreatedAt time.Time
	ExpiresAt time.Time
	Device    Device
}
