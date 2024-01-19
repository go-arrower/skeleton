// the template package uses white-box tests, so this is not a _test package.
package testdata

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

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
	LContent        = "layout" // todo why is there also LDefaultContent?
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

var SingleNonDefaultLayout = fstest.MapFS{
	"pages/p0.page.html": {Data: []byte(P0Content)},
	"global.layout.html": {Data: []byte(LContent)},
}

var MultipleLayoutsWithDefaultLayout = fstest.MapFS{
	"pages/p0.html":       {Data: []byte(P0Content)},
	"global.layout.html":  {Data: []byte(LContent)},
	"default.layout.html": {Data: []byte(LDefaultContent + ` {{template "content" .}}`)},
	"other.layout.html":   {Data: []byte(LOtherContent + ` {{template "content" .}}`)},
}

var ConflictingTemplateFiles = fstest.MapFS{
	"components/conflict.html": {Data: []byte(C0Content)},
	"pages/conflict.html":      {Data: []byte(P0Content)},
}

func NewEchoContext(t *testing.T) echo.Context {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	return echo.New().NewContext(req, rec)
}

func GenRandomPages(numPages int) (fstest.MapFS, []string) {
	fs := fstest.MapFS{
		"default.layout.html": {Data: []byte(LContent + ` {{template "content" .}}`)},
	}

	var pageNames []string

	for i := 0; i < numPages; i++ {
		p := randomString(5)
		fs["pages/"+p+".html"] = &fstest.MapFile{Data: []byte(p)} //nolint:exhaustruct

		pageNames = append(pageNames, p)
	}

	return fs, pageNames
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString(n int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec // used for ids, not security

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rnd.Intn(len(letters))]
	}

	return string(b)
}
