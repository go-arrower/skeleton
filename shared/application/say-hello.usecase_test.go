package application_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-arrower/arrower/alog"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/shared/application"
)

func TestSayHelloRequestHandler_H(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		handler := application.NewSayHelloRequestHandler(alog.NewTest(buf))

		res, err := handler.H(context.Background(), application.SayHelloRequest{Name: "Peter"})
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
		assert.Contains(t, buf.String(), "Peter")
	})

	t.Run("validation failures", func(t *testing.T) {
		t.Parallel()

		tests := map[string]struct {
			name string
		}{
			"empty name": {
				"",
			},
			"to long name": {
				"this name is way way way toooooo long to be valid",
			},
		}

		for name, tt := range tests {
			tt := tt
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				buf := &bytes.Buffer{}
				handler := application.NewSayHelloRequestHandler(alog.NewTest(buf))

				res, err := handler.H(context.Background(), application.SayHelloRequest{Name: tt.name})
				assert.Error(t, err)
				assert.Empty(t, res)
				assert.Empty(t, buf.String())
			})
		}
	})

	t.Run("general failure case", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		handler := application.NewSayHelloRequestHandler(alog.NewTest(buf))

		res, err := handler.H(context.Background(), application.SayHelloRequest{Name: "illegal-name"})
		t.Log(err)
		assert.Error(t, err)
		assert.Empty(t, res)
		assert.Empty(t, buf.String())
	})
}
