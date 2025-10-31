CREATE TYPE problem_difficulty AS ENUM ('Easy', 'Medium', 'Hard');
CREATE TYPE match_status AS ENUM ('Won', 'Canceled', 'Reverted');
CREATE TYPE submission_status AS ENUM (
    'Accepted', 'Compile Error', 'Memory Limit Exceeded', 
    'Output Limit Exceeded', 'Runtime Error', 
    'Time Limit Exceeded', 'Wrong Answer'
);

CREATE TABLE users (
  id            BIGINT PRIMARY KEY,
  access_token  TEXT NOT NULL,
  username      TEXT NOT NULL,
  discriminator VARCHAR(4) NOT NULL,
  lc_username   TEXT NOT NULL DEFAULT '',
  avatar_url    TEXT NOT NULL DEFAULT '',
  created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  rating        SMALLINT DEFAULT 1000,
  UNIQUE (username, discriminator)
);

CREATE TABLE problems (
    id         INTEGER PRIMARY KEY,
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL,
    difficulty problem_difficulty NOT NULL,
    is_paid    BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE tags (
    id   SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE problem_tags (
    problem_id INTEGER NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    tag_id     INT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (problem_id, tag_id)
);

CREATE TABLE matches (
    id          UUID PRIMARY KEY,
    problem_id  INTEGER NOT NULL REFERENCES problems(id) ON DELETE RESTRICT,
    is_rated    BOOLEAN NOT NULL DEFAULT false,
    status      match_status NOT NULL,
    winner_id   BIGINT REFERENCES users(id) ON DELETE SET NULL,
    start_time  TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time    TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE match_players (
    match_id  UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (match_id, player_id)
);

CREATE TABLE submissions (
    match_id            UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    submission_id       BIGINT NOT NULL,
    player_id           BIGINT REFERENCES users(id) ON DELETE SET NULL,
    passed_test_cases   INT NOT NULL,
    total_test_cases    INT NOT NULL,
    status              submission_status NOT NULL,
    runtime             INT,
    runtime_percentile  DECIMAL,
    memory              INT,
    memory_percentile   DECIMAL,
    lang                TEXT,
    submitted_at        TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (match_id, submission_id)
);
