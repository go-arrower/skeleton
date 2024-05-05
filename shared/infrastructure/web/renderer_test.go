package web_test

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"math/rand"
	"sync"
	"testing"

	"github.com/go-arrower/arrower/alog"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/go-arrower/skeleton/shared/infrastructure/web"
	"github.com/go-arrower/skeleton/shared/infrastructure/web/testdata"
)

var errSomeError = errors.New("some-error")

func TestNewRenderer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("empty fs", func(t *testing.T) {
		t.Parallel()

		r, err := web.NewRenderer(alog.NewNoopLogger(), noop.NewTracerProvider(), testdata.FilesEmpty, template.FuncMap{}, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
	})

	t.Run("missing fs", func(t *testing.T) {
		t.Parallel()

		r, err := web.NewRenderer(alog.NewNoopLogger(), noop.NewTracerProvider(), nil, template.FuncMap{}, false)
		assert.ErrorIs(t, err, web.ErrCreateRendererFailed)
		assert.Nil(t, r)
	})

	t.Run("custom func map", func(t *testing.T) {
		t.Parallel()

		r, err := web.NewRenderer(alog.NewNoopLogger(), noop.NewTracerProvider(), testdata.FilesSharedViewsWithCustomFuncs(), template.FuncMap{
			"customFunc": func() string { return "hello custom func" },
		}, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)

		// Don't execute the following two tests in parallel.
		// Their order prevents a regression bug where they either pass or fail:
		// If the component is rendered first, the page did fail, if the page renders first it did succeed.
		// This is because the page does depend on the components,
		// and if (component) templates are executed at least ones already, Go template returns an error.
		// The desired goal is that the developer can freely use any combination and order of components and pages.
		// todo could be it's own regression test, as now it is mixed with the custom funcs

		t.Run("in component", func(t *testing.T) {
			buf := &bytes.Buffer{}
			err = r.Render(ctx, buf, web.SharedViews, "#use-func-map", nil)

			assert.NoError(t, err)
			assert.Equal(t, "hello custom func", buf.String())
		})

		t.Run("in page", func(t *testing.T) {
			buf := &bytes.Buffer{}
			err = r.Render(ctx, buf, web.SharedViews, "use-func-map", nil)

			assert.NoError(t, err)
			assert.Contains(t, buf.String(), "hello custom func")
			assert.Contains(t, buf.String(), "hello")
		})
	})
}

func TestRenderer_Render(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("shared views", func(t *testing.T) {
		t.Parallel()

		t.Run("components", func(t *testing.T) {
			t.Parallel()

			renderer, err := web.NewRenderer(alog.NewNoopLogger(), noop.NewTracerProvider(), testdata.FilesSharedViews(), template.FuncMap{}, false)
			assert.NoError(t, err)

			t.Run("component", func(t *testing.T) {
				t.Parallel()

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "#c0", nil)
				assert.NoError(t, err)

				assert.Equal(t, testdata.C0Content, buf.String())
			})

			t.Run("non existing component", func(t *testing.T) {
				t.Parallel()

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "#non-existing", nil)

				assert.ErrorIs(t, err, web.ErrRenderFailed)
				assert.ErrorIs(t, err, web.ErrNotExistsComponent) // FIXME this test fails sometimes, very very rarely (1/30 or fewer)
				assert.Empty(t, buf.String())
			})
		})

		t.Run("pages", func(t *testing.T) {
			t.Parallel()

			renderer, err := web.NewRenderer(alog.NewNoopLogger(), noop.NewTracerProvider(), testdata.FilesSharedViews(), template.FuncMap{}, false)
			assert.NoError(t, err)

			t.Run("page without default base layout", func(t *testing.T) {
				t.Parallel()

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "p0", nil)
				assert.NoError(t, err)

				assert.Contains(t, buf.String(), testdata.P0Content)
				assert.NotContains(t, buf.String(), testdata.BaseLayoutContent)
				assert.NotContains(t, buf.String(), testdata.BaseLayoutPagePlaceholder)
				assert.NotContains(t, buf.String(), testdata.BaseLayoutContentPlaceholder)
			})

			t.Run("page without default base but components", func(t *testing.T) {
				t.Parallel()

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "p1", nil)
				assert.NoError(t, err)

				assert.Contains(t, buf.String(), testdata.P1Content)
				assert.Contains(t, buf.String(), testdata.C0Content)
				assert.NotContains(t, buf.String(), testdata.BaseLayoutContent)
				assert.NotContains(t, buf.String(), testdata.BaseLayoutPagePlaceholder)
				assert.NotContains(t, buf.String(), testdata.P0Content)
				assert.NotContains(t, buf.String(), testdata.C1Content)
				assert.NotContains(t, buf.String(), testdata.BaseLayoutContentPlaceholder)
			})

			t.Run("non existing page", func(t *testing.T) {
				t.Parallel()

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "non-existing", nil)

				assert.ErrorIs(t, err, web.ErrRenderFailed)
				assert.ErrorIs(t, err, web.ErrNotExistsPage)
				assert.Empty(t, buf.String())
			})

			t.Run("page without base layout", func(t *testing.T) {
				t.Parallel()

				renderer, err := web.NewRenderer(alog.NewNoopLogger(), noop.NewTracerProvider(), testdata.FilesSharedViewsWithoutBase(), template.FuncMap{}, false)
				assert.NoError(t, err)

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "p1", nil)
				assert.NoError(t, err)

				assert.Contains(t, buf.String(), testdata.P1Content)
				assert.Contains(t, buf.String(), testdata.C0Content)
				assert.NotContains(t, buf.String(), testdata.BaseLayoutContent)
				assert.NotContains(t, buf.String(), testdata.BaseLayoutPagePlaceholder)
				assert.NotContains(t, buf.String(), testdata.BaseLayoutContentPlaceholder)
			})
		})

		t.Run("page fragments", func(t *testing.T) {
			t.Parallel()

			renderer, err := web.NewRenderer(alog.NewNoopLogger(), noop.NewTracerProvider(), testdata.FilesSharedViews(), template.FuncMap{}, false)
			assert.NoError(t, err)

			t.Run("whole page", func(t *testing.T) {
				t.Parallel()

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "p2", nil)
				assert.NoError(t, err)

				assert.Contains(t, buf.String(), testdata.P2Content)
				assert.Contains(t, buf.String(), testdata.F0Content)
				assert.Contains(t, buf.String(), testdata.F1Content)
			})

			t.Run("fragment 0", func(t *testing.T) {
				t.Parallel()

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "p2#f0", nil)
				assert.NoError(t, err)

				assert.Equal(t, testdata.F0Content, buf.String())
			})

			t.Run("fragment 1", func(t *testing.T) {
				t.Parallel()

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "p2#f1", nil)
				assert.NoError(t, err)

				assert.Equal(t, testdata.F1Content, buf.String())
			})

			t.Run("non existing fragment", func(t *testing.T) {
				t.Parallel()

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "p1#f1", nil)

				assert.ErrorIs(t, err, web.ErrRenderFailed)
				assert.ErrorIs(t, err, web.ErrNotExistsFragment)
				assert.Empty(t, buf.String())
			})
		})

		t.Run("bases", func(t *testing.T) {
			t.Parallel()

			renderer, err := web.NewRenderer(alog.NewNoopLogger(), noop.NewTracerProvider(), testdata.FilesSharedViewsWithMultiBase(), template.FuncMap{}, false)
			assert.NoError(t, err)

			t.Run("page with different base layouts", func(t *testing.T) {
				t.Parallel()

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "global=>p1", nil)
				assert.NoError(t, err)

				t.Log(buf.String())
				assert.Contains(t, buf.String(), "globalLayout") // todo
				assert.Contains(t, buf.String(), testdata.P1Content)
				assert.Contains(t, buf.String(), testdata.C0Content)
				assert.NotContains(t, buf.String(), testdata.C1Content)

				buf.Reset()
				err = renderer.Render(ctx, buf, web.SharedViews, "other=>p1", nil)
				assert.NoError(t, err)

				assert.Contains(t, buf.String(), "otherLayout") // todo
				assert.NotContains(t, buf.String(), testdata.BaseDefaultLayoutContent)
				assert.Contains(t, buf.String(), testdata.P1Content)
				assert.Contains(t, buf.String(), testdata.C0Content)
				assert.NotContains(t, buf.String(), testdata.C1Content)
			})

			t.Run("access base that does not exist", func(t *testing.T) {
				t.Parallel()

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "nonExisting=>p0", nil)

				assert.ErrorIs(t, err, web.ErrRenderFailed)
				assert.ErrorIs(t, err, web.ErrNotExistsLayout)
				assert.Empty(t, buf.String())
			})

			t.Run("automatic default base", func(t *testing.T) {
				t.Parallel()

				renderer, err := web.NewRenderer(alog.NewNoopLogger(), noop.NewTracerProvider(), testdata.FilesSharedViewsWithDefaultBase(), template.FuncMap{}, false)
				assert.NoError(t, err)

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "p0", nil)
				assert.NoError(t, err)

				assert.Contains(t, buf.String(), testdata.BaseDefaultLayoutContent)
				assert.Contains(t, buf.String(), testdata.P0Content)
			})

			t.Run("explicit default base", func(t *testing.T) {
				t.Parallel()

				renderer, err := web.NewRenderer(alog.NewNoopLogger(), noop.NewTracerProvider(), testdata.FilesSharedViewsWithDefaultBase(), template.FuncMap{}, false)
				assert.NoError(t, err)

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "default=>p0", nil)
				assert.NoError(t, err)

				assert.Contains(t, buf.String(), testdata.BaseDefaultLayoutContent)
				assert.Contains(t, buf.String(), testdata.P0Content)
			})

			t.Run("base without context layout", func(t *testing.T) {
				t.Parallel()
				t.Skip() // FIXME this test is new and needs checking

				renderer, err := web.NewRenderer(alog.NewNoopLogger(), noop.NewTracerProvider(), testdata.FilesSharedViewsWithDefaultBaseWithoutLayout(), template.FuncMap{}, false)
				assert.NoError(t, err)

				buf := &bytes.Buffer{}
				err = renderer.Render(ctx, buf, web.SharedViews, "default=>p0", nil)
				assert.NoError(t, err)

				assert.Contains(t, buf.String(), testdata.BaseDefaultLayoutContent)
				assert.Contains(t, buf.String(), testdata.P0Content)
			})

			// todo
			// base without context layout block in FilesSharedViewsWithDefaultBase
			// base without content block
		})

		t.Run("render in parallel", func(t *testing.T) {
			t.Parallel()

			const numPages = 10
			fs, pageNames := testdata.GenRandomPages(numPages)

			renderer, err := web.NewRenderer(alog.NewNoopLogger(), noop.NewTracerProvider(), fs, template.FuncMap{}, true)
			assert.NoError(t, err)

			wg := &sync.WaitGroup{}

			const numPageLoads = 1000

			wg.Add(numPageLoads)
			for i := 0; i < numPageLoads; i++ {
				go func() {
					n := rand.Intn(numPages) //nolint:gosec // rand used to simulate a page visit; not for secure code

					page := pageNames[n]
					buf := &bytes.Buffer{}

					err := renderer.Render(ctx, buf, web.SharedViews, page, nil)
					assert.NoError(t, err, page)
					assert.Contains(t, buf.String(), page)

					wg.Done()
				}()
			}

			wg.Wait()
		})
	})

	t.Run("context views", func(t *testing.T) {
		t.Parallel()

		t.Run("shared component", func(t *testing.T) {
			t.Parallel()

			renderer, _ := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesSharedViews(), template.FuncMap{}, false)
			err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
			assert.NoError(t, err)

			buf := &bytes.Buffer{}
			err = renderer.Render(ctx, buf, testdata.ExampleContext, "#c1", nil)
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.C1Content)
		})

		t.Run("overwrite shared component", func(t *testing.T) {
			t.Parallel()

			renderer, _ := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesSharedViews(), template.FuncMap{}, false)
			err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
			assert.NoError(t, err)

			buf := &bytes.Buffer{}
			err = renderer.Render(ctx, buf, testdata.ExampleContext, "#c0", nil)
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.C0ContextContent, "context component overwrites shared component with same name")
		})

		t.Run("context layout and page", func(t *testing.T) {
			t.Parallel()

			renderer, _ := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesSharedViewsWithDefaultBase(), template.FuncMap{}, false)
			err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
			assert.NoError(t, err)

			buf := &bytes.Buffer{}
			err = renderer.Render(ctx, buf, testdata.ExampleContext, "p0", nil)
			assert.NoError(t, err)
			t.Log(buf.String())
			assert.Contains(t, buf.String(), testdata.P0ContextContent)
			assert.Contains(t, buf.String(), testdata.C0ContextContent)
			assert.Contains(t, buf.String(), testdata.BaseDefaultLayoutContent)
			assert.Contains(t, buf.String(), testdata.ContextLayoutPagePlaceholder)
			assert.NotContains(t, buf.String(), testdata.ContextLayoutContentPlaceholder)
			assert.NotContains(t, buf.String(), testdata.BaseLayoutPagePlaceholder)
			assert.NotContains(t, buf.String(), testdata.BaseLayoutContentPlaceholder)
			assert.NotContains(t, buf.String(), testdata.C0Content, "c0 is overwritten")
		})

		// todo page with different context layout
		// TODO test case for hot reload
		// TODO test case for a context that does not have a layout (use base layout block as fallback)
		/*
			detect context layouts
			change context layout with template naming pattern
			render a page with "otherLayout"
			render a context page that includes a default component
			context has template but no global template => create placeholder global to make it work anyway
			change "defaultContextLayout" in the context layout to be different from the global layout, so it can be asserted on which one is used
		*/

		t.Run("context page fragment", func(t *testing.T) {
			t.Parallel()

			renderer, _ := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesSharedViewsWithDefaultBase(), template.FuncMap{}, false)
			err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
			assert.NoError(t, err)

			buf := &bytes.Buffer{}
			err = renderer.Render(ctx, buf, testdata.ExampleContext, "p1#f", nil)
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), "fragment")
			assert.NotContains(t, buf.String(), testdata.P1ContextContent, "should not contain page content")
			assert.NotContains(t, buf.String(), testdata.ContextLayoutContentPlaceholder, "should not contain content placeholder")
			assert.NotContains(t, buf.String(), testdata.BaseDefaultLayoutContent, "should not contain any base layout")
		})

		t.Run("context page rendered as admin", func(t *testing.T) {
			t.Parallel()

			renderer, _ := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesSharedViewsWithDefaultBase(), template.FuncMap{}, false)
			err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
			assert.NoError(t, err)
			err = renderer.AddContext("admin", testdata.ContextAdmin)
			assert.NoError(t, err)

			buf := &bytes.Buffer{}
			err = renderer.Render(ctx, buf, "/admin/"+testdata.ExampleContext, "p0", nil)
			assert.NoError(t, err)

			// todo recheck all assertions
			assert.Contains(t, buf.String(), testdata.P0ContextContent)
			assert.Contains(t, buf.String(), testdata.BaseDefaultLayoutContent)
			assert.Contains(t, buf.String(), "adminLayout")
			// assert.NotContains(t, buf.String(), testdata.LDefaultContextContent)
			assert.NotContains(t, buf.String(), testdata.BaseLayoutContentPlaceholder)
			assert.NotContains(t, buf.String(), "adminPlaceholder")
		})

		t.Run("shared page", func(t *testing.T) {
			t.Parallel()

			renderer, _ := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesSharedViewsWithDefaultBase(), template.FuncMap{}, false)
			err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
			assert.NoError(t, err)

			buf := &bytes.Buffer{}
			err = renderer.Render(ctx, buf, testdata.ExampleContext, "shared", nil)
			assert.NoError(t, err)

			// todo recheck all assertions
			assert.Contains(t, buf.String(), testdata.P0Content)
			assert.Contains(t, buf.String(), testdata.C0ContextContent, "shared component got overwritten")
			assert.Contains(t, buf.String(), testdata.BaseDefaultLayoutContent)
			// assert.Contains(t, buf.String(), testdata.LDefaultContextContent)
			assert.NotContains(t, buf.String(), testdata.BaseLayoutContentPlaceholder)
		})

		t.Run("overwrite shared page", func(t *testing.T) {
			t.Parallel()

			renderer, _ := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesSharedViewsWithDefaultBase(), template.FuncMap{}, false)
			err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
			assert.NoError(t, err)

			buf := &bytes.Buffer{}
			err = renderer.Render(ctx, buf, testdata.ExampleContext, "conflict-page", nil)
			assert.NoError(t, err)

			// todo recheck all assertions
			assert.Contains(t, buf.String(), `context conflict`)
			assert.NotContains(t, buf.String(), testdata.P0Content)
			assert.Contains(t, buf.String(), testdata.BaseDefaultLayoutContent)
			// assert.Contains(t, buf.String(), testdata.LDefaultContextContent)
			assert.NotContains(t, buf.String(), testdata.BaseLayoutContentPlaceholder)
		})
	})

	t.Run("hot reload", func(t *testing.T) {
		t.Parallel()

		t.Run("context layout and page", func(t *testing.T) {
			t.Parallel()

			renderer, _ := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesSharedViewsWithDefaultBase(), template.FuncMap{}, true)
			err := renderer.AddContext(testdata.ExampleContext, testdata.ContextViews)
			assert.NoError(t, err)

			buf := &bytes.Buffer{}
			err = renderer.Render(ctx, buf, testdata.ExampleContext, "p0", nil)
			assert.NoError(t, err)
			t.Log(buf.String())
			assert.Contains(t, buf.String(), testdata.P0ContextContent)
			assert.Contains(t, buf.String(), testdata.C0ContextContent)
			assert.Contains(t, buf.String(), testdata.BaseDefaultLayoutContent)
			assert.Contains(t, buf.String(), testdata.ContextLayoutPagePlaceholder)
			assert.NotContains(t, buf.String(), testdata.ContextLayoutContentPlaceholder)
			assert.NotContains(t, buf.String(), testdata.BaseLayoutPagePlaceholder)
			assert.NotContains(t, buf.String(), testdata.BaseLayoutContentPlaceholder)
			assert.NotContains(t, buf.String(), testdata.C0Content, "c0 is overwritten")
		})
	})
}

func TestRenderer_AddContext(t *testing.T) {
	t.Parallel()

	t.Run("add context", func(t *testing.T) {
		t.Parallel()

		r, _ := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesEmpty, template.FuncMap{}, false)

		err := r.AddContext(testdata.ExampleContext, testdata.FilesEmpty)
		assert.NoError(t, err)
	})

	t.Run("context already added", func(t *testing.T) {
		t.Parallel()

		r, _ := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesEmpty, template.FuncMap{}, false)

		err := r.AddContext(testdata.ExampleContext, testdata.FilesEmpty)
		assert.NoError(t, err)

		err = r.AddContext(testdata.ExampleContext, testdata.FilesEmpty)
		assert.ErrorIs(t, err, web.ErrContextNotAdded)
	})

	t.Run("context needs name", func(t *testing.T) {
		t.Parallel()

		r, _ := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesEmpty, template.FuncMap{}, false)

		err := r.AddContext("", testdata.FilesEmpty)
		assert.ErrorIs(t, err, web.ErrContextNotAdded)
	})

	t.Run("no files", func(t *testing.T) {
		t.Parallel()

		r, _ := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesEmpty, template.FuncMap{}, false)

		err := r.AddContext(testdata.ExampleContext, nil)
		assert.ErrorIs(t, err, web.ErrContextNotAdded)
	})
}

// TODO cases
// render page with other base
// render with other layout
// check all the non existing cases for base, layout, page, component, fragment (shared & context)

func TestRenderer_AddBaseData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	/*
		todo concurrency & race
	*/

	t.Run("add base data", func(t *testing.T) {
		t.Parallel()

		r, err := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesAddBaseData(), template.FuncMap{}, false)
		assert.NoError(t, err)

		err = r.AddBaseData("default", func(_ context.Context) (map[string]any, error) {
			return map[string]any{
				"baseTitle": "baseTitle",
			}, nil
		})
		assert.NoError(t, err)

		err = r.AddBaseData("", func(_ context.Context) (map[string]any, error) {
			return web.Map{
				"BaseHeader": "baseHeader",
			}, nil
		})
		assert.NoError(t, err)

		buf := &bytes.Buffer{}
		err = r.Render(ctx, buf, web.SharedViews, "p0", nil)
		assert.NoError(t, err)

		assert.Contains(t, buf.String(), "baseTitle")
		assert.Contains(t, buf.String(), "baseHeader")
		assert.Contains(t, buf.String(), testdata.P0Content)
	})

	t.Run("overwrite base data", func(t *testing.T) {
		t.Parallel()

		r, err := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesAddBaseData(), template.FuncMap{}, false)
		assert.NoError(t, err)

		err = r.AddBaseData("default", func(_ context.Context) (map[string]any, error) {
			return map[string]any{
				"baseTitle": "baseTitle 0",
			}, nil
		})
		assert.NoError(t, err)
		err = r.AddBaseData("", func(_ context.Context) (map[string]any, error) {
			return web.Map{
				"baseTitle": "baseTitle 1",
			}, nil
		})
		assert.NoError(t, err)

		t.Run("last base data wins", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = r.Render(ctx, buf, web.SharedViews, "p0", nil)
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), "baseTitle 1")
			assert.NotContains(t, buf.String(), "baseTitle 0")
		})

		t.Run("any map", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = r.Render(ctx, buf, web.SharedViews, "p0", map[string]any{
				"baseTitle": "baseTitle 2",
			})
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), "baseTitle 2")
			assert.NotContains(t, buf.String(), "baseTitle 1")
			assert.NotContains(t, buf.String(), "baseTitle 0")
		})

		t.Run("string map", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = r.Render(ctx, buf, web.SharedViews, "p0", map[string]string{
				"baseTitle": "baseTitle 2",
			})
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), "baseTitle 2")
			assert.NotContains(t, buf.String(), "baseTitle 1")
			assert.NotContains(t, buf.String(), "baseTitle 0")
		})

		type someType struct {
			Name string
		}

		t.Run("struct type", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = r.Render(ctx, buf, web.SharedViews, "p0", someType{Name: "someName"})
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), "someName")
			assert.Contains(t, buf.String(), "baseTitle 1")
			assert.NotContains(t, buf.String(), "baseTitle 0")
		})

		t.Run("slice type", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = r.Render(ctx, buf, web.SharedViews, "p0", []someType{{Name: "someName"}})
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), "<li>someName</li>")
			assert.Contains(t, buf.String(), "baseTitle 1")
			assert.NotContains(t, buf.String(), "baseTitle 0")
		})
	})

	t.Run("data func fails", func(t *testing.T) {
		t.Parallel()

		r, err := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesAddBaseData(), template.FuncMap{}, false)
		assert.NoError(t, err)

		err = r.AddBaseData("default", func(_ context.Context) (map[string]any, error) {
			return nil, errSomeError
		})
		assert.NoError(t, err)

		buf := &bytes.Buffer{}
		err = r.Render(ctx, buf, web.SharedViews, "p0", nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, web.ErrRenderFailed)
		assert.Empty(t, buf.String())
	})

	t.Run("base not exists", func(t *testing.T) {
		t.Parallel()

		r, err := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesAddBaseData(), template.FuncMap{}, false)
		assert.NoError(t, err)

		err = r.AddBaseData("non-existing", func(_ context.Context) (map[string]any, error) {
			return map[string]any{}, nil
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, web.ErrCreateRendererFailed)
	})
}

func TestRenderer_AddLayoutData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	/*
		todo
		concurrency & race
		different base
	*/

	t.Run("layout data", func(t *testing.T) {
		t.Parallel()

		r := testRendererReadyForLayoutData(t)

		err := r.AddLayoutData(testdata.ExampleContext, "default", func(_ context.Context) (map[string]any, error) {
			return map[string]any{
				"layoutTitle": "layoutTitle",
			}, nil
		})
		assert.NoError(t, err)
		err = r.AddLayoutData(testdata.ExampleContext, "", func(_ context.Context) (map[string]any, error) {
			return map[string]any{
				"LayoutHeader": "layoutHeader",
			}, nil
		})
		assert.NoError(t, err)

		// ensure the base and layout data is available in all context template and page combinations.
		tests := map[string]struct {
			templateName string
		}{
			"page":                {"p0"},
			"default layout page": {"default=>p0"},
			"other layout page":   {"other=>p0"},
		}

		for testName, tt := range tests {
			t.Run(testName, func(t *testing.T) {
				t.Parallel()
				buf := &bytes.Buffer{}

				err = r.Render(ctx, buf, testdata.ExampleContext, tt.templateName, map[string]any{
					"baseTitle": "baseTitle 2",
				})
				assert.NoError(t, err)

				assert.Contains(t, buf.String(), "baseTitle 2")
				assert.Contains(t, buf.String(), "layoutTitle")
				assert.Contains(t, buf.String(), "layoutHeader")
			})
		}
	})

	t.Run("context does not exist", func(t *testing.T) {
		t.Parallel()

		r := testRendererReadyForLayoutData(t)

		err := r.AddLayoutData("not-exists", "default", func(_ context.Context) (map[string]any, error) {
			return map[string]any{}, nil
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, web.ErrCreateRendererFailed)
	})

	t.Run("layout does not exist", func(t *testing.T) {
		t.Parallel()

		r := testRendererReadyForLayoutData(t)

		err := r.AddLayoutData(testdata.ExampleContext, "non-exists", func(_ context.Context) (map[string]any, error) {
			return map[string]any{}, nil
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, web.ErrCreateRendererFailed)
	})

	t.Run("data func fails", func(t *testing.T) {
		t.Parallel()

		r := testRendererReadyForLayoutData(t)

		err := r.AddLayoutData(testdata.ExampleContext, "default", func(ctx context.Context) (map[string]any, error) {
			return nil, errSomeError
		})
		assert.NoError(t, err)

		buf := &bytes.Buffer{}
		err = r.Render(ctx, buf, testdata.ExampleContext, "p0", nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, web.ErrRenderFailed)
		assert.Empty(t, buf.String())
	})
}

func testRendererReadyForLayoutData(t *testing.T) *web.Renderer {
	t.Helper()

	renderer, err := web.NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.FilesAddBaseData(), template.FuncMap{}, false)
	assert.NoError(t, err)

	err = renderer.AddBaseData("default", func(_ context.Context) (map[string]any, error) {
		return map[string]any{
			"baseTitle": "baseTitle 0",
		}, nil
	})
	assert.NoError(t, err)

	err = renderer.AddContext(testdata.ExampleContext, testdata.FilesAddLayoutData())
	assert.NoError(t, err)

	return renderer
}
