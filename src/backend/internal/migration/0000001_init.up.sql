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

CREATE TYPE game AS enum ('monopoly_deal');

CREATE TABLE IF NOT EXISTS room
    (
        room_id uuid
            PRIMARY KEY DEFAULT uuidv7(),
        display_name text NOT NULL,
        capacity int NOT NULL,
        occupied int NOT NULL DEFAULT 1,
        game game NOT NULL,
        settings jsonb NOT NULL DEFAULT '{}',
        created_at timestamptz NOT NULL DEFAULT NOW()
    );

CREATE INDEX idx_rooms_settings ON room USING gin (settings);

CREATE TABLE IF NOT EXISTS room_player
    (
        room_id uuid NOT NULL,
        player_id uuid NOT NULL,
        is_ready bool NOT NULL DEFAULT FALSE,
        is_host bool NOT NULL,
        joined_at timestamptz NOT NULL DEFAULT NOW(),
        PRIMARY KEY (room_id, player_id)
    );