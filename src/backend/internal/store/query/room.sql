-- name: CreateRoom :one
   INSERT INTO room (display_name, capacity, game, settings)
   VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRoom :one
SELECT *
  FROM room
 WHERE room_id = $1;

-- name: GetRoomByPlayer :one
SELECT r.*
  FROM room r
           INNER JOIN room_player rp
           ON rp.room_id = r.room_id
 WHERE rp.player_id = $1;

-- name: UpdateRoomCapacity :one
   UPDATE room
      SET capacity = $1
    WHERE room_id = $2
RETURNING *;

-- name: UpdateRoomOccupied :one
   UPDATE room
      SET occupied = $1
    WHERE room_id = $2
RETURNING *;

-- name: DeleteRoom :exec
DELETE
  FROM room
 WHERE room_id = $1;

-- name: IncrementRoomOccupied :one
   UPDATE room
      SET occupied = occupied + 1
    WHERE room_id = $1
RETURNING *;

-- name: DecrementRoomOccupied :one
   UPDATE room
      SET occupied = occupied - 1
    WHERE room_id = $1
RETURNING *;

-- name: UpdateRoomSettings :one
   UPDATE room
      SET capacity = $1,
          game = $2,
          settings = $3
    WHERE room_id = $4
RETURNING *;

-- name: ListRooms :many
SELECT rp.*, r.display_name AS room_display_name, r.capacity AS room_capacity, r.occupied AS room_occupied,
       r.game AS room_game, r.settings AS room_settings, r.created_at AS room_created_at,
       p.display_name AS host_display_name, p.image_url AS host_image_url, COUNT(*) OVER () AS total_count
  FROM room_player rp
           INNER JOIN room r
           ON r.room_id = rp.room_id
           INNER JOIN player p
           ON p.player_id = rp.player_id
 WHERE (r.room_id::text = sqlc.narg('search')::text OR r.room_id::text ILIKE '%' || sqlc.narg('search') || '%' OR
        p.player_id::text = sqlc.narg('search') OR p.player_id::text ILIKE '%' || sqlc.narg('search') || '%' OR
        r.display_name = sqlc.narg('search') OR r.display_name ILIKE '%' || sqlc.narg('search') || '%' OR
        p.display_name = sqlc.narg('search') OR p.display_name ILIKE '%' || sqlc.narg('search') || '%' OR
        sqlc.narg('search') IS NULL)
   AND (r.game = sqlc.narg('game') OR sqlc.narg('game') IS NULL)
   AND rp.is_host
 ORDER BY r.display_name
 LIMIT $1 OFFSET $2;
