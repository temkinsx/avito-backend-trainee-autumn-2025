package testutils

import "avito-backend-trainee-autumn-2025/internal/domain"

// Готовые структуры пользователей
var (
	User1 = &domain.User{
		ID:       User1ID,
		Name:     "User One",
		TeamName: TestTeam,
		IsActive: true,
	}

	User2 = &domain.User{
		ID:       User2ID,
		Name:     "User Two",
		TeamName: TestTeam,
		IsActive: true,
	}

	User3 = &domain.User{
		ID:       User3ID,
		Name:     "User Three",
		TeamName: TestTeam,
		IsActive: true,
	}

	User4 = &domain.User{
		ID:       User4ID,
		Name:     "User Four",
		TeamName: OtherTeam,
		IsActive: true,
	}

	User5 = &domain.User{
		ID:       User5ID,
		Name:     "User Five",
		TeamName: OtherTeam,
		IsActive: true,
	}

	User6 = &domain.User{
		ID:       User6ID,
		Name:     "User Six",
		TeamName: OtherTeam,
		IsActive: false,
	}
)

// Команды
var (
	TeamTest = &domain.Team{
		Name: TestTeam,
		Members: []*domain.User{
			User1, User2, User3,
		},
	}

	TeamOther = &domain.Team{
		Name: OtherTeam,
		Members: []*domain.User{
			User4, User5, User6,
		},
	}
)

// Pull Requests (без ревьюверов, только сами PR)
var (
	PR1 = &domain.PullRequest{
		ID:       PR1ID,
		Name:     PR1Name,
		AuthorID: User1ID,
		Status:   domain.StatusOpen,
	}

	PR2 = &domain.PullRequest{
		ID:       PR2ID,
		Name:     PR2Name,
		AuthorID: User1ID,
		Status:   domain.StatusMerged,
	}

	PR3 = &domain.PullRequest{
		ID:       PR3ID,
		Name:     PR3Name,
		AuthorID: User4ID,
		Status:   domain.StatusOpen,
	}
)
