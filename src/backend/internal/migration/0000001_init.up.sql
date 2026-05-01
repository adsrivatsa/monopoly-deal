CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS player
    (
        player_id uuid PRIMARY KEY,
        display_name text NOT NULL,
        email text NOT NULL,
        image_url text NOT NULL,
        refresh_token_id uuid NOT NULL
    );

CREATE TYPE game_type AS enum ('monopoly_deal');

CREATE TABLE IF NOT EXISTS room
    (
        room_id uuid PRIMARY KEY,
        display_name text NOT NULL,
        capacity int NOT NULL,
        occupied int NOT NULL DEFAULT 1,
        game game_type NOT NULL,
        settings bytea NOT NULL,
        created_at timestamptz NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS room_player
    (
        room_id uuid NOT NULL,
        player_id uuid NOT NULL,
        is_ready bool NOT NULL DEFAULT FALSE,
        is_host bool NOT NULL,
        joined_at timestamptz NOT NULL DEFAULT NOW(),
        PRIMARY KEY (room_id, player_id)
    );

CREATE TABLE IF NOT EXISTS game
    (
        game_id uuid PRIMARY KEY,
        display_name text NOT NULL,
        game game_type NOT NULL,
        game_state bytea NOT NULL,
        completed bool NOT NULL DEFAULT FALSE,
        winner uuid REFERENCES player (player_id),
        created_at timestamptz NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS game_player
    (
        game_id uuid NOT NULL,
        player_id uuid NOT NULL,
        PRIMARY KEY (game_id, player_id)
    );

CREATE TABLE IF NOT EXISTS game_history
    (
        game_id uuid NOT NULL REFERENCES game (game_id),
        seq_num int2 NOT NULL,
        action_kind text NOT NULL,
        action_version int NOT NULL,
        action bytea NOT NULL,
        created_at timestamptz NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS game_timeout
    (
        game_id uuid NOT NULL REFERENCES game (game_id),
        player_id uuid NOT NULL REFERENCES player (player_id),
        demand_id text,
        token_id uuid NOT NULL,
        created_at timestamptz NOT NULL DEFAULT NOW()
    );

CREATE UNIQUE INDEX ON game_timeout (game_id, player_id) WHERE demand_id IS NULL;
CREATE UNIQUE INDEX ON game_timeout (game_id, player_id, demand_id) WHERE demand_id IS NOT NULL;