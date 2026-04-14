-- name: CreateRoom :one
   INSERT INTO room (display_name, capacity)
   VALUES ($1, $2)
RETURNING *;

-- name: GetRoom :one
SELECT *
  FROM room
 WHERE room_id = $1;

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