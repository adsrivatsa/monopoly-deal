CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS player(
    player_id uuid PRIMARY KEY DEFAULT uuidv7(),
    display_name text NOT NULL,
    email text NOT NULL,
    image_url text NOT NULL,
    refresh_token_id uuid NOT NULL DEFAULT uuidv7()
);

CREATE TYPE room_status AS ENUM(
    'lobby',
    'game',
    'completed'
);

CREATE TABLE IF NOT EXISTS room(
    room_id uuid PRIMARY KEY DEFAULT uuidv7(),
    display_name text NOT NULL,
    capacity int NOT NULL,
    room_status room_status NOT NULL DEFAULT 'lobby',
    game_state text
);

CREATE TABLE IF NOT EXISTS room_player(
    room_id uuid NOT NULL REFERENCES room(room_id),
    player_id uuid NOT NULL REFERENCES player(player_id),
    joined_at timestamptz NOT NULL DEFAULT now(),
    is_host boolean NOT NULL,
    PRIMARY KEY (room_id, player_id)
);

