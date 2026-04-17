-- name: CreateGamePlayer :one
   INSERT INTO game_player (game_id, player_id)
   VALUES ($1, $2)
RETURNING *;
