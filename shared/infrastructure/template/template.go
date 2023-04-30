package template

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

var (
	ErrLoadFailed        = errors.New("load renderer failed")
	ErrInvalidFS         = fmt.Errorf("%w: invalid fs", ErrLoadFailed)
	ErrRenderFailed      = errors.New("rendering failed")
	ErrTemplateNotExists = errors.New("template does not exist")
)

const separator = "=>"

type Renderer struct {
	viewFS     fs.FS
	rawLayouts map[string]string
	rawPages   map[string]string
	templates  map[string]*template.Template

	components    *template.Template
	defaultLayout string

	isContextRenderer bool // true, if the renderer became a Context renderer and is not shared anymore.
	hotReload         bool

	mu sync.Mutex
}

// NewRenderer take multiple FS or can Context views be added later?
// It prepares a renderer for HTML web views.
func NewRenderer(viewFS fs.FS, hotReload bool) (*Renderer, error) {
	if viewFS == nil {
		return nil, ErrInvalidFS
	}

	componentTemplates, pageTemplates, rawPages, rawLayouts, err := prepareRenderer(viewFS)
	if err != nil {
		return nil, err
	}

	defaultLayout := getDefaultLayout(rawLayouts)

	return &Renderer{
		viewFS:            viewFS,
		rawLayouts:        rawLayouts,
		rawPages:          rawPages,
		isContextRenderer: false,
		components:        componentTemplates,
		defaultLayout:     defaultLayout,
		templates:         pageTemplates,
		hotReload:         hotReload,
		mu:                sync.Mutex{},
	}, nil
}

func getDefaultLayout(rawLayouts map[string]string) string {
	var defaultLayout string

	if len(rawLayouts) == 1 {
		for k := range rawLayouts {
			defaultLayout = k
		}
	} else {
		for k := range rawLayouts {
			if k == "default" {
				defaultLayout = k

				break
			}
		}
	}

	return defaultLayout
}

func prepareRenderer(viewFS fs.FS) (*template.Template, map[string]*template.Template, map[string]string, map[string]string, error) {
	components, err := fs.Glob(viewFS, "components/*.html")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("%w: could not get components from fs: %v", ErrInvalidFS, err)
	}

	componentTemplates := template.New("")

	for _, c := range components {
		file, err := readFile(viewFS, c) //nolint:govet // govet is too pedantic for shadowing errors
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("%w: could not read component file: %s: %v", ErrInvalidFS, file, err)
		}

		name := componentName(c)

		_, err = componentTemplates.New(name).Parse(file)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("%w: could not parse component: %s: %v", ErrInvalidFS, file, err)
		}
	}

	log.Println("loaded components", len(componentTemplates.Templates()), componentTemplates.DefinedTemplates())

	pageTemplates := make(map[string]*template.Template)

	pages, err := fs.Glob(viewFS, "pages/*.html")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("%w: could not get pages from fs: %v", ErrInvalidFS, err)
	}

	rawPages := make(map[string]string)

	for _, page := range pages {
		file, err := readFile(viewFS, page) //nolint:govet // govet is too pedantic for shadowing errors
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("%w: could not read page file: %s: %v", ErrInvalidFS, file, err)
		}

		tmp, err := componentTemplates.Clone()
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("%w: could not clone component templates: %v", ErrLoadFailed, err)
		}

		pn := pageName(page)
		rawPages[pn] = file

		pageTemplates[pn], err = tmp.New(pn).Parse(file)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("%w: could not parse page file: %s: %v", ErrInvalidFS, file, err)
		}
	}

	log.Println("loaded pages", len(pageTemplates))

	for _, p := range pageTemplates {
		log.Println("page:", p.Name(), p.DefinedTemplates())
	}

	layouts, err := fs.Glob(viewFS, "*.html")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("%w: could not get layouts from fs: %v", ErrInvalidFS, err)
	}

	rawLayouts := make(map[string]string)

	for _, l := range layouts {
		file, err := readFile(viewFS, l)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("%w: could not read layout file: %s: %v", ErrInvalidFS, file, err)
		}

		ln := layoutName(l)
		rawLayouts[ln] = file
	}

	log.Println("layouts", rawLayouts)

	return componentTemplates, pageTemplates, rawPages, rawLayouts, nil
}

func componentName(componentName string) string {
	name := strings.TrimPrefix(componentName, "components/")
	name = strings.TrimSuffix(name, ".html")

	return name
}

func layoutName(layoutName string) string {
	name := strings.TrimSuffix(layoutName, ".layout.html")

	return name
}

func pageName(pageName string) string {
	name := strings.TrimPrefix(pageName, "pages/")
	name = strings.TrimSuffix(name, ".page.html")

	return name
}

func readFile(sfs fs.FS, name string) (string, error) {
	file, err := sfs.Open(name)
	if err != nil {
		return "", fmt.Errorf("%v", err) //nolint:errorlint,goerr113 // do not expose err to arrower api
	}

	var buf bytes.Buffer

	_, err = io.Copy(&buf, file)
	if err != nil {
		return "", fmt.Errorf("could not read: %v", err) //nolint:errorlint,goerr113 // do not expose err to arrower api
	}

	return buf.String(), nil
}

func (r *Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	layout, page := parseLayoutAndPage(name)

	if strings.HasPrefix(name, separator) {
		layout = r.defaultLayout
	}

	if _, ok := r.rawLayouts[layout]; layout != "" && !ok {
		return fmt.Errorf("%w: layout does not exist", ErrRenderFailed)
	}

	cleanedName := layout + "=>" + page
	if layout == "" {
		cleanedName = page
	}

	log.Println("Render for template", name, "cleanedName", cleanedName)

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.hotReload {
		componentTemplates, pageTemplates, rawPages, rawLayouts, err := prepareRenderer(r.viewFS)
		if err != nil {
			return err
		}

		r.rawLayouts = rawLayouts
		r.defaultLayout = getDefaultLayout(rawLayouts)
		r.rawPages = rawPages
		r.components = componentTemplates
		r.templates = pageTemplates
	}

	templ, found := r.templates[name]
	if !found || r.hotReload {
		log.Println("template does not exist", name)

		log.Println("layout:", layout, "page:", page)

		newTemplate, err := r.components.Clone()
		if err != nil {
			return fmt.Errorf("%w: could not clone: %v", ErrRenderFailed, err)
		}

		log.Println("newTemplate details:", newTemplate.Name(), newTemplate.DefinedTemplates())

		_, err = newTemplate.New(cleanedName).Parse(r.rawLayouts[layout])
		if err != nil {
			return fmt.Errorf("%w: could not parse layout: %v", ErrRenderFailed, err)
		}

		log.Println("newTemplate details:", newTemplate.Name(), newTemplate.DefinedTemplates())

		if _, ok := r.rawPages[page]; !ok && !strings.HasSuffix(page, ".component") {
			return fmt.Errorf("%w: page does not exist", ErrRenderFailed)
		}

		_, err = newTemplate.New("content").Parse(r.rawPages[page])
		if err != nil {
			return fmt.Errorf("%w: could not parse page: %v", ErrRenderFailed, err)
		}

		r.templates[cleanedName] = newTemplate
		templ = newTemplate // "found" the template

		log.Println("newTemplate added:", newTemplate.Name(), newTemplate.DefinedTemplates())
	}

	log.Println("found template", templ.Name(), templ.DefinedTemplates())

	err := templ.ExecuteTemplate(w, cleanedName, data)
	if err != nil {
		return fmt.Errorf("%w: could not execute template: %v", ErrRenderFailed, err)
	}

	return nil
}

// parseLayoutAndPage accepts:
// - page
// - layout=>page
// - layout=>sub-layout=>page
// and returns the layout (composed if with sub-layout) and the page.
func parseLayoutAndPage(name string) (string, string) {
	const maxCompositionSegments = 3 // how many segments after separated by the separator

	elem := strings.Split(name, separator)

	length := len(elem)

	if length > maxCompositionSegments { // invalid pattern
		return "", ""
	}

	if length == 1 {
		return "", strings.TrimSpace(elem[0])
	}

	l := strings.TrimSpace(elem[0])
	if length == maxCompositionSegments {
		l = l + separator + strings.TrimSpace(elem[1])
	}

	return l, strings.TrimSpace(elem[length-1])
}

// Layout returns the default layout of this renderer.
func (r *Renderer) Layout() string {
	return r.defaultLayout
}

// SetDefaultLayout sets the default layout.
func (r *Renderer) SetDefaultLayout(l string) error {
	if _, ok := r.rawLayouts[l]; !ok {
		return ErrTemplateNotExists
	}

	r.defaultLayout = l

	return nil
}
