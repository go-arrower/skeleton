package web

import (
	"context"
	"fmt"

	"github.com/go-arrower/skeleton/contexts/auth"
	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/contexts/admin"
)

const appTitle = "skeleton"

func NewDefaultPresenter(settingsAPI admin.SettingsAPI) *DefaultPresenter {
	return &DefaultPresenter{
		settingsAPI: settingsAPI,
	}
}

type DefaultPresenter struct {
	settingsAPI admin.SettingsAPI
}

type BasePage struct {
	Title   string      // HTML title
	Flashes interface{} // flash messages

	D echo.Map

	ShowRegistrationBtn bool
	ShowLoginBtn        bool
	ShowLogoutBtn       bool
}

type MapBasePage map[string]any

func (p *DefaultPresenter) MapDefaultBasePage(ctx context.Context, title string, keyVals ...echo.Map) (MapBasePage, error) {
	docTitle := fmt.Sprintf("%s - %s", title, appTitle)
	if title == "" {
		docTitle = appTitle
	}

	isRegisterActive, _ := p.settingsAPI.Setting(ctx, admin.SettingRegistration)
	isLoginActive, _ := p.settingsAPI.Setting(ctx, admin.SettingLogin)

	showLoginBtn := isLoginActive.Bool() && !auth.IsLoggedIn(ctx)

	basePage := MapBasePage{
		"Title":                    docTitle,
		"Flashes":                  nil,
		"ShowRegistrationBtn":      isRegisterActive.Bool() && !auth.IsLoggedIn(ctx),
		"ShowLoginBtn":             showLoginBtn,
		"ShowLogoutBtn":            auth.IsLoggedIn(ctx),
		"ShowAdminBtn":             auth.IsSuperUser(ctx),
		"ShowLoggedInAsUserBanner": auth.IsLoggedInAsOtherUser(ctx),
	}

	if len(keyVals) > 0 {
		for k, v := range keyVals[0] {
			basePage[k] = v
		}
	}

	return basePage, nil
}
func (p *DefaultPresenter) MustMapDefaultBasePage(ctx context.Context, title string, keyVals ...echo.Map) MapBasePage {
	r, _ := p.MapDefaultBasePage(ctx, title, keyVals...)
	return r
}

func (p *DefaultPresenter) DefaultBasePage(ctx context.Context, title string, keyVals ...echo.Map) BasePage {
	docTitle := fmt.Sprintf("%s - %s", title, appTitle)
	if title == "" {
		docTitle = appTitle
	}

	d := echo.Map{}
	if len(keyVals) > 0 {
		d = keyVals[0]
	}

	isRegisterActive, _ := p.settingsAPI.Setting(ctx, admin.SettingRegistration)

	return BasePage{
		Title:               docTitle,
		Flashes:             nil,
		D:                   d,
		ShowRegistrationBtn: isRegisterActive.Bool(),
		ShowLoginBtn:        admin.SettingValue("").Bool(),
		ShowLogoutBtn:       admin.SettingValue("").Bool(),
	}
}