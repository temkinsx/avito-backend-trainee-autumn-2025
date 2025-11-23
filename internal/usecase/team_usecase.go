package usecase

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
)

type teamUsecase struct {
	teamRepository domain.TeamRepository
	userRepository domain.UserRepository
	txManager      domain.TxManager
}

func NewTeamUsecase(teamRepository domain.TeamRepository, userRepository domain.UserRepository, txManager domain.TxManager) domain.TeamUsecase {
	return &teamUsecase{
		teamRepository: teamRepository,
		userRepository: userRepository,
		txManager:      txManager,
	}
}

func (tu *teamUsecase) Add(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	exists, err := tu.teamRepository.Exists(ctx, team.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrNotFound
	}

	err = tu.txManager.WithinTx(ctx, func(ctx context.Context, repos *domain.Repos) error {
		if err := repos.Team.Create(ctx, team.Name); err != nil {
			return err
		}

		for _, m := range team.Members {
			u := &domain.User{
				ID:       m.ID,
				Name:     m.Name,
				TeamName: team.Name,
				IsActive: m.IsActive,
			}
			if err := repos.User.Upsert(ctx, u); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (tu *teamUsecase) ListByName(ctx context.Context, teamName string) (*domain.Team, error) {
	exists, err := tu.teamRepository.Exists(ctx, teamName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrNotFound
	}

	members, err := tu.userRepository.FetchByTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}

	return &domain.Team{
		Name:    teamName,
		Members: members,
	}, nil
}
