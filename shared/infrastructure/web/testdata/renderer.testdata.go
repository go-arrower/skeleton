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
	C0Content                    = "c0"
	C1Content                    = "c1"
	P0Content                    = "p0"
	P1Content                    = "p1"
	P2Content                    = "p2"
	F0Content                    = "f0"
	F1Content                    = "f1"
	BaseLayoutContent            = "baseLayout"
	BaseLayoutPagePlaceholder    = "pageLayout placeholder"
	BaseLayoutContentPlaceholder = "content placeholder"
	BaseDefaultLayoutContent     = "defaultBaseLayout"
)

var FilesEmpty = fstest.MapFS{}

func FilesSharedViews() fstest.MapFS {
	return fstest.MapFS{
		"components/c0.html":       {Data: []byte(C0Content)},
		"components/c1.html":       {Data: []byte(C1Content)},
		"pages/p0.html":            {Data: []byte(P0Content)},
		"pages/p1.html":            {Data: []byte(P1Content + ` {{template "c0" .}}`)},
		"pages/p2.html":            {Data: []byte(P2Content + fmt.Sprintf(`{{block "f0" .}}%s{{end}} {{block "f1" .}}%s{{end}}`, F0Content, F1Content))},
		"pages/shared.html":        {Data: []byte(P0Content + ` {{template "c0" .}}`)},
		"pages/conflict-page.html": {Data: []byte(P0Content)},
		"global.base.html": {Data: []byte(BaseLayoutContent + `
    {{block "layout" .}}
		` + BaseLayoutPagePlaceholder + `
        {{block "content" .}}
            ` + BaseLayoutContentPlaceholder + `
        {{end}}
    {{end}}`)},
	}
}

func FilesSharedViewsWithoutBase() fstest.MapFS {
	fs := FilesSharedViews()
	delete(fs, "global.base.html")

	return fs
}

// TODO "otherLayout" might be a mistake below, as it is in global layout file
func FilesSharedViewsWithMultiBase() fstest.MapFS {
	fs := FilesSharedViews()

	fs["global.base.html"] = &fstest.MapFile{Data: []byte(`<!DOCTYPE html>
<html lang="en">
<body>
	globalLayout
    {{block "layout" .}}
        {{block "content" .}}
            contentPlaceholder
        {{end}}
    {{end}}
</body>
</html>`),
	}
	fs["other.base.html"] = &fstest.MapFile{Data: []byte(`otherLayout
	{{block "layout" .}}
	   {{block "content" .}}
	       contentPlaceholder
	   {{end}}
	{{end}}`),
	}

	return fs
}

func FilesSharedViewsWithDefaultBase() fstest.MapFS {
	fs := FilesSharedViews()

	fs["default.base.html"] = &fstest.MapFile{Data: []byte(BaseDefaultLayoutContent +
		`{{block "layout" .}}` + BaseLayoutPagePlaceholder + `
			{{block "content" .}}
            	` + BaseLayoutContentPlaceholder + `
            {{end}}
		{{end}}`)}

	return fs
}

func FilesSharedViewsWithDefaultBaseWithoutLayout() fstest.MapFS {
	fs := FilesSharedViews()

	fs["default.base.html"] = &fstest.MapFile{Data: []byte(BaseDefaultLayoutContent + ` {{block "content" .}}`)}

	return fs
}

func FilesSharedViewsWithCustomFuncs() fstest.MapFS {
	fs := FilesSharedViews()

	fs["components/use-func-map.html"] = &fstest.MapFile{Data: []byte(`{{ customFunc }}`)}
	fs["pages/use-func-map.html"] = &fstest.MapFile{Data: []byte(`{{ hello }} {{ customFunc }}`)}

	return fs
}

var SingleNonDefaultLayout = fstest.MapFS{ // TODO remove?
	"pages/p0.page.html": {Data: []byte(P0Content)},
	"global.base.html":   {Data: []byte(BaseLayoutContent)},
}

func NewEchoContext(t *testing.T) echo.Context { // TODO rename to Test instead of new
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	return echo.New().NewContext(req, rec)
}

func NewExampleContextEchoContext(t *testing.T) echo.Context { // TODO rename to Test instead of new
	t.Helper()

	c := NewEchoContext(t)
	c.SetPath(fmt.Sprintf("/%s", ExampleContext))

	return c
}

func GenRandomPages(numPages int) (fstest.MapFS, []string) {
	fs := fstest.MapFS{
		"default.base.html": {Data: []byte(BaseDefaultLayoutContent + ` {{template "content" .}}`)},
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
