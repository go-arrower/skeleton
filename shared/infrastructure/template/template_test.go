// Use white box testing, to make it easier to assert on the inner workings of partially loaded and cached templates.
// if a white box test case fails, consider just deleting it over fixing it, to prevent coupling to the implementation.
//
//nolint:testpackage
package template

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/shared/infrastructure/template/testdata"
	"github.com/go-arrower/skeleton/shared/interfaces/web/views"
)

func TestNewRenderer(t *testing.T) {
	t.Parallel()

	t.Run("construct renderer", func(t *testing.T) {
		t.Parallel()

		r, err := NewRenderer(views.SharedViews, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
	})

	t.Run("fail on missing files", func(t *testing.T) {
		t.Parallel()

		r, err := NewRenderer(nil, false)
		assert.Error(t, err)
		assert.Nil(t, r)
	})

	// white box test, if it fails feel free to delete it
	t.Run("initialise raw renderer", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(testdata.SimpleFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		// assert component templates from testdata.SimpleFiles is loaded
		// contains always an empty component, so expected 2 +1 => 3
		assert.Len(t, renderer.components.Templates(), 3)

		// assert pages are loaded
		assert.Len(t, renderer.templates, 2)
		// if the file is called p0.page.html, the template is called p0
		assert.NotEmpty(t, renderer.templates["p0"])
		assert.NotEmpty(t, renderer.templates["p1"])
		assert.Empty(t, renderer.templates["non-existent"])

		// assert each page has itself and all components loaded as a template
		for _, page := range renderer.templates {
			assert.Len(t, page.Templates(), 4)
		}

		// assert template is cached
		// if the file is called global.layout.html, the template is called global
		assert.Len(t, renderer.rawLayouts, 1)
		assert.NotEmpty(t, renderer.rawLayouts["global"])
		assert.Empty(t, renderer.rawLayouts["non-existent"])
	})

	// white box test, if it fails feel free to delete it
	t.Run("fs with no files", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(testdata.EmptyFiles, false)
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

	t.Run("render shared pages without layout", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(testdata.SimpleFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		buf := &bytes.Buffer{}
		err = renderer.Render(buf, "p0", nil, nil)
		assert.NoError(t, err)

		assert.Contains(t, buf.String(), testdata.P0Content)
	})

	t.Run("render shared pages with components", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(testdata.LayoutsPagesAndComponents, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		buf := &bytes.Buffer{}
		err = renderer.Render(buf, "p0", nil, nil)
		assert.NoError(t, err)

		assert.Contains(t, buf.String(), testdata.P0Content)
		assert.Contains(t, buf.String(), testdata.C0Content)
		assert.NotContains(t, buf.String(), testdata.C1Content)
	})

	t.Run("render shared page with different layouts", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(testdata.LayoutsPagesAndComponents, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		buf := &bytes.Buffer{}
		err = renderer.Render(buf, "global=>p0", nil, nil)
		assert.NoError(t, err)

		assert.Contains(t, buf.String(), testdata.LContent)
		assert.Contains(t, buf.String(), testdata.P0Content)
		assert.Contains(t, buf.String(), testdata.C0Content)
		assert.NotContains(t, buf.String(), testdata.C1Content)

		buf.Reset()
		err = renderer.Render(buf, "other=>p0", nil, nil)
		assert.NoError(t, err)

		assert.Contains(t, buf.String(), "other")
		assert.NotContains(t, buf.String(), testdata.LContent)
		assert.Contains(t, buf.String(), testdata.P0Content)
		assert.Contains(t, buf.String(), testdata.C0Content)
		assert.NotContains(t, buf.String(), testdata.C1Content)
	})

	// white box test, if it fails feel free to delete it
	t.Run("render multiple pages and increase template cache", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(testdata.LayoutsPagesAndComponents, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		// assert only the pure pages are loaded, because others are not cached yet
		assert.Len(t, renderer.templates, 2)

		buf := &bytes.Buffer{}
		err = renderer.Render(buf, "global=>p0", nil, nil)
		assert.NoError(t, err)
		assert.Len(t, renderer.templates, 3) // previous templates + global=>p0

		assert.Contains(t, buf.String(), testdata.LContent)
		assert.Contains(t, buf.String(), testdata.P0Content)
		assert.Contains(t, buf.String(), testdata.C0Content)
		assert.NotContains(t, buf.String(), testdata.C1Content)

		buf.Reset()
		err = renderer.Render(buf, "global=>p1", nil, nil)
		assert.NoError(t, err)
		assert.Len(t, renderer.templates, 4) // previous templates + global=>p1

		assert.Contains(t, buf.String(), testdata.LContent)
		assert.Contains(t, buf.String(), testdata.P1Content)
		assert.NotContains(t, buf.String(), testdata.C0Content)
		assert.NotContains(t, buf.String(), testdata.C1Content)

		buf.Reset()
		err = renderer.Render(buf, "global=>p1", nil, nil)
		assert.NoError(t, err)
		assert.Len(t, renderer.templates, 4) // template is cached already, so no change
	})

	t.Run("render component", func(t *testing.T) {
		t.Parallel()

		renderer, err := NewRenderer(testdata.SimpleFiles, false)
		assert.NoError(t, err)
		assert.NotNil(t, renderer)

		buf := &bytes.Buffer{}
		err = renderer.Render(buf, "c0.component", nil, nil)
		assert.NoError(t, err)

		assert.Contains(t, buf.String(), testdata.C0Content)
	})

	/*
		test cases
		- Test parallel rendering, to prevent race conditions
		- use layout name that does not exist
		- hotreload from local fs
		- call page & layouts that do not exist
	*/
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
