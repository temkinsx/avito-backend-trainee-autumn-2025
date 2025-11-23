package testutils

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// users:
// u1,u2,u3 -> test_team
// u4,u5,u6 -> other_team

// pull_requests:
// pr1 (OPEN)   author u1
// pr2 (MERGED) author u1
// pr3 (OPEN)   author u4

// pr_reviewers:
// pr1: u2, u3
// pr2: u2
// pr3: u5

const (
	// users
	User1ID = "u1"
	User2ID = "u2"
	User3ID = "u3"

	User4ID = "u4"
	User5ID = "u5"
	User6ID = "u6"

	// teams
	TestTeam  = "test_team"
	OtherTeam = "other_team"

	// PRs
	PR1ID = "pr1"
	PR2ID = "pr2"
	PR3ID = "pr3"

	PR1Name = "pr_test1"
	PR2Name = "pr_test2"
	PR3Name = "pr_test3"
)

func PrepareTestTablesWithFixtures(ctx context.Context, db *pgxpool.Pool) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
        TRUNCATE pr_reviewers CASCADE;
        TRUNCATE pull_requests CASCADE;
        TRUNCATE users CASCADE;
        TRUNCATE teams CASCADE;
    `)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO teams (name)
        VALUES ($1), ($2)
    `, TestTeam, OtherTeam)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO users (id, name, team_name, is_active)
        VALUES 
            ($1, 'User One',   $7, true),
            ($2, 'User Two',   $7, true),
            ($3, 'User Three', $7, true),
            ($4, 'User Four',  $8, true),
            ($5, 'User Five',  $8, true),
            ($6, 'User Six',   $8, false)
    `,
		User1ID, User2ID, User3ID,
		User4ID, User5ID, User6ID,
		TestTeam, OtherTeam,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO pull_requests (id, name, author_id, status)
        VALUES
            ($1, $4, $7, 'OPEN'),
            ($2, $5, $7, 'MERGED'),
            ($3, $6, $8, 'OPEN')
    `,
		PR1ID, PR2ID, PR3ID,
		PR1Name, PR2Name, PR3Name,
		User1ID, User4ID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO pr_reviewers (pr_id, user_id)
        VALUES
            ($1, $4),
            ($1, $5),
            ($2, $4),
            ($3, $6)
    `,
		PR1ID, PR2ID, PR3ID,
		User2ID, User3ID, User5ID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
