package pages_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/shared/domain"
	"github.com/go-arrower/skeleton/shared/views/pages"
)

func TestHelloPage_ShowTeamBanner(t *testing.T) {
	t.Parallel()

	page := pages.PresentHello(domain.TeamMember{
		Name:         "Peter",
		TeamTime:     time.Now(),
		IsTeamMember: true,
	})

	assert.NotEmpty(t, page.TeamTimeFmt)
	assert.True(t, page.ShowTeamBanner())
}
