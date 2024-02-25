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
	ErrContextNotAdded    = errors.New("context not added")
)

const (
	templateSeparator = "=>"
	fragmentSeparator = "#"
)

type Renderer struct {
	logger alog.Logger
	tracer trace.Tracer

	views map[string]viewTemplates

	hotReload bool

	mu    sync.Mutex // todo remove or protect views (?)
	cache sync.Map
}

type viewTemplates struct {
	viewFS fs.FS

	rawLayouts    map[string]string // todo can this be removed and read from viewFS on demand?
	rawPages      map[string]string
	defaultLayout string

	components *template.Template
}

// NewRenderer take multiple FS or can Context views be added later?
// It prepares a renderer for HTML web views.
func NewRenderer(logger alog.Logger, traceProvider trace.TracerProvider, viewFS fs.FS, hotReload bool) (*Renderer, error) {
	if viewFS == nil {
		return nil, ErrInvalidFS
	}

	logger = logger.WithGroup("arrower.renderer")
	tracer := traceProvider.Tracer("arrower.renderer")

	view := map[string]viewTemplates{}
	view[""], _ = prepareView(context.Background(), logger, viewFS) // todo err check

	logger.LogAttrs(nil, alog.LevelInfo,
		"renderer created",
		slog.Bool("hot_reload", hotReload),
		slog.String("default_layout", view[""].defaultLayout),
	)

	return &Renderer{
		logger:    logger,
		tracer:    tracer,
		hotReload: hotReload,
		mu:        sync.Mutex{},
		cache:     sync.Map{},
		views:     view,
	}, nil
}

func prepareView(ctx context.Context, logger alog.Logger, viewFS fs.FS) (viewTemplates, error) {
	components, err := fs.Glob(viewFS, "components/*.html")
	if err != nil {
		return viewTemplates{}, fmt.Errorf("%w: could not get components from fs: %v", ErrInvalidFS, err)
	}

	componentTemplates := template.New("<empty>").Funcs(sprig.FuncMap())

	for _, c := range components {
		file, err := readFile(viewFS, c) //nolint:govet // govet is too pedantic for shadowing errors
		if err != nil {
			return viewTemplates{}, fmt.Errorf("%w: could not read component file: %s: %v", ErrInvalidFS, file, err)
		}

		name := componentName(c)

		_, err = componentTemplates.New(name).Parse(file)
		if err != nil {
			return viewTemplates{}, fmt.Errorf("%w: could not parse component: %s: %v", ErrInvalidFS, file, err)
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
		return viewTemplates{}, fmt.Errorf("%w: could not get pages from fs: %v", ErrInvalidFS, err)
	}

	rawPages := make(map[string]string)

	for _, page := range pages {
		file, err := readFile(viewFS, page) //nolint:govet // govet is too pedantic for shadowing errors
		if err != nil {
			return viewTemplates{}, fmt.Errorf("%w: could not read page file: %s: %v", ErrInvalidFS, file, err)
		}

		tmp, err := componentTemplates.Clone()
		if err != nil {
			return viewTemplates{}, fmt.Errorf("%w: could not clone component templates: %v", ErrLoadFailed, err)
		}

		pn := pageName(page)
		rawPages[pn] = file

		pageTemplates[pn], err = tmp.New(pn).Parse(file)
		if err != nil {
			return viewTemplates{}, fmt.Errorf("%w: could not parse page file: %s: %v", ErrInvalidFS, file, err)
		}
	}

	logger.LogAttrs(ctx, alog.LevelDebug,
		"loaded pages",
		slog.Int("page_count", len(pageTemplates)),
		slog.Any("page_templates", rawTemplateNames(rawPages)),
	)

	layouts, err := fs.Glob(viewFS, "*.html")
	if err != nil {
		return viewTemplates{}, fmt.Errorf("%w: could not get layouts from fs: %v", ErrInvalidFS, err)
	}

	var defaultLayout string
	rawLayouts := make(map[string]string)

	for _, l := range layouts {
		file, err := readFile(viewFS, l)
		if err != nil {
			return viewTemplates{}, fmt.Errorf("%w: could not read layout file: %s: %v", ErrInvalidFS, file, err)
		}

		ln := layoutName(l)
		rawLayouts[ln] = file

		if ln == "default" {
			defaultLayout = "default"
		}
	}

	//if len(layouts) == 1 {
	//	defaultLayout = layoutName(layouts[0])
	//} // FIXME: decide if a default template should be set or not, both has it's advantages, but maybe be less "magic" and only accept it as default if it is called "default" to prevent breking, old pages, ones a default layout get's created

	logger.LogAttrs(ctx, alog.LevelDebug,
		"loaded layouts",
		slog.Int("layout_count", len(rawLayouts)),
		slog.Any("layout_templates", rawTemplateNames(rawLayouts)),
	)

	return viewTemplates{
		viewFS:        viewFS,
		rawLayouts:    rawLayouts,
		rawPages:      rawPages,
		defaultLayout: defaultLayout,
		components:    componentTemplates,
	}, nil
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

	parsedTempl, _ := r.getParsedTemplate(c, name)

	if parsedTempl.context == "" { // todo should this be in parseTemplateName?
		parsedTempl.layout = parsedTempl.contextLayout
		parsedTempl.contextLayout = ""
	}

	isComponent := parsedTempl.isComponent
	if isComponent { // todo cache and check cache first (?)
		fmt.Println("IS COMPONENT", parsedTempl.key())
		t := r.views[parsedTempl.context].components.Lookup(parsedTempl.fragment)
		if t == nil {
			return ErrNotExistsComponent
		}

		t.AddParseTree(parsedTempl.key(), t.Tree)
		r.cache.Store(parsedTempl.key(), t)
	}
	if !isComponent && parsedTempl.layout == "" {
		parsedTempl.layout = r.views[""].defaultLayout
	}

	fmt.Printf("RENDER TEMPLATE %+v %s \n", parsedTempl, parsedTempl.key())
	var templ *template.Template
	_, found := r.cache.Load(parsedTempl.key())
	if !found {
		fmt.Println("template not found => generate it")
		/*
			if component => already handled above
			if page with layout(s) => build template
			if page without layouts => build template by using a default layout
		*/

		var err error
		r.views[parsedTempl.context], err = prepareView(ctx, r.logger, r.views[parsedTempl.context].viewFS)

		newTemplate, err := r.views[parsedTempl.context].components.Clone()
		if err != nil {
			return fmt.Errorf("%w", ErrTemplateNotExists)
		}

		//isPageWithoutLayout := parsedTempl.layout == "" && parsedTempl.contextLayout == ""

		if "" == r.views[""].rawLayouts[parsedTempl.layout] {
			return fmt.Errorf("%w: default", ErrNotExistsLayout)
		}

		newTemplate, err = newTemplate.New(parsedTempl.key()).Parse(r.views[""].rawLayouts[parsedTempl.layout])
		if err != nil {
			return fmt.Errorf("%w: could not parse default layout: %v", ErrRenderFailed, err)
		}

		if parsedTempl.renderAsAdminPage {
			if "" == r.views["admin"].rawLayouts[parsedTempl.layout] {
				return ErrNotExistsLayout
			}

			newTemplate, err = newTemplate.New("layout").Parse(r.views["admin"].rawLayouts[parsedTempl.contextLayout])
			if err != nil {
				return fmt.Errorf("%w: could not parse admin layout: %v", ErrRenderFailed, err)
			}
		} else {
			if parsedTempl.isComponent {
				if "" == r.views[parsedTempl.context].rawLayouts[parsedTempl.layout] {
					return ErrNotExistsLayout
				}

				newTemplate, err = newTemplate.New("layout").Parse(r.views[parsedTempl.context].rawLayouts[parsedTempl.contextLayout])
				if err != nil {
					return fmt.Errorf("%w: could not parse context layout: %v", ErrRenderFailed, err)
				}
			}
		}

		pageExists := r.views[parsedTempl.context].rawPages[parsedTempl.template] != ""
		if !pageExists {
			return ErrTemplateNotExists
		}

		newTemplate, err = newTemplate.New("content").Parse(r.views[parsedTempl.context].rawPages[parsedTempl.template])
		if err != nil {
			return fmt.Errorf("%w: could not parse page: %v", ErrRenderFailed, err)
		}

		newTemplate.Funcs(template.FuncMap{
			"reverse": c.Echo().Reverse,
		})

		r.cache.Store(parsedTempl.key(), newTemplate)
		templ = newTemplate // "found" the template // todo set t instead to save second lookup from the sync.Map

		r.logger.LogAttrs(ctx, alog.LevelInfo,
			"template cached",
			slog.String("called_template", name),
			//slog.String("actual_template", cleanedName),
			//slog.String("layout", layout),
			//slog.String("page", page),
			slog.Any("templates", templateNames(templ)),
		)
	}

	/*
		check if cached or hot reload is on
		cache: context/gl=>cl=>p CALL key() on parsedTemplate
		generate template and cache

		render template or fragment
		(?) htmx support for partial rendering
	*/

	t, _ := r.cache.Load(parsedTempl.key())
	templ = t.(*template.Template)

	dumpAllNamedTemplatesRenderedWithData(templ, data)
	fmt.Println("RENDER FOR", parsedTempl.templateName())

	if nil == templ.Lookup(parsedTempl.templateName()) {
		return ErrNotExistsFragment
	}

	err := templ.ExecuteTemplate(w, parsedTempl.templateName(), data)
	if err != nil {
		return fmt.Errorf("%w: could not execute template: %v", ErrRenderFailed, err)
	}

	return nil
}

// isRegisteredContext returns if a call is to be rendered for a context registered via AddContext.
// If false => it is a shared view. // TODO refactor
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

		_, exists := r.views[p]
		if exists {
			if isAdmin {
				return true, true, p
			}
			return true, false, p
		}
	}

	if isAdmin {
		return true, true, "admin" // todo return normal context name, as the flag isAdmin is returned already
	}

	return false, false, ""
}

type parsedTemplate struct {
	context           string
	layout            string
	contextLayout     string
	template          string
	fragment          string
	renderAsAdminPage bool
	isComponent       bool
}

func (t parsedTemplate) key() string {
	if t.isComponent {
		return fmt.Sprintf("%s/%s", t.context, t.fragment)
	}

	return fmt.Sprintf("%s/%s%s%s%s%s", t.context, t.layout, templateSeparator, t.contextLayout, templateSeparator, t.template)
}

func (t parsedTemplate) templateName() string {
	if t.fragment != "" {
		return t.fragment
	}

	return t.key()
}

func (r *Renderer) getParsedTemplate(c echo.Context, name string) (parsedTemplate, error) {
	parsedTempl, _ := parseTemplateName(name)

	isContext, isAdmin, contextName := r.isRegisteredContext(c)
	parsedTempl.renderAsAdminPage = isAdmin
	parsedTempl.context = contextName

	if isContext {
		parsedTempl.contextLayout = r.views[contextName].defaultLayout // todo: only if it's a page and not component?
	}

	return parsedTempl, nil
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
		isComponent   bool
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

	if len(name) > 0 {
		isComponent = string(name[0]) == fragmentSeparator
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
		isComponent:   isComponent,
	}, nil
}

func (r *Renderer) AddContext(name string, viewFS fs.FS) error {
	if name == "" {
		return fmt.Errorf("%w: set a name", ErrContextNotAdded)
	}

	if viewFS == nil {
		return fmt.Errorf("%w: no view files", ErrContextNotAdded)
	}

	if _, exists := r.views[name]; exists {
		return fmt.Errorf("%w: already added", ErrContextNotAdded)
	}

	// todo clean
	r.views[name], _ = prepareView(context.Background(), r.logger, viewFS)
	m := r.views[name]
	cc, _ := r.views[""].components.Clone()

	for _, t := range m.components.Templates() {
		cc, _ = m.components.AddParseTree(t.Name(), t.Tree)
	}

	m.components = cc
	r.views[name] = m

	return nil
}

// --- --- ---
//
// Helpers used for white box testing.
// Hopefully, these functions make it harder to break the (partially useful) tests
// if larger refactoring is done on the Renderer's structure.
// Feel free to delete them anytime! Don't feel forced to test implementation detail!
//
// --- --- ---

// layout returns the default layout of this renderer.
func (r *Renderer) layout() string {
	return r.views[""].defaultLayout
}

func (r *Renderer) viewsForContext(name string) viewTemplates {
	return r.views[name]
}

func (r *Renderer) totalCachedTemplates() int { //nolint
	c := 0
	r.cache.Range(func(_, _ any) bool {
		c++
		return true
	})
	return c
}

// rawTemplateNames takes the names of the templates from the keys of the map.
func rawTemplateNames(pages map[string]string) []string {
	var names []string

	for k := range pages {
		names = append(names, k)
	}

	return names
}

// dumpAllNamedTemplatesRenderedWithData pretty prints all templates
// within the given *template.Template. Use it for convenient debugging.
func dumpAllNamedTemplatesRenderedWithData(templ *template.Template, data interface{}) {
	templ, err := templ.Clone() // ones ExecuteTemplate is called the template cannot be pared any more and could fail calling code.
	if err != nil {
		fmt.Println("CAN NOT DUMP TEMPLATE: ", err)
		return
	}

	fmt.Println()
	fmt.Println("--- --- ---   --- --- ---   --- --- ---")
	fmt.Println("--- --- ---   Render all templates:", strings.TrimPrefix(templ.DefinedTemplates(), "; defined templates are: "))
	fmt.Println("--- --- ---   --- --- ---   --- --- ---")

	buf := &bytes.Buffer{}
	for _, t := range templ.Templates() {
		fmt.Printf("--- --- ---   %s:\n", t.Name())

		templ.ExecuteTemplate(buf, t.Name(), data)

		fmt.Println(buf.String())
		buf.Reset()
	}

	fmt.Println("--- --- ---   --- --- ---   --- --- ---")
	fmt.Println()
}
