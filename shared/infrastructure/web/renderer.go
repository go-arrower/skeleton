package web

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
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrCreateRendererFailed = errors.New("create renderer failed")

	ErrRenderFailed       = errors.New("rendering failed")
	ErrNotExistsComponent = fmt.Errorf("%w: component does not exist", ErrRenderFailed)
	ErrNotExistsPage      = fmt.Errorf("%w: page does not exist", ErrRenderFailed)
	ErrNotExistsFragment  = fmt.Errorf("%w: fragment does not exist", ErrRenderFailed)
	ErrNotExistsLayout    = fmt.Errorf("%w: layout does not exist", ErrRenderFailed) // todo new err for base as well?
	ErrContextNotAdded    = errors.New("context not added")
)

const (
	SharedViews = ""

	templateSeparator = "=>"
	fragmentSeparator = "#"
)

// NewRenderer prepares a renderer for HTML web views.
func NewRenderer(logger alog.Logger, traceProvider trace.TracerProvider, viewFS fs.FS, funcMap template.FuncMap, hotReload bool) (*Renderer, error) {
	if viewFS == nil {
		return nil, fmt.Errorf("%w: missing views", ErrCreateRendererFailed)
	}

	ctx := context.Background()

	logger = logger.WithGroup("arrower.renderer")
	tracer := traceProvider.Tracer("arrower.renderer")

	views := map[string]viewTemplates{}

	view, err := prepareViewTemplates(ctx, logger, viewFS, funcMap, false)
	if err != nil {
		return nil, fmt.Errorf("%w: could not load views: %w", ErrCreateRendererFailed, err)
	}

	views[SharedViews] = view

	logger.LogAttrs(ctx, alog.LevelInfo,
		"renderer created",
		slog.Bool("hot_reload", hotReload),
		slog.String("default_layout", views[SharedViews].defaultLayout),
	)

	return &Renderer{
		logger:    logger,
		tracer:    tracer,
		cache:     sync.Map{},
		mu:        sync.Mutex{},
		views:     views,
		funcMap:   funcMap,
		hotReload: hotReload,
	}, nil
}

type Renderer struct {
	logger alog.Logger
	tracer trace.Tracer

	cache sync.Map

	mu    sync.Mutex
	views map[string]viewTemplates

	funcMap   template.FuncMap
	hotReload bool
}

type viewTemplates struct {
	viewFS fs.FS

	rawLayouts    map[string]string // todo can this be removed and read from viewFS on demand?
	rawPages      map[string]string
	defaultLayout string

	components *template.Template
}

func (r *Renderer) Render(ctx context.Context, w io.Writer, contextName string, templateName string, data interface{}) error {
	span := trace.SpanFromContext(ctx)

	_, innerSpan := span.TracerProvider().Tracer("arrower.renderer").Start(ctx, "render")
	defer innerSpan.End()

	if r.hotReload {
		r.mu.Lock() // todo is this lock still reqired, as the cache is delted via range now. Instead of = sync.Map{} of previous implementation => for the r.views

		// delete all keys
		r.cache.Range(func(key interface{}, value interface{}) bool {
			r.cache.Delete(key)
			return true
		})

		for k, v := range r.views {
			isCont, _, _ := isContext(k)
			r.views[k], _ = prepareViewTemplates(context.Background(), r.logger, v.viewFS, r.funcMap, isCont)
		}

		r.mu.Unlock()
	}

	parsedTempl, err := r.getParsedTemplate(contextName, templateName)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	r.logger.LogAttrs(ctx, alog.LevelInfo,
		"render template",
		slog.String("original_name", templateName),
		slog.String("cache_key", parsedTempl.key()),
	)

	var templ *template.Template

	t, found := r.cache.Load(parsedTempl.key())
	if found {
		templ = t.(*template.Template)
	} else {
		isContext := parsedTempl.context != ""
		newTemplate, err := r.buildPageTemplate(isContext, parsedTempl)
		if err != nil {
			return err
		}

		r.cache.Store(parsedTempl.key(), newTemplate)
		templ = newTemplate

		r.logger.LogAttrs(ctx, alog.LevelInfo,
			"template cached",
			slog.String("original_name", templateName),
			slog.String("cache_key", parsedTempl.key()),
			slog.Any("templates", templateNames(templ)),
		)
	}

	/*
		(?) htmx support for partial rendering
	*/

	if nil == templ.Lookup(parsedTempl.templateName()) {
		return ErrNotExistsFragment
	}

	err = templ.ExecuteTemplate(w, parsedTempl.templateName(), data)
	if err != nil {
		return fmt.Errorf("%w: could not execute template: %v", ErrRenderFailed, err) //nolint:errorlint // prevent err in api
	}

	return nil
}

func isContext(context string) (bool, bool, string) {
	if context == "" {
		return false, false, SharedViews
	}

	if strings.HasPrefix(strings.TrimPrefix(context, "/"), "admin/") {
		return true, true, strings.TrimPrefix(strings.TrimPrefix(context, "/"), "admin/")
	}

	return true, false, context
}

func (r *Renderer) getParsedTemplate(context string, templateName string) (parsedTemplate, error) {
	r.mu.Lock() // todo this is in the "hot path" could this be removed?
	defer r.mu.Unlock()

	parsedTempl, err := parseTemplateName(templateName)
	if err != nil {
		return parsedTemplate{}, fmt.Errorf("could not parse template name: %w", err)
	}

	isContext, isAdmin, contextName := isContext(context)

	parsedTempl.context = contextName
	parsedTempl.renderAsAdminPage = isAdmin

	if isContext {
		parsedTempl.contextLayout = r.views[contextName].defaultLayout
	}

	isSharedView := parsedTempl.context == SharedViews
	if isSharedView {
		parsedTempl.baseLayout = parsedTempl.contextLayout
		parsedTempl.contextLayout = ""
	}

	if !parsedTempl.isComponent && parsedTempl.baseLayout == "" {
		parsedTempl.baseLayout = r.views[SharedViews].defaultLayout
	}

	return parsedTempl, nil
}

func (r *Renderer) buildPageTemplate(isContext bool, parsedTempl parsedTemplate) (*template.Template, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if parsedTempl.isComponent {
		newTemplate := r.views[parsedTempl.context].components.Lookup(parsedTempl.fragment)
		if newTemplate == nil {
			return nil, ErrNotExistsComponent
		}

		newTemplate, err := newTemplate.AddParseTree(parsedTempl.key(), newTemplate.Tree)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrRenderFailed, err)
		}

		return newTemplate.Clone()
	}

	newTemplate, err := r.views[parsedTempl.context].components.Clone()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRenderFailed, err)
	}

	isPageWithoutLayout := parsedTempl.baseLayout == "" && parsedTempl.contextLayout == ""
	if isPageWithoutLayout {
		newTemplate, _ = newTemplate.New(parsedTempl.key()).Parse(`{{block "content" .}}{{end}}`)
	} else {
		if _, found := r.views[SharedViews].rawLayouts[parsedTempl.baseLayout]; !found {
			return nil, fmt.Errorf("%w: default", ErrNotExistsLayout)
		}

		newTemplate, err = newTemplate.New(parsedTempl.key()).Parse(r.views[SharedViews].rawLayouts[parsedTempl.baseLayout])
		if err != nil {
			return nil, fmt.Errorf("%w: could not parse base: %v", ErrRenderFailed, err) //nolint:errorlint,lll // prevent err in api
		}

		if isContext {
			newTemplate, err = newTemplate.New("layout").Parse(r.views[parsedTempl.context].rawLayouts[parsedTempl.contextLayout])
			if err != nil {
				return nil, fmt.Errorf("%w: could not parse layout: %v", ErrRenderFailed, err) //nolint:errorlint,lll // prevent err in api
			}
		}

		if parsedTempl.renderAsAdminPage {
			if r.views["admin"].rawLayouts[parsedTempl.baseLayout] == "" {
				return nil, ErrNotExistsLayout
			}

			newTemplate, err = newTemplate.New("layout").Parse(r.views["admin"].rawLayouts[parsedTempl.contextLayout])
			if err != nil {
				return nil, fmt.Errorf("%w: could not parse admin layout: %v", ErrRenderFailed, err) //nolint:errorlint,lll // prevent err in api
			}
		}
	}

	page, contextPageExists := r.views[parsedTempl.context].rawPages[parsedTempl.page]

	if !contextPageExists {
		p, sharedPageExists := r.views[SharedViews].rawPages[parsedTempl.page]
		if !sharedPageExists {
			return nil, ErrNotExistsPage
		}

		page = p
	}

	newTemplate, err = newTemplate.New("content").Parse(page)
	if err != nil {
		return nil, fmt.Errorf("%w: could not parse page: %v", ErrRenderFailed, err) //nolint:errorlint // prevent err in api
	}

	return newTemplate, nil
}

func prepareViewTemplates(ctx context.Context, logger alog.Logger, viewFS fs.FS, funcMap template.FuncMap, isContext bool) (viewTemplates, error) {
	components, err := fs.Glob(viewFS, "components/*.html")
	if err != nil {
		return viewTemplates{}, fmt.Errorf("could not get components from fs: %v", err) //nolint:errorlint,lll // prevent err in api
	}

	componentTemplates := template.New("<empty>").Funcs(sprig.FuncMap()).Funcs(funcMap)

	for _, c := range components {
		file, err := readFile(viewFS, c) //nolint:govet // govet is too pedantic for shadowing errors
		if err != nil {
			return viewTemplates{}, fmt.Errorf("could not read component file: %s: %v", file, err) //nolint:errorlint,lll // prevent err in api
		}

		name := componentName(c)

		_, err = componentTemplates.New(name).Parse(file)
		if err != nil {
			return viewTemplates{}, fmt.Errorf("could not parse component: %s: %v", file, err) //nolint:errorlint,lll // prevent err in api
		}
	}

	logger.LogAttrs(ctx, alog.LevelDebug,
		"loaded components",
		slog.Int("component_count", len(componentTemplates.Templates())),
		slog.Any("component_templates", templateNames(componentTemplates)),
	)

	pages, err := fs.Glob(viewFS, "pages/*.html")
	if err != nil {
		return viewTemplates{}, fmt.Errorf("could not get pages from fs: %v", err) //nolint:errorlint,lll // prevent err in api
	}

	rawPages := make(map[string]string)

	for _, page := range pages {
		file, err := readFile(viewFS, page) //nolint:govet // govet is too pedantic for shadowing errors
		if err != nil {
			return viewTemplates{}, fmt.Errorf("could not read page file: %s: %v", file, err) //nolint:errorlint,lll // prevent err in api
		}

		pn := pageName(page)
		rawPages[pn] = file
	}

	logger.LogAttrs(ctx, alog.LevelDebug,
		"loaded pages",
		//slog.Int("page_count", len(pageTemplates)),
		slog.Int("page_count", len(rawPages)),
		slog.Any("page_templates", rawTemplateNames(rawPages)),
	)

	layouts, err := fs.Glob(viewFS, "*.html")
	if err != nil {
		return viewTemplates{}, fmt.Errorf("could not get layouts from fs: %v", err) //nolint:errorlint,lll // prevent err in api
	}

	var defaultLayout string

	rawLayouts := make(map[string]string)

	for _, l := range layouts {
		file, err := readFile(viewFS, l)
		if err != nil {
			// todo rename error from layout to base
			return viewTemplates{}, fmt.Errorf("could not read layout file: %s: %v", file, err) //nolint:errorlint,lll // prevent err in api
		}

		ln := baseName(l)
		if isContext {
			ln = layoutName(l)
		}

		rawLayouts[ln] = file

		if ln == "default" {
			defaultLayout = "default"
		}
	}

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
	names := make([]string, n)

	for i, t := range templates.Templates() {
		names[i] = t.Name()
	}

	return names
}

func baseName(baseName string) string {
	return strings.TrimSuffix(baseName, ".base.html")
}

func layoutName(layoutName string) string {
	return strings.TrimSuffix(layoutName, ".layout.html")
}

func componentName(componentName string) string {
	return strings.TrimSuffix(
		strings.TrimPrefix(componentName, "components/"),
		".html",
	)
}

func pageName(pageName string) string {
	return strings.TrimSuffix(
		strings.TrimPrefix(pageName, "pages/"),
		".html",
	)
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

type parsedTemplate struct {
	context           string
	baseLayout        string
	contextLayout     string
	page              string
	fragment          string
	renderAsAdminPage bool
	isComponent       bool
}

func (t parsedTemplate) key() string {
	if t.isComponent {
		return fmt.Sprintf("%s/%s", t.context, t.fragment)
	}

	return fmt.Sprintf("%s/%s%s%s%s%s", t.context, t.baseLayout, templateSeparator, t.contextLayout, templateSeparator, t.page)
}

func (t parsedTemplate) templateName() string {
	if t.fragment != "" {
		return t.fragment
	}

	return t.key()
}

func parseTemplateName(name string) (parsedTemplate, error) {
	const ( // todo combine with templateSeparator and fragmentSeparator
		maxCompositionSegments = 3 // how many segments after separated by the separator
		maxFragmentSegments    = 2
	)

	elem := strings.Split(name, templateSeparator)
	length := len(elem)

	if length > maxCompositionSegments { // invalid pattern
		return parsedTemplate{}, fmt.Errorf("%w", ErrRenderFailed) // todo make error more speaking, by adding context
	}

	var (
		layout        string
		contextLayout string
		page          string
		fragment      string
		isComponent   bool
	)

	if length == 1 {
		page = strings.TrimSpace(elem[0])
	}

	if length == 2 {
		contextLayout = strings.TrimSpace(elem[0])
		page = strings.TrimSpace(elem[1])
	}

	if length == 3 {
		layout = strings.TrimSpace(elem[0])
		contextLayout = strings.TrimSpace(elem[1])
		page = strings.TrimSpace(elem[2])
	}

	fragments := strings.Split(page, fragmentSeparator)
	if len(fragments) > maxFragmentSegments { // invalid pattern
		return parsedTemplate{}, fmt.Errorf("%w", ErrRenderFailed)
	}

	if len(fragments) == 2 {
		page = strings.TrimSpace(fragments[0])
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
	if isInvalid(layout) || isInvalid(contextLayout) || isInvalid(page) || isInvalid(fragment) {
		return parsedTemplate{}, fmt.Errorf("%w", ErrRenderFailed)
	}

	return parsedTemplate{
		baseLayout:    layout,
		contextLayout: contextLayout,
		page:          page,
		fragment:      fragment,
		isComponent:   isComponent,
	}, nil
}

func (r *Renderer) AddContext(name string, viewFS fs.FS) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" {
		return fmt.Errorf("%w: set a name", ErrContextNotAdded)
	}

	if viewFS == nil {
		return fmt.Errorf("%w: no view files", ErrContextNotAdded)
	}

	if _, exists := r.views[name]; exists {
		return fmt.Errorf("%w: already added", ErrContextNotAdded)
	}

	// todo clean, not sure about original meaning of comment. BUT now the bool flag looks like a smell
	view, err := prepareViewTemplates(context.Background(), r.logger, viewFS, r.funcMap, true)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	r.views[name] = view
	tmp := r.views[name]

	cc, err := r.views[SharedViews].components.Clone()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	for _, t := range tmp.components.Templates() {
		c, _ := cc.AddParseTree(t.Name(), t.Tree)
		cc = c
	}

	tmp.components = cc
	r.views[name] = tmp

	return nil
}