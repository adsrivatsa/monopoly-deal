-- name: CreateGamePlayer :one
   INSERT INTO game_player (game_id, player_id)
   VALUES ($1, $2)
RETURNING *;

-- name: CreateGamePlayersFromRoom :exec
INSERT INTO game_player(game_id, player_id)
SELECT $1, player_id
  FROM room_player
 WHERE room_id = $2;