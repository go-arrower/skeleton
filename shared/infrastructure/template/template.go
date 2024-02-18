package template

import (
	"bytes"
	"context"
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
	ErrLoadFailed         = errors.New("load renderer failed")
	ErrInvalidFS          = fmt.Errorf("%w: invalid fs", ErrLoadFailed)
	ErrRenderFailed       = errors.New("rendering failed")
	ErrNotExistsComponent = fmt.Errorf("%w: component does not exist", ErrRenderFailed)
	ErrNotExistsPage      = fmt.Errorf("%w: page does not exist", ErrRenderFailed)
	ErrNotExistsFragment  = fmt.Errorf("%w: fragment does not exist", ErrRenderFailed)
	ErrNotExistsLayout    = fmt.Errorf("%w: layout does not exist", ErrRenderFailed)
	ErrTemplateNotExists  = fmt.Errorf("%w: template does not exist", ErrRenderFailed)
)

const (
	templateSeparator = "=>"
	fragmentSeparator = "#"
)

type Renderer struct {
	logger alog.Logger
	tracer trace.Tracer

	viewFS     fs.FS
	rawLayouts map[string]string
	rawPages   map[string]string

	contextViewFS map[string]fs.FS

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

	componentTemplates, pageTemplates, rawPages, rawLayouts, err := prepareRenderer(context.Background(), logger, viewFS)
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
		contextViewFS:     map[string]fs.FS{},
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

	for k := range rawLayouts {
		if k == "default" {
			defaultLayout = k

			break
		}
	}

	return defaultLayout
}

func prepareRenderer(ctx context.Context, logger alog.Logger, viewFS fs.FS) (*template.Template, map[string]*template.Template, map[string]string, map[string]string, error) {
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

	logger.LogAttrs(ctx, alog.LevelDebug,
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

	logger.LogAttrs(ctx, alog.LevelDebug,
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

	logger.LogAttrs(ctx, alog.LevelDebug,
		"loaded layouts",
		slog.Int("layout_count", len(rawLayouts)),
		slog.Any("layout_templates", rawTemplateNames(rawLayouts)),
	)

	return componentTemplates, pageTemplates, rawPages, rawLayouts, nil
}

// rawTemplateNames takes the names of the templates from the keys of the map.
func rawTemplateNames(pages map[string]string) []string {
	var names []string

	for k := range pages {
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
	name = strings.TrimSuffix(name, ".html")

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
	ctx := c.Request().Context()
	span := trace.SpanFromContext(ctx)

	_, innerSpan := span.TracerProvider().Tracer("arrower.renderer").Start(ctx, "render")
	defer innerSpan.End()

	isContextView, isAdmin, contextName := r.isRegisteredContext(c)
	fmt.Println("is context view:", isContextView, "isAdmin", isAdmin, ":>", contextName)

	origName := name
	layout, page := parseLayoutAndPage(strings.Split(name, "#")[0])

	if _, ok := r.rawPages[page]; ok && layout == "" {
		layout = r.defaultLayout
	}

	if _, ok := r.rawLayouts[layout]; layout != "" && !ok {
		return fmt.Errorf("%w", ErrNotExistsLayout)
	}

	cleanedName := layout + "=>" + page
	if layout == "" {
		cleanedName = page
	}

	r.logger.LogAttrs(ctx, alog.LevelInfo,
		"render template",
		slog.String("called_template", name),
		slog.String("actual_template", cleanedName),
		slog.Bool("is_context_view", isContextView),
		slog.String("context_view", contextName),
	)

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.hotReload {
		r.logger.LogAttrs(ctx, alog.LevelDebug, "hot reload all templates")

		componentTemplates, pageTemplates, rawPages, rawLayouts, err := prepareRenderer(ctx, r.logger, r.viewFS)
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
		r.logger.LogAttrs(ctx, alog.LevelDebug,
			"template not cached",
			slog.String("called_template", name),
			slog.String("layout", layout),
			slog.String("page", page),
		)

		newTemplate, err := r.components.Clone() // FIXME in prepare..() the page has already a clone of components=> might be unnecessary work
		if err != nil {
			return fmt.Errorf("%w", ErrTemplateNotExists)
		}

		_, err = newTemplate.New(cleanedName).Parse(r.rawLayouts[layout])
		if err != nil {
			return fmt.Errorf("%w: could not parse layout: %v", ErrRenderFailed, err)
		}

		if _, ok := r.rawPages[page]; !ok {
			newTemplate = r.components.Lookup(page)
			if newTemplate == nil && !isContextView { // TODO
				return fmt.Errorf("%w", ErrTemplateNotExists)
			}
		}

		if isContextView { // TODO
			goto outside
		}

		_, err = newTemplate.New("content").Parse(r.rawPages[page])
		if err != nil {
			return fmt.Errorf("%w: could not parse page: %v", ErrRenderFailed, err)
		}

		r.templates[cleanedName] = newTemplate
		templ = newTemplate // "found" the template

		//
		templ.Funcs(template.FuncMap{
			"reverse": c.Echo().Reverse,
		})

		r.logger.LogAttrs(ctx, alog.LevelInfo,
			"template cached",
			slog.String("called_template", name),
			slog.String("actual_template", cleanedName),
			slog.String("layout", layout),
			slog.String("page", page),
			slog.Any("templates", templateNames(templ)),
		)
	}

outside:
	if isContextView {
		fmt.Println("load context templates (force reload, as of debugging)")

		componentTemplates, pageTemplates, rawPages, rawLayouts, err := prepareRenderer(ctx, r.logger, r.contextViewFS[contextName])
		_ = componentTemplates
		_ = pageTemplates
		_ = rawPages
		_ = rawLayouts
		_ = err
		//fmt.Println(componentTemplates, pageTemplates, rawPages, rawLayouts)

		if layout == "" {
			layout = "default" // todo
		}
		cleanedName = layout + "=>default=>" + page
		fmt.Println(cleanedName)

		newTemplate, _ := r.components.Clone()

		// global layout
		_, err = newTemplate.New(cleanedName).Parse(r.rawLayouts[layout])
		//fmt.Println("load layout", err, r.rawLayouts[layout])

		// context layout
		if isAdmin {
			_, _, _, rawLayouts, _ := prepareRenderer(ctx, r.logger, r.contextViewFS["admin"])
			_, err = newTemplate.New(cleanedName).Parse(rawLayouts["default"]) // todo extract from template name
			fmt.Println("load context layout", isAdmin, err, rawLayouts["default"])
		} else {
			_, err = newTemplate.New(cleanedName).Parse(rawLayouts["default"]) // todo extract from template name
			fmt.Println("load context layout", isAdmin, err, rawLayouts["default"])
		}

		// page
		_, err = newTemplate.New("content").Parse(rawPages[page])
		//fmt.Println("load context page", err)

		r.templates[cleanedName] = newTemplate
		templ = newTemplate // "found" the template

		//
		templ.Funcs(template.FuncMap{
			"reverse": c.Echo().Reverse,
		})
	}

	renderTemplate := cleanedName

	p := strings.Split(origName, "#")
	if len(p) == 2 {
		renderTemplate = p[1]
	}

	{ // check if fragment exists
		found := false

		for _, f := range templ.Templates() { // todo use Lookup() instead
			if f.Name() == renderTemplate {
				found = true
			}
		}

		if !found {
			return fmt.Errorf("%w", ErrNotExistsFragment)
		}
	}

	err := templ.ExecuteTemplate(w, renderTemplate, data)
	if err != nil {
		return fmt.Errorf("%w: could not execute template: %v", ErrRenderFailed, err)
	}

	return nil
}

// isRegisteredContext returns if a call is to be rendered for a context registered via AddContext.
// If false => it is a shared view.
func (r *Renderer) isRegisteredContext(c echo.Context) (bool, bool, string) {
	paths := strings.Split(c.Path(), "/")

	isAdmin := false

	for _, p := range paths {
		if p == "" {
			continue
		}

		if p == "admin" {
			isAdmin = true
			continue
		}

		_, exists := r.contextViewFS[p]
		if exists {
			if isAdmin {
				return true, true, p
			}
			return true, false, p
		}
	}

	if isAdmin {
		return true, true, "admin"
	}

	return false, false, ""
}

// parseLayoutAndPage accepts:
// - page
// - layout=>page
// - layout=>sub-layout=>page
// and returns the layout (composed if with sub-layout) and the page.
func parseLayoutAndPage(name string) (string, string) {
	const maxCompositionSegments = 3 // how many segments after separated by the separator

	elem := strings.Split(name, templateSeparator)

	length := len(elem)

	if length > maxCompositionSegments { // invalid pattern
		return "", ""
	}

	if length == 1 {
		return "", strings.TrimSpace(elem[0])
	}

	l := strings.TrimSpace(elem[0])
	if length == maxCompositionSegments {
		l = l + templateSeparator + strings.TrimSpace(elem[1])
	}

	return l, strings.TrimSpace(elem[length-1])
}

type parsedTemplate struct {
	layout        string
	contextLayout string
	template      string
	fragment      string
}

func parseTemplateName(name string) (parsedTemplate, error) {
	const ( // todo combine with templateSeparator and fragmentSeparator
		maxCompositionSegments = 3 // how many segments after separated by the separator
		maxFragmentSegments    = 2
	)

	elem := strings.Split(name, templateSeparator)
	length := len(elem)

	if length > maxCompositionSegments { // invalid pattern
		return parsedTemplate{}, fmt.Errorf("%w", ErrRenderFailed)
	}

	var (
		layout        string
		contextLayout string
		tmpl          string
		fragment      string
	)

	if length == 1 {
		tmpl = strings.TrimSpace(elem[0])
	}

	if length == 2 {
		contextLayout = strings.TrimSpace(elem[0])
		tmpl = strings.TrimSpace(elem[1])
	}

	if length == 3 {
		layout = strings.TrimSpace(elem[0])
		contextLayout = strings.TrimSpace(elem[1])
		tmpl = strings.TrimSpace(elem[2])
	}

	fragments := strings.Split(tmpl, fragmentSeparator)
	if len(fragments) > maxFragmentSegments { // invalid pattern
		return parsedTemplate{}, fmt.Errorf("%w", ErrRenderFailed)
	}

	if len(fragments) == 2 {
		tmpl = strings.TrimSpace(fragments[0])
		fragment = strings.TrimSpace(fragments[1])

		if fragment == "" { // invalid pattern
			return parsedTemplate{}, fmt.Errorf("%w", ErrRenderFailed)
		}
	}

	isInvalid := func(s string) bool {
		return strings.Contains(s, templateSeparator) || strings.Contains(s, fragmentSeparator)
	}
	if isInvalid(layout) || isInvalid(contextLayout) || isInvalid(tmpl) || isInvalid(fragment) {
		return parsedTemplate{}, fmt.Errorf("%w", ErrRenderFailed)
	}

	return parsedTemplate{
		layout:        layout,
		contextLayout: contextLayout,
		template:      tmpl,
		fragment:      fragment,
	}, nil
}

// Layout returns the default layout of this renderer.
func (r *Renderer) Layout() string { // todo can be private
	return r.defaultLayout
}

// SetDefaultLayout sets the default layout.
func (r *Renderer) SetDefaultLayout(l string) error { // todo can be private (?)
	if _, ok := r.rawLayouts[l]; !ok {
		return ErrNotExistsLayout
	}

	r.defaultLayout = l

	return nil
}

func (r *Renderer) AddContext(name string, viewFS fs.FS) error {
	r.contextViewFS[name] = viewFS

	return nil
}
