CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS player(
    player_id uuid PRIMARY KEY DEFAULT uuidv7(),
    display_name text NOT NULL,
    email text NOT NULL,
    image_url text NOT NULL,
    refresh_token_id uuid NOT NULL DEFAULT uuidv7()
);
