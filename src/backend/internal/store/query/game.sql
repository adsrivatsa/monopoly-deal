-- name: CreateGame :one
   INSERT INTO game (game_id, display_name, game, game_state)
   VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetGameByPlayer :one
SELECT g.*
  FROM game g
           INNER JOIN game_player gp
           ON gp.game_id = g.game_id
 WHERE gp.player_id = $1
   AND NOT g.completed;

-- name: UpdateGameState :one
   UPDATE game
      SET game_state = $1
    WHERE game_id = $2
RETURNING *;

-- name: CompleteGame :one
UPDATE game SET completed = TRUE, winner = $1 WHERE game_id = $2 RETURNING *;

-- name: GetGameIDByPlayer :one
SELECT g.game_id FROM game g
INNER JOIN game_player gp
    ON gp.game_id = g.game_id
WHERE gp.player_id = $1
  AND NOT g.completed;