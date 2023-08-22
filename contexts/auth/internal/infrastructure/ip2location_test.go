package infrastructure_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/infrastructure"
)

func TestNewIp2LocationService(t *testing.T) {
	t.Parallel()

	t.Run("get with default db path", func(t *testing.T) {
		t.Parallel()

		ip := infrastructure.NewIP2LocationService("")
		_, err := ip.ResolveIP("127.0.0.1")
		assert.NoError(t, err)
	})

	t.Run("get with db path", func(t *testing.T) {
		t.Parallel()

		ip := infrastructure.NewIP2LocationService("data/IP-COUNTRY-REGION-CITY.BIN")
		_, err := ip.ResolveIP("127.0.0.1")
		assert.NoError(t, err)
	})
}

func TestIP2Location_ResolveIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testName string
		ip       string
		expIP    user.ResolvedIP
		err      error
	}{
		{
			"empty ip",
			"",
			user.ResolvedIP{},
			infrastructure.ErrInvalidIP,
		},
		{
			"invalid ip",
			"this-is-not-an-ip-address",
			user.ResolvedIP{},
			infrastructure.ErrInvalidIP,
		},
		{
			"valid ip",
			"87.118.100.175",
			user.ResolvedIP{
				IP:          net.ParseIP("87.118.100.175"),
				Country:     "Germany",
				CountryCode: "DE",
				Region:      "Thuringen",
				City:        "Erfurt",
			},
			nil,
		},
	}

	ip := infrastructure.NewIP2LocationService("")

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			res, err := ip.ResolveIP(tt.ip)
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.expIP, res)
		})
	}
}
