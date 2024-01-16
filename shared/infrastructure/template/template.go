package template

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"strings"
	"sync"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-arrower/arrower/alog"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrLoadFailed        = errors.New("load renderer failed")
	ErrInvalidFS         = fmt.Errorf("%w: invalid fs", ErrLoadFailed)
	ErrRenderFailed      = errors.New("rendering failed")
	ErrTemplateNotExists = errors.New("template does not exist")
)

const separator = "=>"

type Renderer struct {
	logger alog.Logger
	tracer trace.Tracer

	viewFS     fs.FS
	rawLayouts map[string]string
	rawPages   map[string]string

	templates     map[string]*template.Template
	components    *template.Template
	defaultLayout string

	isContextRenderer bool // true, if the renderer became a Context renderer and is not shared anymore.
	hotReload         bool

	mu sync.Mutex
}

// NewRenderer take multiple FS or can Context views be added later?
// It prepares a renderer for HTML web views.
func NewRenderer(logger alog.Logger, traceProvider trace.TracerProvider, viewFS fs.FS, hotReload bool) (*Renderer, error) {
	if viewFS == nil {
		return nil, ErrInvalidFS
	}

	logger = logger.WithGroup("arrower.renderer")
	tracer := traceProvider.Tracer("arrower.renderer")

	componentTemplates, pageTemplates, rawPages, rawLayouts, err := prepareRenderer(logger, viewFS)
	if err != nil {
		return nil, err
	}

	defaultLayout := getDefaultLayout(rawLayouts)

	logger.LogAttrs(nil, alog.LevelInfo,
		"renderer created",
		slog.Bool("hot_reload", hotReload),
		slog.String("default_layout", defaultLayout),
	)

	return &Renderer{
		logger:            logger,
		tracer:            tracer,
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

func prepareRenderer(logger alog.Logger, viewFS fs.FS) (*template.Template, map[string]*template.Template, map[string]string, map[string]string, error) {
	components, err := fs.Glob(viewFS, "components/*.html")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("%w: could not get components from fs: %v", ErrInvalidFS, err)
	}

	componentTemplates := template.New("<empty>").Funcs(sprig.FuncMap())

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

	logger.LogAttrs(nil, alog.LevelDebug,
		"loaded components",
		slog.Int("component_count", len(componentTemplates.Templates())),
		slog.Any("component_templates", templateNames(componentTemplates)),
	)

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

	logger.LogAttrs(nil, alog.LevelDebug,
		"loaded pages",
		slog.Int("page_count", len(pageTemplates)),
		slog.Any("page_templates", rawTemplateNames(rawPages)),
	)

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

	logger.LogAttrs(nil, alog.LevelDebug,
		"loaded layouts",
		slog.Int("layout_count", len(rawLayouts)),
		slog.Any("layout_templates", rawTemplateNames(rawLayouts)),
	)

	return componentTemplates, pageTemplates, rawPages, rawLayouts, nil
}

// rawTemplateNames takes the names of the templates from the keys of the map.
func rawTemplateNames(pages map[string]string) []string {
	var names []string

	for k, _ := range pages {
		names = append(names, k)
	}

	return names
}

func templateNames(templates *template.Template) []string {
	n := len(templates.Templates())
	var names = make([]string, n)

	for i, t := range templates.Templates() {
		names[i] = t.Name()
	}

	return names
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
	_, span := r.tracer.Start(c.Request().Context(), "render")
	defer span.End()

	origName := name
	layout, page := parseLayoutAndPage(strings.Split(name, "#")[0])

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

	r.logger.LogAttrs(nil, alog.LevelInfo,
		"render template",
		slog.String("called_template", name),
		slog.String("actual_template", cleanedName),
	)

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.hotReload {
		r.logger.LogAttrs(nil, alog.LevelDebug, "hot reload all templates")

		componentTemplates, pageTemplates, rawPages, rawLayouts, err := prepareRenderer(r.logger, r.viewFS)
		if err != nil {
			return err
		}

		r.rawLayouts = rawLayouts
		r.defaultLayout = getDefaultLayout(rawLayouts)
		r.rawPages = rawPages
		r.components = componentTemplates
		r.templates = pageTemplates
	}

	//if strings.HasSuffix(cleanedName, ".component") {
	//	err := r.components.ExecuteTemplate(w, cleanedName, data)
	//	if err != nil {
	//		return fmt.Errorf("%w: could not execute component template: %v", ErrRenderFailed, err)
	//	}
	//
	//	return nil
	//}

	templ, found := r.templates[cleanedName]
	if !found || r.hotReload {
		r.logger.LogAttrs(nil, alog.LevelDebug,
			"template not cached",
			slog.String("called_template", name),
			slog.String("layout", layout),
			slog.String("page", page),
		)

		newTemplate, err := r.components.Clone() // FIXME in prepare..() the page has already a clone of components=> might be unnecessary work
		if err != nil {
			return fmt.Errorf("%w: could not clone: %v", ErrRenderFailed, err)
		}

		_, err = newTemplate.New(cleanedName).Parse(r.rawLayouts[layout])
		if err != nil {
			return fmt.Errorf("%w: could not parse layout: %v", ErrRenderFailed, err)
		}

		if _, ok := r.rawPages[page]; !ok && !strings.HasSuffix(page, ".component") {
			return fmt.Errorf("%w: page does not exist", ErrRenderFailed)
		}

		_, err = newTemplate.New("content").Parse(r.rawPages[page])
		if err != nil {
			return fmt.Errorf("%w: could not parse page: %v", ErrRenderFailed, err)
		}

		r.templates[cleanedName] = newTemplate
		templ = newTemplate // "found" the template

		r.logger.LogAttrs(nil, alog.LevelInfo,
			"template cached",
			slog.String("called_template", name),
			slog.String("actual_template", cleanedName),
			slog.String("layout", layout),
			slog.String("page", page),
			slog.Any("templates", templateNames(templ)),
		)
	}

	renderTemplate := cleanedName

	p := strings.Split(origName, "#")
	if len(p) == 2 {
		renderTemplate = p[1]
	}
	// return error in the else case

	err := templ.ExecuteTemplate(w, renderTemplate, data)
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
