package views

import (
	"embed"
)

//go:embed *.html components/*.html pages/*.html
var SharedViews embed.FS
