package admin

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type SettingsAPI interface {
	Setting(ctx context.Context, setting SettingKey) (SettingValue, error)
	Settings(ctx context.Context, settings ...SettingKey) ([]SettingValue, error)
	SettingsByContext(ctx context.Context, context string) ([]SettingValue, error)

	Create(ctx context.Context, setting Setting) error
}

type (
	SettingKey string

	// SettingValue represents the actual value. Use NewSettingsValue.
	// If you use one of the helper methods to cast the value to a different type:
	// - you have to ensure the type is correct
	// - you have to serialise it
	// OR just use the NewSettingsValue, which ensures this for you.
	SettingValue string

	Setting struct {
		Key   SettingKey
		Value SettingValue
	}
)

// NewSettingsKey construct a SettingKey.
func NewSettingsKey(context string, key string) SettingKey {
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

// func (v SettingValue) Float64() float64       { return 0 }
// func (v SettingValue) Map() map[string]string { return nil }
// func (v SettingValue) Time() time.Time        { return time.Time{} }
// func (v SettingValue) JSON() json.RawMessage  { return nil }

// NewSettingsValue returns a valid SettingValue for val.
// Use it in cases, when you can not convert to SettingValue yourself.
func NewSettingsValue(val any) SettingValue { //nolint:gocyclo,cyclop // function is long but not complex
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