package web_test

import (
	"errors"
	"io"

	"github.com/labstack/echo/v4"
)

var errUCFailed = errors.New("use case error")

type emptyRenderer struct{}

func (t *emptyRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return nil
}

// newTestRouter is a helper for unit tests, by returning a valid web router.
func newTestRouter() *echo.Echo {
	e := echo.New()
	e.Renderer = &emptyRenderer{}

	return e
}
