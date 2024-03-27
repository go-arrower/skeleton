package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrInvalidName = errors.New("invalid name")

type MemberName string

type TeamMember struct {
	Name         MemberName
	TeamTime     time.Time
	IsTeamMember bool
}

func IsTeamMember(name MemberName) (TeamMember, error) {
	if !isValidName(name) {
		return TeamMember{}, ErrInvalidName
	}

	return TeamMember{
		Name:         name,
		TeamTime:     time.Now(),
		IsTeamMember: isTeamMember(name),
	}, nil
}

func isValidName(name MemberName) bool {
	const illegalChar = "-"

	if strings.Contains(string(name), illegalChar) {
		return false
	}

	return true
}

func isTeamMember(name MemberName) bool {
	teamMembers := map[MemberName]struct{}{
		"Peter": {},
	}

	if _, ok := teamMembers[name]; ok {
		return true
	}

	return false
}
