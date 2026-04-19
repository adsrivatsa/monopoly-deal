-- name: CreateGame :one
   INSERT INTO game (display_name, game, game_state)
   VALUES ($1, $2, $3)
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
      SET game_state = $1,
          sequence_num = sequence_num + 1
    WHERE game_id = $2
RETURNING *;