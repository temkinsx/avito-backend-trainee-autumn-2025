package dto

import "avito-backend-trainee-autumn-2025/internal/domain"

type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamDTO struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

type TeamAddRequest = TeamDTO

type TeamAddResponse struct {
	Team TeamDTO `json:"team"`
}

type TeamGetResponse = TeamDTO

func ToTeamMemberDTOs(users []*domain.User) []TeamMemberDTO {
	res := make([]TeamMemberDTO, 0, len(users))

	for _, u := range users {
		res = append(res, TeamMemberDTO{
			UserID:   u.ID,
			Username: u.Name,
			IsActive: u.IsActive,
		})
	}

	return res
}

func ToTeamDTO(team *domain.Team) TeamDTO {
	return TeamDTO{
		TeamName: team.Name,
		Members:  ToTeamMemberDTOs(team.Members),
	}
}
