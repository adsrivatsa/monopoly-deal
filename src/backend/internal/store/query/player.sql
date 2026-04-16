-- name: CreatePlayer :one
   INSERT INTO player(display_name, email, image_url)
   VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPlayer :one
SELECT *
  FROM player
 WHERE (player_id = sqlc.narg('player_id') OR sqlc.narg('player_id') IS NULL)
   AND (email = sqlc.narg('email') OR sqlc.narg('email') IS NULL)
   AND NOT (sqlc.narg('player_id') IS NULL AND sqlc.narg('email') IS NULL);

-- name: UpdatePlayer :one
   UPDATE player
      SET display_name = $1
    WHERE player_id = $2
RETURNING *;

-- name: GetPlayers :many
SELECT *
  FROM player
 WHERE (played_id = ANY (@player_ids::uuid[]));

-- name: GetPlayersByRoom :many
SELECT p.*, rp.is_ready, rp.is_host
  FROM player p
           INNER JOIN room_player rp
           ON rp.player_id = p.player_id
 WHERE rp.room_id = $1
 ORDER BY rp.joined_at;