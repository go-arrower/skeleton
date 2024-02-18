// Use white box testing, to make it easier to assert on the inner workings of partially loaded and cached templates.
// If a white box test case fails, consider just deleting it over fixing it, to prevent coupling to the implementation.
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
	views2 "github.com/go-arrower/skeleton/shared/views"
)

func TestNewRenderer(t *testing.T) {
	t.Parallel()

	t.Run("construct renderer", func(t *testing.T) {
		t.Parallel()

		r, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), views2.SharedViews, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
	})

	t.Run("fail on missing files", func(t *testing.T) {
		t.Parallel()

		r, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), nil, false)
		assert.Error(t, err)
		assert.Nil(t, r)
	})

	// white box test. if it fails, feel free to delete it
	t.Run("initialise raw renderer", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.TemplateFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		// assert component templates from testdata.TemplateFiles is loaded
		// contains always an empty component, so expected 1 + 2 => 3
		assert.Len(t, renderer.components.Templates(), 3)

		// assert pages are loaded
		assert.Len(t, renderer.templates, 3)
		// if the file is called p0.page.html, the template is called p0
		assert.NotEmpty(t, renderer.templates["p0"])
		assert.NotEmpty(t, renderer.templates["p1"])
		assert.NotEmpty(t, renderer.templates["p2"])
		assert.Empty(t, renderer.templates["non-existent"])

		// assert the page has itself and all dependencies loaded as a template
		assert.Len(t, renderer.templates["p0"].Templates(), 4, "expect: <empty>, c0, c1, p0")
		assert.Len(t, renderer.templates["p1"].Templates(), 4)
		assert.Len(t, renderer.templates["p2"].Templates(), 6, "expect: <empty>, components, fragments, and itself as page")

		// assert template is cached
		// if the file is called global.layout.html, the template is called global
		assert.Len(t, renderer.rawLayouts, 1)
		assert.NotEmpty(t, renderer.rawLayouts["global"])
		assert.Empty(t, renderer.rawLayouts["non-existent"])
	})

	// white box test. if it fails, feel free to delete it
	t.Run("fs with no files", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.EmptyFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		assert.Len(t, renderer.components.Templates(), 1)
		assert.Len(t, renderer.templates, 0)
		assert.Len(t, renderer.rawPages, 0)
		assert.Len(t, renderer.rawLayouts, 0)
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
			err = renderer.Render(buf, "c0", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			assert.Equal(t, testdata.C0Content, buf.String())
		})

		t.Run("non existing component", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "non-existing", nil, testdata.NewEchoContext(t))

			assert.ErrorIs(t, err, ErrTemplateNotExists)
			assert.Empty(t, buf.String())
		})

		t.Run("component and page with same name", func(t *testing.T) {
			t.Parallel()

			renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.ConflictingTemplateFiles, false)
			assert.NoError(t, err)
			assert.NotNil(t, renderer)

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "conflict", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.P0Content)
			assert.NotContains(t, buf.String(), testdata.C0Content)
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

			assert.Equal(t, buf.String(), testdata.P0Content)
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

			assert.Contains(t, buf.String(), testdata.LOtherContent)
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
			err = renderer.SetDefaultLayout("other")
			assert.NoError(t, err)
			buf.Reset()

			err = renderer.Render(buf, "p0", nil, testdata.NewEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.LOtherContent)
			assert.Contains(t, buf.String(), testdata.P0Content)
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
		assert.Len(t, renderer.templates, 2)

		buf := &bytes.Buffer{}
		err = renderer.Render(buf, "global=>p0", nil, testdata.NewEchoContext(t))
		assert.NoError(t, err)
		assert.Len(t, renderer.templates, 3) // previous templates + global=>p0

		assert.Contains(t, buf.String(), testdata.LOtherContent)
		assert.Contains(t, buf.String(), testdata.P0Content)
		assert.Contains(t, buf.String(), testdata.C0Content)
		assert.NotContains(t, buf.String(), testdata.C1Content)

		buf.Reset()
		err = renderer.Render(buf, "global=>p1", nil, testdata.NewEchoContext(t))
		assert.NoError(t, err)
		assert.Len(t, renderer.templates, 4) // previous templates + global=>p1

		assert.Contains(t, buf.String(), testdata.LOtherContent)
		assert.Contains(t, buf.String(), testdata.P1Content)
		assert.NotContains(t, buf.String(), testdata.P0Content)
		assert.NotContains(t, buf.String(), testdata.C0Content)
		assert.NotContains(t, buf.String(), testdata.C1Content)

		buf.Reset()
		err = renderer.Render(buf, "global=>p1", nil, testdata.NewEchoContext(t))
		assert.NoError(t, err)
		assert.Len(t, renderer.templates, 4) // template is cached already, so no change
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
		"component or page":        {"p", parsedTemplate{template: "p"}, nil},
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

		assert.Equal(t, "", renderer.Layout())
	})

	t.Run("layout ex, but not the default one", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.SingleNonDefaultLayout, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		assert.Equal(t, "", renderer.Layout())
	})

	t.Run("multiple layouts but with default", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.MultipleLayoutsWithDefaultLayout, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		assert.Equal(t, "default", renderer.Layout())
	})
}

func TestRenderer_SetDefaultLayout(t *testing.T) {
	t.Parallel()

	t.Run("set existing default layout", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.MultipleLayoutsWithDefaultLayout, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		err = renderer.SetDefaultLayout("other")
		assert.NoError(t, err)
		assert.Equal(t, "other", renderer.Layout())
	})

	t.Run("set non existing layout", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.MultipleLayoutsWithDefaultLayout, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		err = renderer.SetDefaultLayout("non-existing")
		assert.ErrorIs(t, err, ErrNotExistsLayout)
		assert.Equal(t, "default", renderer.Layout())
	})
}

// white box test. if it fails, feel free to delete it
func TestParseLayoutAndPage(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		name           string
		expectedLayout string
		expectedPage   string
	}{
		"empty": {
			"",
			"",
			"",
		},
		"just page": {
			"p0",
			"",
			"p0",
		},
		"layout and page": {
			"l=>p",
			"l",
			"p",
		},
		"trim whitespaces": {
			" l => p ",
			"l",
			"p",
		},
		"layout, sub-layout, and page": {
			"l=>s=>p",
			"l=>s",
			"p",
		},
	}

	for name, tt := range tests {
		tt := tt

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			l, p := parseLayoutAndPage(tt.name)
			assert.Equal(t, tt.expectedLayout, l)
			assert.Equal(t, tt.expectedPage, p)
		})
	}
}

func TestRenderer_AddContext(t *testing.T) {
	t.Parallel()

	t.Run("add context", func(t *testing.T) {
		t.Parallel()

		r, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.EmptyFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)

		err = r.AddContext(testdata.ExampleContext, testdata.EmptyFiles)
		assert.NoError(t, err)
	})

	// build path automatically, as it is a convention (?)
	// err on nil files
	// err on empty context
	// err if context is already loaded
	// assert list of loaded contexts
}

func TestRenderer_RenderContext(t *testing.T) {
	t.Parallel()

	/*
		render shared page as before
		render context page
		render context page with fragment
		detect context layouts
		context page and sahred page can have same name
		context components: coexist or overwrite shared components
	*/

	t.Run("page", func(t *testing.T) {
		t.Parallel()

		renderer, _ := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.SharedViews, true)
		err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
		assert.NoError(t, err)

		buf := &bytes.Buffer{}
		err = renderer.Render(buf, "p0", nil, testdata.NewEchoContext(t))
		assert.NoError(t, err)

		assert.Contains(t, buf.String(), testdata.P0Content)
		assert.Contains(t, buf.String(), "defaultLayout")
		assert.Contains(t, buf.String(), "defaultLayoutContextLayoutPlaceholder")
		assert.NotContains(t, buf.String(), "defaultLayoutContextContentPlaceholder")

		buf.Reset()
		c := testdata.NewEchoContext(t)
		c.SetPath(fmt.Sprintf("/%s", testdata.ExampleContext))
		err = renderer.Render(buf, "p0", nil, c)
		assert.NoError(t, err)

		assert.Contains(t, buf.String(), "context p0")
		// todo assert on context component => overwrite
		assert.Contains(t, buf.String(), "defaultLayout")
		assert.Contains(t, buf.String(), "contextLayout")
		assert.NotContains(t, buf.String(), "defaultLayoutContextLayoutPlaceholder")
		assert.NotContains(t, buf.String(), "defaultLayoutContextContentPlaceholder")
		assert.NotContains(t, buf.String(), "contextPlaceholder")
	})
}
