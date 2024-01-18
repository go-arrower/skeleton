// Use white box testing, to make it easier to assert on the inner workings of partially loaded and cached templates.
// if a white box test case fails, consider just deleting it over fixing it, to prevent coupling to the implementation.
//
//nolint:testpackage
package template

// todo rename package to renderer, or web, or echo?

import (
	"bytes"
	"math/rand"
	"os"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"go.opentelemetry.io/otel/trace/noop"

	"github.com/go-arrower/arrower/alog"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/shared/infrastructure/template/testdata"
	"github.com/go-arrower/skeleton/shared/interfaces/web/views"
)

func TestNewRenderer(t *testing.T) {
	t.Parallel()

	t.Run("construct renderer", func(t *testing.T) {
		t.Parallel()

		r, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), views.SharedViews, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
	})

	t.Run("fail on missing files", func(t *testing.T) {
		t.Parallel()

		r, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), nil, false)
		assert.Error(t, err)
		assert.Nil(t, r)
	})

	// white box test, if it fails feel free to delete it
	t.Run("initialise raw renderer", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.TemplateFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		// assert component templates from testdata.TemplateFiles is loaded
		// contains always an empty component, so expected 2 +1 => 3
		assert.Len(t, renderer.components.Templates(), 3)

		// assert pages are loaded
		assert.Len(t, renderer.templates, 3)
		// if the file is called p0.page.html, the template is called p0
		assert.NotEmpty(t, renderer.templates["p0"])
		assert.NotEmpty(t, renderer.templates["p1"])
		assert.NotEmpty(t, renderer.templates["p2"])
		assert.Empty(t, renderer.templates["non-existent"])

		// assert each page has itself and all components loaded as a template
		// todo: this is whitebox test... does this make sense?
		//for _, page := range renderer.templates {
		//	fmt.Println("PAGE", page.Name())
		//	fmt.Println(page.Templates())
		//	assert.Len(t, page.Templates(), 3) // todo update comment: 3 is number of components as above
		//}

		// assert template is cached
		// if the file is called global.layout.html, the template is called global
		assert.Len(t, renderer.rawLayouts, 1)
		assert.NotEmpty(t, renderer.rawLayouts["global"])
		assert.Empty(t, renderer.rawLayouts["non-existent"])
	})

	// white box test, if it fails feel free to delete it
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

		// todo log to nil
		renderer, err := NewRenderer(alog.NewTest(os.Stderr), noop.NewTracerProvider(), testdata.TemplateFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		t.Run("component only", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "c0", nil, testdata.NewEmptyEchoContext(t))
			assert.NoError(t, err)

			assert.Equal(t, testdata.C0Content, buf.String())
		})

		t.Run("non existing component", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "non-existing", nil, testdata.NewEmptyEchoContext(t))

			assert.Error(t, err)
			assert.Empty(t, buf.String())
		})

		t.Run("component and page with same name", func(t *testing.T) {
			t.Parallel()

			// TODO
		})
	})

	t.Run("pages", func(t *testing.T) {
		t.Parallel()

		// todo log to nil
		renderer, err := NewRenderer(alog.NewTest(os.Stderr), noop.NewTracerProvider(), testdata.TemplateFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		t.Run("raw page without layout", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "p0", nil, testdata.NewEmptyEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.P0Content)
			assert.NotContains(t, buf.String(), testdata.LContent)
		})

		t.Run("non existing page", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "non-existing", nil, testdata.NewEmptyEchoContext(t))
			assert.Error(t, err) // todo use ErrorIs
			assert.Empty(t, buf.String())
		})

		t.Run("page with components", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "p1", nil, testdata.NewEmptyEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.P1Content)
			assert.Contains(t, buf.String(), testdata.C0Content) // todo: does the renderer have to load LayoutsPagesAndComponents instead ???
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
			err = renderer.Render(buf, "p2", nil, testdata.NewEmptyEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.P2Content)
		})

		t.Run("fragment 0", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "p2#f0", nil, testdata.NewEmptyEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.F0Content)
		})

		t.Run("fragment 1", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "p2#f1", nil, testdata.NewEmptyEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.F1Content)
		})

		t.Run("non existing fragment", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "p1#f1", nil, testdata.NewEmptyEchoContext(t))
			assert.Error(t, err) // todo switch to ErrorIs
			assert.Empty(t, buf.String())
		})
	})

	t.Run("layouts", func(t *testing.T) {
		t.Parallel()

		// todo nil
		renderer, err := NewRenderer(alog.NewTest(os.Stderr), noop.NewTracerProvider(), testdata.LayoutsPagesAndComponents, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		t.Run("page with different layouts", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "global=>p0", nil, testdata.NewEmptyEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.LContent)
			assert.Contains(t, buf.String(), testdata.P0Content)
			assert.Contains(t, buf.String(), testdata.C0Content)
			assert.NotContains(t, buf.String(), testdata.C1Content)

			buf.Reset()
			err = renderer.Render(buf, "other=>p0", nil, testdata.NewEmptyEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.LOtherContent)
			assert.NotContains(t, buf.String(), testdata.LContent)
			assert.Contains(t, buf.String(), testdata.P0Content)
			assert.Contains(t, buf.String(), testdata.C0Content)
			assert.NotContains(t, buf.String(), testdata.C1Content)
		})

		t.Run("access layout that does not exist", func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "nonExisting=>p0", nil, testdata.NewEmptyEchoContext(t))
			assert.Error(t, err)

			assert.Empty(t, buf.String())
		})

		t.Run("rely on default layout when rendering page", func(t *testing.T) {
			t.Parallel()

			renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.LayoutWithDefault, false)
			assert.NoError(t, err)
			assert.NotNil(t, renderer)

			buf := &bytes.Buffer{}
			err = renderer.Render(buf, "=>p0", nil, testdata.NewEmptyEchoContext(t))
			assert.NoError(t, err)

			t.Log(buf.String())
			assert.Contains(t, buf.String(), testdata.LDefaultContent)
			assert.Contains(t, buf.String(), testdata.P0Content)

			// change default layout and render same page again
			err = renderer.SetDefaultLayout("other")
			assert.NoError(t, err)

			buf.Reset()
			err = renderer.Render(buf, "=>p0", nil, testdata.NewEmptyEchoContext(t))
			assert.NoError(t, err)

			assert.Contains(t, buf.String(), testdata.LOtherContent)
			assert.Contains(t, buf.String(), testdata.P0Content)
		})
	})

	// white box test: if it fails, feel free to delete it
	t.Run("multiple pages and increase template cache", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.LayoutsPagesAndComponents, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		// assert only the pure pages are loaded, because others are not cached yet
		assert.Len(t, renderer.templates, 2)

		buf := &bytes.Buffer{}
		err = renderer.Render(buf, "global=>p0", nil, testdata.NewEmptyEchoContext(t))
		assert.NoError(t, err)
		assert.Len(t, renderer.templates, 3) // previous templates + global=>p0

		assert.Contains(t, buf.String(), testdata.LContent)
		assert.Contains(t, buf.String(), testdata.P0Content)
		assert.Contains(t, buf.String(), testdata.C0Content)
		assert.NotContains(t, buf.String(), testdata.C1Content)

		buf.Reset()
		err = renderer.Render(buf, "global=>p1", nil, testdata.NewEmptyEchoContext(t))
		assert.NoError(t, err)
		assert.Len(t, renderer.templates, 4) // previous templates + global=>p1

		assert.Contains(t, buf.String(), testdata.LContent)
		assert.Contains(t, buf.String(), testdata.P1Content)
		assert.NotContains(t, buf.String(), testdata.C0Content)
		assert.NotContains(t, buf.String(), testdata.C1Content)

		buf.Reset()
		err = renderer.Render(buf, "global=>p1", nil, testdata.NewEmptyEchoContext(t))
		assert.NoError(t, err)
		assert.Len(t, renderer.templates, 4) // template is cached already, so no change
	})

	t.Run("render in parallel", func(t *testing.T) {
		t.Parallel()

		// setup multiple random pages
		fs := fstest.MapFS{
			"default.layout.html": {Data: []byte(testdata.LContent + ` {{template "content" .}}`)},
		}

		const numPages = 10
		var pages []string

		for i := 0; i < numPages; i++ {
			p := randomString(5)
			fs["pages/"+p+".html"] = &fstest.MapFile{Data: []byte(p)} //nolint:exhaustruct

			pages = append(pages, p)
		}

		// test
		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), fs, true)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		wg := &sync.WaitGroup{}

		const numPageLoads = 100
		wg.Add(numPageLoads)
		for i := 0; i < numPageLoads; i++ {
			go func() {
				n := rand.Intn(numPages) //nolint:gosec // used for simulating page visit not for security

				page := pages[n]

				buf := &bytes.Buffer{}
				err := renderer.Render(buf, "=>"+page, nil, testdata.NewEmptyEchoContext(t))
				assert.NoError(t, err, page)
				assert.Contains(t, buf.String(), page)

				wg.Done()
			}()
		}

		wg.Wait()
	})
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

	t.Run("only one layout file, so it becomes the default", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.LayoutOneLayout, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		assert.Equal(t, "global", renderer.Layout())
	})

	t.Run("multiple layouts but with default", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.LayoutWithDefault, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		assert.Equal(t, "default", renderer.Layout())
	})
}

func TestRenderer_SetDefaultLayout(t *testing.T) {
	t.Parallel()

	t.Run("set existing default layout", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.LayoutWithDefault, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		err = renderer.SetDefaultLayout("other")
		assert.NoError(t, err)
		assert.Equal(t, "other", renderer.Layout())
	})

	t.Run("set non existing layout", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(alog.NewTest(nil), noop.NewTracerProvider(), testdata.LayoutWithDefault, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		err = renderer.SetDefaultLayout("non-existing")
		assert.Error(t, err)
		assert.Equal(t, "default", renderer.Layout())
	})
}

// white box test, if it fails feel free to delete it.
func TestParseLayoutAndPage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testName       string
		name           string
		expectedLayout string
		expectedPage   string
	}{
		{
			"empty",
			"",
			"",
			"",
		},
		{
			"just page",
			"p0",
			"",
			"p0",
		},
		{
			"layout and page",
			"l=>p",
			"l",
			"p",
		},
		{
			"trim whitespaces",
			" l => p ",
			"l",
			"p",
		},
		{
			"layout, sub-layout, and page",
			"l=>s=>p",
			"l=>s",
			"p",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			l, p := parseLayoutAndPage(tt.name)
			assert.Equal(t, tt.expectedLayout, l)
			assert.Equal(t, tt.expectedPage, p)
		})
	}
}

/*
Additional API
- AddContext to add more views that are not global but context specific
	- call twice, and it should fail, as it is already a context renderer
	- keep original renderer unchanged, so it can be continued to used as is and for other contexts as well
	- call with global layout, sub-layout, and page
- E-Mail renderer instead of web renderer
*/

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString(n int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec // used for ids, not security

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rnd.Intn(len(letters))]
	}

	return string(b)
}
