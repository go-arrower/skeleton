package testdata

import "testing/fstest"

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
    {{ block "layout" . }}
		defaultLayoutContextLayoutPlaceholder
        {{block "content" .}}
			defaultLayoutContextContentPlaceholder
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

var ContextViews = fstest.MapFS{
	"components/c0.html": {Data: []byte(`context c0`)},
	"pages/p0.html":      {Data: []byte(`context p0 {{template "c0" .}}`)},
	"pages/p1.html":      {Data: []byte(`context p1`)},
	"default.layout.html": {Data: []byte(`
    {{ define "layout" }}
		contextLayout
        {{block "content" . }}
			contextPlaceholder
        {{end}}
    {{end}}`)},
}

var ContextAdmin = fstest.MapFS{
	"default.layout.html": {Data: []byte(`
    {{ define "layout" }}
		adminLayout
        {{block "content" . }}
			adminPlaceholder
        {{end}}
    {{end}}`)},
}
