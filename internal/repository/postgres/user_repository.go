package postgres

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type userRepository struct {
	q Querier
}

func NewUserRepository(q Querier) domain.UserRepository {
	return &userRepository{q: q}
}

func (ur *userRepository) Upsert(ctx context.Context, user *domain.User) error {
	const q = `
		INSERT INTO users (id, name, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name,
		    team_name = EXCLUDED.team_name,
		    is_active = EXCLUDED.is_active;
	`

	_, err := ur.q.Exec(ctx, q, user.ID, user.Name, user.TeamName, user.IsActive)
	return err
}

func (ur *userRepository) FetchByID(ctx context.Context, userID string) (*domain.User, error) {
	const q = `
		SELECT id, name, team_name, is_active
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := ur.q.QueryRow(ctx, q, userID).Scan(&user.ID, &user.Name, &user.TeamName, &user.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (ur *userRepository) FetchByTeam(ctx context.Context, teamName string) ([]*domain.User, error) {
	const q = `
		SELECT id, name, team_name, is_active
		FROM users
		WHERE team_name = $1
	`

	rows, err := ur.q.Query(ctx, q, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (*domain.User, error) {
		var user domain.User
		if err := r.Scan(&user.ID, &user.Name, &user.TeamName, &user.IsActive); err != nil {
			return nil, err
		}

		return &user, nil
	})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (ur *userRepository) FetchActiveByTeam(ctx context.Context, teamName string, excludeIDs ...string) ([]*domain.User, error) {
	var (
		rows pgx.Rows
		err  error
	)

	if len(excludeIDs) == 0 {
		const qNoExclude = `
			SELECT id, name, team_name, is_active
			FROM users
			WHERE team_name = $1
			  AND is_active = TRUE;
		`
		rows, err = ur.q.Query(ctx, qNoExclude, teamName)
	} else {
		const q = `
			SELECT id, name, team_name, is_active
			FROM users
			WHERE team_name = $1
			  AND is_active = TRUE
			  AND NOT (id = ANY($2));
		`
		rows, err = ur.q.Query(ctx, q, teamName, excludeIDs)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (*domain.User, error) {
		var u domain.User
		if err := r.Scan(&u.ID, &u.Name, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		return &u, nil
	})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (ur *userRepository) Exists(ctx context.Context, userID string) (bool, error) {
	const q = `
        SELECT EXISTS (
            SELECT 1
            FROM users
            WHERE id = $1
        );
    `

	var exists bool
	err := ur.q.QueryRow(ctx, q, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (ur *userRepository) UpdateIsActive(ctx context.Context, userID string, active bool) (*domain.User, error) {
	const q = `
        UPDATE users
        SET is_active = $1
        WHERE id = $2
        RETURNING id, name, team_name, is_active;
    `

	var user domain.User
	err := ur.q.QueryRow(ctx, q, active, userID).Scan(
		&user.ID,
		&user.Name,
		&user.TeamName,
		&user.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}
