CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS player (
    player_id UUID PRIMARY KEY DEFAULT uuidv7(),
    display_name TEXT NOT NULL,
    email TEXT NOT NULL,
    image_url TEXT NOT NULL
)
