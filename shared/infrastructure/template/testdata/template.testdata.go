// the template package uses white-box tests, so this is not a _test package.
package testdata

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v4"
)

const (
	C0Content       = "c0"
	C1Content       = "c1"
	P0Content       = "p0"
	P1Content       = "p1"
	P2Content       = "p2"
	F0Content       = "f0"
	F1Content       = "f1"
	LContent        = "layout"
	LDefaultContent = "defaultLayout"
	LOtherContent   = "otherLayout"
)

var EmptyFiles = fstest.MapFS{}

var TemplateFiles = fstest.MapFS{
	"components/c0.html": {Data: []byte(C0Content)},
	"components/c1.html": {Data: []byte(C1Content)},
	"pages/p0.html":      {Data: []byte(P0Content)},
	"pages/p1.html":      {Data: []byte(P1Content + ` {{template "c0" .}}`)},
	"pages/p2.html":      {Data: []byte(P2Content + fmt.Sprintf(`{{block "f0" .}}%s{{end}} {{block "f1" .}}%s{{end}}`, F0Content, F1Content))},
	"global.layout.html": {Data: []byte(LContent)},
}

var LayoutsPagesAndComponents = fstest.MapFS{
	"components/c0.html": {Data: []byte(C0Content)},
	"components/c1.html": {Data: []byte(C1Content)},
	"pages/p0.html":      {Data: []byte(P0Content + ` {{template "c0" .}}`)},
	"pages/p1.html":      {Data: []byte(P1Content)},
	"global.layout.html": {Data: []byte(`<!DOCTYPE html>
<html lang="en">
<body>
	layout
    {{ block "layout" . }}
        {{block "content" .}}
            Fallback, if "content" is not defined elsewhere
        {{end}}
    {{end}}
</body>
</html>`)},
	"other.layout.html": {Data: []byte(`otherLayout
    {{ block "layout" . }}
        {{block "content" .}}
            Fallback, if "content" is not defined elsewhere
        {{end}}
    {{end}}`)},
}

var LayoutOneLayout = fstest.MapFS{
	"pages/p0.page.html": {Data: []byte(P0Content)},
	"global.layout.html": {Data: []byte(LContent)},
}

var LayoutWithDefault = fstest.MapFS{
	"pages/p0.html":       {Data: []byte(P0Content)},
	"global.layout.html":  {Data: []byte(LContent)},
	"default.layout.html": {Data: []byte(LDefaultContent + ` {{template "content" .}}`)},
	"other.layout.html":   {Data: []byte(LOtherContent + ` {{template "content" .}}`)},
}

func NewEmptyEchoContext(t *testing.T) echo.Context {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	return echo.New().NewContext(req, rec)
}
