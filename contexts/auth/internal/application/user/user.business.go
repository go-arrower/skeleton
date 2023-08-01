package user

import (
	"time"

	"golang.org/x/crypto/bcrypt"

	"golang.org/x/text/language"

	"github.com/google/uuid"
)

// User represents a user of the software, that can perform all the auth functionalities.
type User struct {
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
	// a quick helper for simple stuff, if you have a complicated profile => do it in your Context, as it's the better place
	Profile  Profile  // limit the length of keys & values // { plan: 'silver', team_id: 'a111' }
	Profile2 Profile2 // limit the length of keys & values // { plan: 'silver', team_id: 'a111' }
	// email, phone???

	Verified  VerifiedFlag
	Blocked   BlockedFlag
	SuperUser SuperUserFlag
}

// ID is the primary identifier of a User.
type ID string

func NewID() ID {
	return ID(uuid.NewString())
}

type Login string

type PasswordHash string

func (pw PasswordHash) Matches(checkPW string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(pw), []byte(checkPW)); err == nil {
		return true
	}

	return false
}

// String prevents a hash to exponentially leak by masking it in functions like fmt.
func (pw PasswordHash) String() string { return "xxxxxx" }

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

func NewBirthday(day Day, month Month, year Year) (Birthday, error) { return Birthday{}, nil } // todo test the API and types, if they cast automatically and the constructor is convenient to use

func (b Birthday) String() string { return "" }

// func (b Birthday) Format(layout string) string { return "" }
// func (b Birthday) Format(loc Locale) string { return "" }

type Locale language.Tag

type TimeZone string

type URL string

func NewURL(url string) (URL, error) { return "", nil }

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

func (t BlockedFlag) IsBlocked() bool { return false }
func (t BlockedFlag) At() time.Time   { return time.Time(t) }

type SuperUserFlag time.Time

func (t SuperUserFlag) IsSuperuser() bool { return false }
func (t SuperUserFlag) At() time.Time     { return time.Time(t) }

type (
	VerificationToken string
	UserRegistered    struct {
		ID         ID
		RecordedAt time.Time
	}
)

func NewUser(...any) User { return User{} }