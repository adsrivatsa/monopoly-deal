-- name: CreateGame :one
   INSERT INTO game (display_name, game, settings, game_state)
   VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetGameByPlayer :one
SELECT g.*
  FROM game g
           INNER JOIN game_player gp
           ON gp.game_id = g.game_id
 WHERE gp.player_id = $1
   AND NOT g.completed;