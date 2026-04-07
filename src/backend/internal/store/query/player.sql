-- name: CreatePlayer :one
INSERT INTO player(display_name, email, image_url)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: GetPlayer :one
SELECT
    *
FROM
    player
WHERE (player_id = sqlc.narg('player_id')
    OR sqlc.narg('player_id') IS NULL)
AND (email = sqlc.narg('email')
    OR sqlc.narg('email') IS NULL)
AND NOT (sqlc.narg('player_id') IS NULL
    AND sqlc.narg('email') IS NULL);

-- name: UpdatePlayer :one
UPDATE
    player
SET
    display_name = $1
WHERE
    player_id = $2
RETURNING
    *;

