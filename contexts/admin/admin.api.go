package admin

import (
	"errors"
	"github.com/go-arrower/arrower/setting"
)

var ErrInvalidSetting = errors.New("invalid setting")

const contextName = "auth"

var (
	// todo rename settings to be more clear: e.g. SettingAllowRegistration
	SettingRegistration = setting.NewKey(contextName, "", "registration.registration_enabled")
	SettingLogin        = setting.NewKey(contextName, "", "registration.login_enabled")
)
