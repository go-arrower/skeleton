package web

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/go-arrower/skeleton/contexts/admin"
	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain"
)

func NewSettingsController(routes *echo.Group, repo domain.SettingRepository) SettingsController {
	repo.Create(context.Background(), admin.Setting{
		Key:   admin.NewSettingKey("", "someKey"),
		Value: admin.NewSettingValue(false),
		UIOptions: admin.Options{
			Type:         admin.Checkbox,
			Label:        "Some Arrower Key",
			Info:         "This is a extra info for the user",
			DefaultValue: admin.NewSettingValue(false),
			ReadOnly:     false,
			Danger:       false,
		},
	})
	repo.Create(context.Background(), admin.Setting{
		Key:   admin.NewSettingKey("", "custom.otherKey"),
		Value: admin.NewSettingValue(false),
		UIOptions: admin.Options{
			Type:         admin.Checkbox,
			Label:        "Some Other Arrower Key",
			Info:         "This is a extra info for the user",
			DefaultValue: admin.NewSettingValue(true),
			ReadOnly:     false,
			Danger:       false,
		},
	})

	return SettingsController{
		r:    routes,
		repo: repo,
		app:  application.NewSettingsApp(repo),
	}
}

type SettingsController struct {
	r *echo.Group

	repo domain.SettingRepository
	app  *application.SettingsApp
}

func (sc SettingsController) List() {
	sc.r.GET("/settings", func(c echo.Context) error {
		settings, _ := sc.repo.All(c.Request().Context())

		return c.Render(http.StatusOK, "=>admin.settings", echo.Map{
			"Settings": settingsToUISettings(settings),
		})
	}).Name = "admin.settings"
}

func (sc SettingsController) Update() {
	sc.r.POST("/settings", func(c echo.Context) error {
		key := c.FormValue("key")
		val := c.FormValue(key)

		if val == "on" { // the HTML default for checked checkboxes
			val = "true"
		}

		s, err := sc.app.UpdateAndGet(c.Request().Context(), admin.SettingKey(key), admin.NewSettingValue(val))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.Render(http.StatusOK, "setting.component", UISetting{
			Key:   string(s.Key),
			DOMID: strings.ReplaceAll(string(s.Key), ".", ""),
			Value: s.Value.String(),
			Type:  int(s.UIOptions.Type),
			Label: s.UIOptions.Label,
			Info:  s.UIOptions.Info,
		})
	})
}

type (
	UISettings map[string][]Group

	Group struct {
		Name     string
		Settings []UISetting
	}

	UISetting struct {
		Key   string
		DOMID string
		Value string
		Type  int
		Label string
		Info  string
	}
)

func settingsToUISettings(all []admin.Setting) UISettings {
	settings := map[admin.SettingKey]UISetting{}
	groups := map[string][]UISetting{}

	for _, s := range all {
		settings[s.Key] = UISetting{
			Key:   string(s.Key),
			DOMID: strings.ReplaceAll(string(s.Key), ".", ""), // htmx does not work with colon although allowed, see comment about jQuery https://stackoverflow.com/a/1077111
			Value: s.Value.String(),
			Type:  int(s.UIOptions.Type),
			Label: s.UIOptions.Label,
			Info:  s.UIOptions.Info,
		}

		kv := s.Key.Context() + groupFromSettingsKey(s.Key)
		groups[kv] = append(groups[kv], settings[s.Key])
	}

	ret := map[string][]Group{}
	for _, g := range groups {
		key := admin.SettingKey(g[0].Key)
		context := cases.Title(language.Und, cases.Compact).String(key.Context())
		ret[context] = append(ret[context], Group{
			Name:     groupFromSettingsKey(key),
			Settings: groups[key.Context()+groupFromSettingsKey(key)],
		})
	}

	return ret
}

func groupFromSettingsKey(key admin.SettingKey) string {
	s := strings.Split(string(key), ".")

	if len(s) <= 2 {
		return "Default"
	}

	return cases.Title(language.Und, cases.Compact).String(s[1])
}
