CREATE TABLE teams
(
    name TEXT PRIMARY KEY
);

CREATE TABLE users
(
    id        TEXT PRIMARY KEY,
    name      TEXT    NOT NULL,
    team_name TEXT    NOT NULL REFERENCES teams (name),
    is_active BOOLEAN NOT NULL
);

CREATE INDEX idx_users_team_name ON users (team_name);

CREATE TABLE pull_requests
(
    id         TEXT PRIMARY KEY,
    name       TEXT        NOT NULL,
    author_id  TEXT REFERENCES users (id),
    status     TEXT        NOT NULL CHECK (status IN ('OPEN', 'MERGED')) DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL                                      DEFAULT now(),
    merged_at  TIMESTAMPTZ
);

CREATE TABLE pr_reviewers
(
    pr_id   TEXT REFERENCES pull_requests (id) ON DELETE CASCADE,
    user_id TEXT REFERENCES users (id),
    PRIMARY KEY (pr_id, user_id)
);