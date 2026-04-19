CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS player
    (
        player_id uuid
            PRIMARY KEY DEFAULT uuidv7(),
        display_name text NOT NULL,
        email text NOT NULL,
        image_url text NOT NULL,
        refresh_token_id uuid NOT NULL DEFAULT uuidv7()
    );

CREATE TYPE game_type AS enum ('monopoly_deal');

CREATE TABLE IF NOT EXISTS room
    (
        room_id uuid
            PRIMARY KEY DEFAULT uuidv7(),
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
        game_id uuid
            PRIMARY KEY DEFAULT uuidv7(),
        display_name text NOT NULL,
        game game_type NOT NULL,
        game_state bytea NOT NULL,
        sequence_num int2 NOT NULL DEFAULT 0,
        completed bool NOT NULL DEFAULT FALSE,
        created_at timestamptz NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS game_player
    (
        game_id uuid NOT NULL,
        player_id uuid NOT NULL,
        PRIMARY KEY (game_id, player_id)
    );