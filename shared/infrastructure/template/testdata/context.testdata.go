package testdata

import "testing/fstest"

const (
	LDefaultContextContent = "defaultContextLayout"
	C0ContextContent       = "context component 0"
	P0ContextContent       = "context p0"
)

var ExampleContext = "example"

var SharedViews = fstest.MapFS{
	"components/c0.html":   {Data: []byte(C0Content)},
	"components/c1.html":   {Data: []byte(C1Content)},
	"pages/shared-p0.html": {Data: []byte(P0Content + ` {{template "c0" .}}`)},
	"pages/shared-p1.html": {Data: []byte(P1Content)},
	"default.layout.html": {Data: []byte(`<!DOCTYPE html>
<html lang="en">
<body>
	defaultLayout
    {{block "layout" .}}
		defaultContextLayout
        {{block "content" .}}
			contentPlaceholder
        {{end}}
    {{end}}
</body>
</html>`)},
	"other.layout.html": {Data: []byte(`otherLayout
    {{block "layout" .}}
        {{block "content" .}}
            contentPlaceholder
        {{end}}
    {{end}}`)},
}

var ContextViews = fstest.MapFS{
	"components/c0.html": {Data: []byte(C0ContextContent)},
	"pages/p0.html":      {Data: []byte(P0ContextContent + ` {{template "c0" .}}`)},
	"pages/p1.html":      {Data: []byte(`context p1 {{block "f" . }}fragment{{end}}`)},
	"default.layout.html": {Data: []byte(`
    {{define "layout"}}
		defaultContextLayout
        {{block "content" .}}
			contentPlaceholder
        {{end}}
    {{end}}`)},
}

var ContextAdmin = fstest.MapFS{
	"default.layout.html": {Data: []byte(`
    {{define "layout"}}
		adminLayout
        {{block "content" .}}
			adminPlaceholder
        {{end}}
    {{end}}`)},
}
