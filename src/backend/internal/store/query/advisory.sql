-- name: QuickPlayBucketXactLock :exec
-- Serializes quick-play find-or-create across app instances (released at transaction end).
SELECT pg_advisory_xact_lock(sqlc.arg('lock_class')::integer, sqlc.arg('game_key')::integer);
