package admin

import (
	"errors"
	"fmt"
	"github.com/go-arrower/arrower/setting"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidSetting = errors.New("invalid setting")
)

const contextName = "auth"

var (
	// todo rename settings to be more clear: e.g. SettingAllowRegistration
	SettingRegistration = setting.NewKey(contextName, "", "registration.registration_enabled")
	SettingLogin        = setting.NewKey(contextName, "", "registration.login_enabled")
)

type (
	SettingKey string

	// SettingValue represents the actual value. Use NewSettingValue.
	// If you use one of the helper methods to cast the value to a different type:
	// - you have to ensure the type is correct
	// - you have to serialise it
	// OR just use the NewSettingValue, which ensures this for you.
	SettingValue string // todo OR struct with Type and helpers to ensure only right values are returned

	Setting struct {
		Key       SettingKey
		Value     SettingValue
		UIOptions Options
	}
)

// NewSettingKey construct a SettingKey.
func NewSettingKey(context string, key string) SettingKey {
	if context == "" && key == "" {
		return ""
	}

	if context == "" {
		context = "default"
	}

	return SettingKey(fmt.Sprintf("%s.%s", context, key))
}

func (key SettingKey) Context() string {
	s := strings.Split(string(key), ".")

	return s[0]
}

func (v SettingValue) String() string { return string(v) }

func (v SettingValue) Bool() bool {
	b, _ := strconv.ParseBool(string(v))

	return b
}

func (v SettingValue) Int() int { return int(v.Int64()) }

func (v SettingValue) Int64() int64 {
	i, _ := strconv.Atoi(string(v))

	return int64(i)
}

// TODO
// func (v SettingValue) Float64() float64       { return 0 }
// func (v SettingValue) Map() map[string]string { return nil }
// func (v SettingValue) Time() time.Time        { return time.Time{} }
// func (v SettingValue) JSON() json.RawMessage  { return nil }

// NewSettingValue returns a valid SettingValue for val.
// Use it in cases, when you can not convert to SettingValue yourself.
func NewSettingValue(val any) SettingValue { //nolint:gocyclo,cyclop // function is long but not complex
	r := reflect.TypeOf(val)

	switch r.Kind() { //nolint:exhaustive // not all cases are valid
	case reflect.String:
		return SettingValue(val.(string)) //nolint:forcetypeassert
	case reflect.Bool:
		return SettingValue(strconv.FormatBool(val.(bool))) //nolint:forcetypeassert
	case reflect.Int:
		return SettingValue(strconv.Itoa(val.(int))) //nolint:forcetypeassert
	case reflect.Int8:
		return SettingValue(strconv.Itoa(int(val.(int8)))) //nolint:forcetypeassert
	case reflect.Int16:
		return SettingValue(strconv.Itoa(int(val.(int16)))) //nolint:forcetypeassert
	case reflect.Int32:
		return SettingValue(strconv.Itoa(int(val.(int32)))) //nolint:forcetypeassert
	case reflect.Int64:
		return SettingValue(strconv.Itoa(int(val.(int64)))) //nolint:forcetypeassert
	case reflect.Uint:
		return SettingValue(strconv.Itoa(int(val.(uint)))) //nolint:forcetypeassert
	case reflect.Uint8:
		return SettingValue(strconv.Itoa(int(val.(uint8)))) //nolint:forcetypeassert
	case reflect.Uint16:
		return SettingValue(strconv.Itoa(int(val.(uint16)))) //nolint:forcetypeassert
	case reflect.Uint32:
		return SettingValue(strconv.Itoa(int(val.(uint32)))) //nolint:forcetypeassert
	case reflect.Uint64:
		return SettingValue(strconv.Itoa(int(val.(uint64)))) //nolint:forcetypeassert
	default:
		return ""
	}
}

type Type int

const (
	Checkbox Type = iota
)

func (t Type) IsValid() bool {
	switch t {
	case Checkbox:
		return true
	}

	return false
}

type Options struct {
	Type         Type
	Label        string
	Info         string
	Placeholder  string
	DefaultValue SettingValue
	ReadOnly     bool
	Danger       bool
	Validators   []SettingValidateFunc
}

type SettingValidateFunc func(s SettingValue) error
