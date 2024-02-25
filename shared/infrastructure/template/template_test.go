// Use white box testing, to make it easier to assert on the inner workings of partially loaded and cached templates.
// If a white box test case fails, consider just deleting it over fixing it to prevent coupling to the implementation.
//
//nolint:testpackage
package template

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/go-arrower/arrower/alog"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/go-arrower/skeleton/shared/infrastructure/template/testdata"
)

func TestNewRenderer(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		r, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.EmptyFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
	})

	t.Run("fail on missing files", func(t *testing.T) {
		t.Parallel()

		r, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), nil, false)
		assert.ErrorIs(t, err, ErrInvalidFS)
		assert.Nil(t, r)
	})

	// white box test. if it fails, feel free to delete it
	t.Run("initialise new", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.TemplateFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		assert.Equal(t, 0, renderer.totalCachedTemplates())

		// assert component templates from testdata.TemplateFiles is loaded
		// contains always an empty component, so expected 1 + 2 => 3
		assert.Len(t, renderer.viewsForContext("").components.Templates(), 3)

		// assert pages are found and extracted
		assert.Len(t, renderer.viewsForContext("").rawPages, 3)
		// if the file is called p0.html, the template is called p0
		assert.NotEmpty(t, renderer.viewsForContext("").rawPages["p0"])
		assert.NotEmpty(t, renderer.viewsForContext("").rawPages["p1"])
		assert.NotEmpty(t, renderer.viewsForContext("").rawPages["p2"])
		assert.Empty(t, renderer.viewsForContext("").rawPages["non-existent"])

		//// assert the page has itself and all dependencies loaded as a template
		//assert.Len(t, renderer.templates["p0"].Templates(), 4, "expect: <empty>, c0, c1, p0")
		//assert.Len(t, renderer.templates["p1"].Templates(), 4)
		//assert.Len(t, renderer.templates["p2"].Templates(), 6, "expect: <empty>, components, fragments, and itself as page")
		//
		//// assert template is cached
		//// if the file is called global.layout.html, the template is called global
		//assert.Len(t, renderer.rawLayouts, 1)
		//assert.NotEmpty(t, renderer.rawLayouts["global"])
		//assert.Empty(t, renderer.rawLayouts["non-existent"])
	})

	// white box test. if it fails, feel free to delete it
	t.Run("fs with no files", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.EmptyFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		assert.Len(t, renderer.viewsForContext("").components.Templates(), 1)
		assert.Equal(t, 0, renderer.totalCachedTemplates())
		//assert.Len(t, renderer.rawPages, 0)
		//assert.Len(t, renderer.rawLayouts, 0)
	})
}

func TestRenderer_Render(t *testing.T) {
	t.Parallel()

	t.Run("components", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.TemplateFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		t.Run("component only", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "#c0", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			assert.Equal(t, testdata.C0Content, buf.String())
		})

		t.Run("non existing component", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "#non-existing", nil, testdata.NewEchoContext(t))

			assert.ErrorIs(t, err, ErrNotExistsComponent)
			assert.Empty(t, buf.String())
		})
	})

	t.Run("pages", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.TemplateFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		t.Run("raw page without layout", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "p0", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.P0Content)
			assert.NotContains(t, buf.String(), testdata.LOtherContent)
		})

		t.Run("non existing page", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "non-existing", nil, testdata.NewEchoContext(t))
			assert.ErrorIs(t, err, ErrTemplateNotExists)
			assert.Empty(t, buf.String())
		})

		t.Run("page with components", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "p1", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.P1Content)
			assert.Contains(t, buf.String(), testdata.C0Content)
			assert.NotContains(t, buf.String(), testdata.C1Content)
		})
	})

	t.Run("page fragments", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.TemplateFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		t.Run("whole page", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "p2", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.P2Content)
			assert.Contains(t, buf.String(), testdata.F0Content)
			assert.Contains(t, buf.String(), testdata.F1Content)
		})

		t.Run("fragment 0", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "p2#f0", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			assert.Equal(t, buf.String(), testdata.F0Content)
		})

		t.Run("fragment 1", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "p2#f1", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			assert.Equal(t, buf.String(), testdata.F1Content)
		})

		t.Run("non existing fragment", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "p1#f1", nil, testdata.NewEchoContext(t))
			assert.ErrorIs(t, err, ErrNotExistsFragment)
			assert.Empty(t, buf.String())
		})
	})

	t.Run("layouts", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.LayoutsPagesAndComponents, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		t.Run("page with different layouts", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "global=>p0", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			t.Log(buf.String())
			assert.Contains(t, buf.String(), "globalLayout")
			assert.Contains(t, buf.String(), testdata.P0Content)
			assert.Contains(t, buf.String(), testdata.C0Content)
			assert.NotContains(t, buf.String(), testdata.C1Content)

			buf.Reset()
			err = renderer.Render(buf, "other=>p0", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.LOtherContent)
			assert.NotContains(t, buf.String(), testdata.LDefaultContent)
			assert.Contains(t, buf.String(), testdata.P0Content)
			assert.Contains(t, buf.String(), testdata.C0Content)
			assert.NotContains(t, buf.String(), testdata.C1Content)
		})

		t.Run("access layout that does not exist", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "nonExisting=>p0", nil, testdata.NewEchoContext(t))
			assert.ErrorIs(t, err, ErrNotExistsLayout)

			assert.Empty(t, buf.String())
		})

		t.Run("rely on default layout when rendering page", func(t *testing.T) {
			t.Parallel()

			renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.MultipleLayoutsWithDefaultLayout, false)
			assert.NoError(t, err)
			assert.NotNil(t, renderer)

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "p0", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.LDefaultContent)
			assert.Contains(t, buf.String(), testdata.P0Content)

			// change default layout and render same page again
			//err = renderer.SetDefaultLayout("other")
			//assert.NoError(t, err)
			//buf.Reset()
			//
			//err = renderer.Render(buf, "p0", nil, testdata.NewEchoContext(t))
			//assert.NoError(t, err)
			//
			//assert.Contains(t, buf.String(), testdata.LOtherContent)
			//assert.Contains(t, buf.String(), testdata.P0Content)
		})

		t.Run("explicitly name default layout anyway", func(t *testing.T) {
			t.Parallel()

			renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.MultipleLayoutsWithDefaultLayout, false)
			assert.NoError(t, err)
			assert.NotNil(t, renderer)

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "default=>p0", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.LDefaultContent)
			assert.Contains(t, buf.String(), testdata.P0Content)
		})
	})

	// white box test. if it fails, feel free to delete it
	t.Run("multiple pages and increase template cache", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.LayoutsPagesAndComponents, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		// assert only the pure pages are loaded, because others are not cached yet
		assert.Equal(t, 0, renderer.totalCachedTemplates())

		buf := &bytes.Buffer{}
		err = renderer.Render(buf, "global=>p0", nil, testdata.NewEchoContext(t))
		assert.NoError(t, err)
		assert.Equal(t, 1, renderer.totalCachedTemplates())

		assert.Contains(t, buf.String(), "globalLayout")
		assert.Contains(t, buf.String(), testdata.P0Content)
		assert.Contains(t, buf.String(), testdata.C0Content)
		assert.NotContains(t, buf.String(), testdata.C1Content)

		buf.Reset()
		err = renderer.Render(buf, "global=>p1", nil, testdata.NewEchoContext(t))
		assert.NoError(t, err)
		assert.Equal(t, 2, renderer.totalCachedTemplates())

		assert.Contains(t, buf.String(), "globalLayout")
		assert.Contains(t, buf.String(), testdata.P1Content)
		assert.NotContains(t, buf.String(), testdata.P0Content)
		assert.NotContains(t, buf.String(), testdata.C0Content)
		assert.NotContains(t, buf.String(), testdata.C1Content)

		buf.Reset()
		err = renderer.Render(buf, "global=>p1", nil, testdata.NewEchoContext(t))
		assert.NoError(t, err)
		assert.Equal(t, 2, renderer.totalCachedTemplates())
	})

	t.Run("render in parallel", func(t *testing.T) {
		t.Parallel()

		const numPages = 10
		fs, pageNames := testdata.GenRandomPages(numPages)

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), fs, true)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		wg := &sync.WaitGroup{}

		const numPageLoads = 100
		wg.Add(numPageLoads)
		for i := 0; i < numPageLoads; i++ {
			go func() {
				n := rand.Intn(numPages) //nolint:gosec // rand used to simulate a page visit; not for secure code

				page := pageNames[n]

				buf := &bytes.Buffer{}
				err := renderer.Render(buf, page, nil, testdata.NewEchoContext(t))
				assert.NoError(t, err, page)
				assert.Contains(t, buf.String(), page)

				wg.Done()
			}()
		}

		wg.Wait()
	})
}

// white box test. if it fails, feel free to delete it
func TestParsedTemplate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		template string
		parsed   parsedTemplate
		err      error
	}{
		"empty":                    {"", parsedTemplate{}, nil},
		"component":                {"#c", parsedTemplate{fragment: "c", isComponent: true}, nil},
		"page":                     {"p", parsedTemplate{template: "p"}, nil},
		"page with fragment":       {"p#f", parsedTemplate{template: "p", fragment: "f"}, nil},
		"page with empty fragment": {"p#", parsedTemplate{}, ErrRenderFailed},
		"context layout":           {"cl=>p", parsedTemplate{contextLayout: "cl", template: "p"}, nil},
		"full layout":              {"gl =>cl=> p", parsedTemplate{layout: "gl", contextLayout: "cl", template: "p"}, nil},
		"complete template name":   {"gl=>cl=>p #f ", parsedTemplate{layout: "gl", contextLayout: "cl", template: "p", fragment: "f"}, nil},
		"too many separators":      {"=>=>=>", parsedTemplate{}, ErrRenderFailed},
		"too many fragments":       {"p#p#", parsedTemplate{}, ErrRenderFailed},
		"separator after fragment": {"gl=>cl=>p#f=>", parsedTemplate{}, ErrRenderFailed},
		"fragment in layouts":      {"gl#=>cl=>p#f", parsedTemplate{}, ErrRenderFailed},
	}

	for name, tt := range tests {
		tt := tt

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			template, err := parseTemplateName(tt.template)
			assert.ErrorIs(t, err, tt.err)
			assert.Equal(t, tt.parsed, template)
		})
	}
}

func TestRenderer_Layout(t *testing.T) {
	t.Parallel()

	t.Run("no default layout present", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.EmptyFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		assert.Equal(t, "", renderer.layout())
	})

	t.Run("layout ex, but not the default one", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.SingleNonDefaultLayout, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		assert.Equal(t, "", renderer.layout())
	})

	t.Run("multiple layouts but with default", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.MultipleLayoutsWithDefaultLayout, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		assert.Equal(t, "default", renderer.layout())
	})
}

func TestRenderer_SetDefaultLayout(t *testing.T) {
	t.Parallel()

	t.Run("set existing default layout", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.MultipleLayoutsWithDefaultLayout, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		//err = renderer.SetDefaultLayout("other")
		//assert.NoError(t, err)
		//assert.Equal(t, "other", renderer.Layout())
	})

	t.Run("set non existing layout", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.MultipleLayoutsWithDefaultLayout, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		//err = renderer.SetDefaultLayout("non-existing")
		//assert.ErrorIs(t, err, ErrNotExistsLayout)
		//assert.Equal(t, "default", renderer.Layout())
	})
}

func TestRenderer_AddContext(t *testing.T) {
	t.Parallel()

	t.Run("add context", func(t *testing.T) {
		t.Parallel()

		r, _ := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.EmptyFiles, false)

		err := r.AddContext(testdata.ExampleContext, testdata.EmptyFiles)
		assert.NoError(t, err)
	})

	t.Run("context already added", func(t *testing.T) {
		t.Parallel()

		r, _ := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.EmptyFiles, false)

		err := r.AddContext(testdata.ExampleContext, testdata.EmptyFiles)
		assert.NoError(t, err)

		err = r.AddContext(testdata.ExampleContext, testdata.EmptyFiles)
		assert.ErrorIs(t, err, ErrContextNotAdded)
	})

	t.Run("context needs name", func(t *testing.T) {
		t.Parallel()

		r, _ := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.EmptyFiles, false)

		err := r.AddContext("", testdata.EmptyFiles)
		assert.ErrorIs(t, err, ErrContextNotAdded)
	})

	t.Run("no files", func(t *testing.T) {
		t.Parallel()

		r, _ := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.EmptyFiles, false)

		err := r.AddContext(testdata.ExampleContext, nil)
		assert.ErrorIs(t, err, ErrContextNotAdded)
	})
}

func TestRenderer_RenderContext(t *testing.T) {
	t.Parallel()

	/*
		render a shared component (that did not get overwritten by a context component)
		detect context layouts
		context page and sahred page can have same name
		~~context components: coexist or overwrite shared components~~
		context can use shared components (if not overwritten)
		change context layout with template naming pattern
		from Context: render shared page
		render a page with "otherLayout"
		render a context page that includes a default component
	*/

	t.Run("shared", func(t *testing.T) {
		t.Parallel()

		t.Run("component", func(t *testing.T) {
			t.Parallel()

			renderer, _ := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.SharedViews, false)
			err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
			assert.NoError(t, err)

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "#c1", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.C1Content)
		})

		t.Run("page", func(t *testing.T) {
			t.Parallel()

			renderer, _ := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.SharedViews, false)
			err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
			assert.NoError(t, err)

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "shared-p0", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.P0Content)
			assert.Contains(t, buf.String(), testdata.LDefaultContent)
			assert.Contains(t, buf.String(), testdata.LDefaultContextContent)
			assert.NotContains(t, buf.String(), testdata.LContentPlaceholder)
		})
	})

	t.Run("context component", func(t *testing.T) {
		t.Parallel()

		renderer, _ := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.SharedViews, false)
		err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
		assert.NoError(t, err)

		buf := &bytes.Buffer{}
		c := testdata.NewEchoContext(t)
		c.SetPath(fmt.Sprintf("/%s", testdata.ExampleContext))

		err = renderer.Render(buf, "#c0", nil, c)
		assert.NoError(t, err)

		assert.Contains(t, buf.String(), testdata.C0ContextContent, "context component overwrites shared component with same name")
	})

	t.Run("context page", func(t *testing.T) {
		t.Parallel()

		renderer, _ := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.SharedViews, false)
		err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
		assert.NoError(t, err)

		buf := &bytes.Buffer{}
		c := testdata.NewEchoContext(t)
		c.SetPath(fmt.Sprintf("/%s", testdata.ExampleContext))

		err = renderer.Render(buf, "p0", nil, c)
		assert.NoError(t, err)

		assert.Contains(t, buf.String(), testdata.P0ContextContent)
		assert.Contains(t, buf.String(), testdata.C0ContextContent)
		assert.NotContains(t, buf.String(), testdata.C0Content)
		assert.Contains(t, buf.String(), testdata.LDefaultContent)
		assert.Contains(t, buf.String(), testdata.LDefaultContextContent)
		assert.NotContains(t, buf.String(), testdata.LContentPlaceholder)
	})

	t.Run("context page fragment", func(t *testing.T) {
		t.Parallel()

		renderer, _ := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.SharedViews, false)
		err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
		assert.NoError(t, err)

		buf := &bytes.Buffer{}
		c := testdata.NewEchoContext(t)
		c.SetPath(fmt.Sprintf("/%s", testdata.ExampleContext))

		err = renderer.Render(buf, "p1#f", nil, c)
		assert.NoError(t, err)

		assert.Contains(t, buf.String(), "fragment")
		assert.NotContains(t, buf.String(), "context p1")
		assert.NotContains(t, buf.String(), testdata.LDefaultContent)
		assert.NotContains(t, buf.String(), testdata.LDefaultContextContent)
	})

	t.Run("context page rendered as admin", func(t *testing.T) {
		t.Parallel()

		renderer, _ := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.SharedViews, false)
		err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
		assert.NoError(t, err)
		err = renderer.AddContext("admin", testdata.ContextAdmin)
		assert.NoError(t, err)

		buf := &bytes.Buffer{}
		c := testdata.NewEchoContext(t)
		c.SetPath(fmt.Sprintf("/admin/%s", testdata.ExampleContext))

		err = renderer.Render(buf, "p0", nil, c)
		assert.NoError(t, err)

		assert.Contains(t, buf.String(), testdata.P0ContextContent)
		//assert.Contains(t, buf.String(), "context c0") // TODO (?) uncomment, as it should overwrite
		assert.Contains(t, buf.String(), testdata.LDefaultContent)
		assert.Contains(t, buf.String(), "adminLayout")
		assert.NotContains(t, buf.String(), testdata.LDefaultContextContent)
		assert.NotContains(t, buf.String(), testdata.LContentPlaceholder)
		assert.NotContains(t, buf.String(), "adminPlaceholder")
	})

	t.Run("shared page from context", func(t *testing.T) {
		t.Parallel()
		t.Skip() // TODO specs unclear:
		// what should be rendered if a context calls a shared view?
		// what would be example use cases?

		renderer, _ := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.SharedViews, false)
		err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
		assert.NoError(t, err)

		buf := &bytes.Buffer{}
		c := testdata.NewEchoContext(t)
		c.SetPath(fmt.Sprintf("/%s", testdata.ExampleContext))

		err = renderer.Render(buf, "shared-p0", nil, c)
		assert.NoError(t, err)

		//assert.Contains(t, buf.String(), "context p0")
		assert.Contains(t, buf.String(), testdata.P0Content)
		assert.Contains(t, buf.String(), testdata.C0Content)
		assert.Contains(t, buf.String(), "defaultLayout")
		assert.Contains(t, buf.String(), "defaultLayoutContextLayoutPlaceholder", "a shared page does not load a context layout")
		assert.NotContains(t, buf.String(), "defaultLayoutContextContentPlaceholder")
		assert.NotContains(t, buf.String(), "contextLayout")
		assert.NotContains(t, buf.String(), "contextPlaceholder")
	})
}

// TODO test case for hot reload
