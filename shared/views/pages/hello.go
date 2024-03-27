package pages

import "github.com/go-arrower/skeleton/shared/domain"

type helloPage struct {
	domain.TeamMember
	TeamTimeFmt string
}

func PresentHello(tm domain.TeamMember) helloPage {
	return helloPage{
		TeamMember:  tm,
		TeamTimeFmt: tm.TeamTime.Format("15:04:05"),
	}
}

func (p helloPage) ShowTeamBanner() bool {
	return p.IsTeamMember
}
