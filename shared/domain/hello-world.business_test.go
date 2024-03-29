package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/shared/domain"
)

func TestIsTeamMember(t *testing.T) {
	t.Parallel()

	t.Run("valid name & team member", func(t *testing.T) {
		t.Parallel()

		name := domain.MemberName("Peter")

		res, err := domain.IsTeamMember(name)
		assert.NoError(t, err)
		assert.Equal(t, name, res.Name)
		assert.True(t, res.IsTeamMember)
		assert.NotEmpty(t, res.TeamTime)
	})

	t.Run("valid name", func(t *testing.T) {
		t.Parallel()

		name := domain.MemberName("Goku")

		res, err := domain.IsTeamMember(name)
		assert.NoError(t, err)
		assert.Equal(t, name, res.Name)
		assert.False(t, res.IsTeamMember)
		assert.NotEmpty(t, res.TeamTime)
	})

	t.Run("invalid name", func(t *testing.T) {
		t.Parallel()

		res, err := domain.IsTeamMember("invalid-name")
		assert.ErrorIs(t, err, domain.ErrInvalidName)
		assert.Empty(t, res)
	})
}
