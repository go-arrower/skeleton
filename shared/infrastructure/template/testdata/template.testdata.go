// the template package uses white-box tests, so this is not a _test package.
package testdata

import "testing/fstest"

const (
	C0Content = "c0"
	C1Content = "c1"
	P0Content = "p0"
	P1Content = "p1"
	LContent  = "layout"
)

var EmptyFiles = fstest.MapFS{}

var SimpleFiles = fstest.MapFS{
	"components/c0.component.html": {Data: []byte(C0Content)},
	"components/c1.component.html": {Data: []byte(C1Content)},
	"pages/p0.page.html":           {Data: []byte(P0Content)},
	"pages/p1.page.html":           {Data: []byte(P1Content)},
	"global.layout.html":           {Data: []byte(LContent)},
}

var LayoutsPagesAndComponents = fstest.MapFS{
	"components/c0.component.html": {Data: []byte(C0Content)},
	"components/c1.component.html": {Data: []byte(C1Content)},
	"pages/p0.page.html":           {Data: []byte(P0Content + ` {{template "c0.component" .}}`)},
	"pages/p1.page.html":           {Data: []byte(P1Content)},
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
	"other.layout.html": {Data: []byte(`other
    {{ block "layout" . }}
        {{block "content" .}}
            Fallback, if "content" is not defined elsewhere
        {{end}}
    {{end}}`)},
}
