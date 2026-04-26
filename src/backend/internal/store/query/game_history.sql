-- name: CreateGameHistory :one
INSERT INTO game_history (game_id, seq_num, action_kind, action_version, action) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: ListGameHistory :many
SELECT * FROM game_history WHERE game_id = $1 ORDER BY seq_num LIMIT $2 OFFSET $3;