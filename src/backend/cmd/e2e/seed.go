package main

import (
	"context"
	"fmt"
	deal_no_mercy "the-deal/internal/engine/deal-no-mercy"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a pgx pool for direct seeding and authoritative state manipulation
// (permitted by the E2E brief: "you may set the deck/hands by writing game
// state directly to the DB between steps if pure play can't reach a scenario").
type DB struct {
	pool *pgxpool.Pool
}

func NewDB(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &DB{pool: pool}, nil
}

func (d *DB) Close() { d.pool.Close() }

// seedPlayer inserts a player row directly (bypassing Google OAuth).
func (d *DB) seedPlayer(ctx context.Context, name string) (uuid.UUID, error) {
	id := uuid.New()
	_, err := d.pool.Exec(ctx, `
		INSERT INTO player (player_id, display_name, email, image_url, refresh_token_id)
		VALUES ($1, $2, $3, $4, $5)`,
		id, name, name+"@e2e.test", "https://example.com/avatar.png", uuid.New())
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}

// cleanup wipes all game/room/player rows so reruns start clean.
func (d *DB) cleanup(ctx context.Context) error {
	stmts := []string{
		`DELETE FROM game_timeout`,
		`DELETE FROM game_history`,
		`DELETE FROM game_player`,
		`UPDATE game SET winner = NULL`,
		`DELETE FROM game`,
		`DELETE FROM room_player`,
		`DELETE FROM room`,
		`DELETE FROM player`,
	}
	for _, s := range stmts {
		if _, err := d.pool.Exec(ctx, s); err != nil {
			return fmt.Errorf("cleanup %q: %w", s, err)
		}
	}
	return nil
}

// loadGame reads and decodes the authoritative game state for a game id.
func (d *DB) loadGame(ctx context.Context, gameID uuid.UUID) (*deal_no_mercy.Game, error) {
	var buf []byte
	err := d.pool.QueryRow(ctx, `SELECT game_state FROM game WHERE game_id = $1`, gameID).Scan(&buf)
	if err != nil {
		return nil, err
	}
	return deal_no_mercy.DecodeMsgpack(buf)
}

// saveGame writes an authoritative game state back to the DB.
func (d *DB) saveGame(ctx context.Context, gameID uuid.UUID, game *deal_no_mercy.Game) error {
	buf, err := game.EncodeMsgpack()
	if err != nil {
		return err
	}
	ct, err := d.pool.Exec(ctx, `UPDATE game SET game_state = $1 WHERE game_id = $2`, buf, gameID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() != 1 {
		return fmt.Errorf("saveGame: expected 1 row, got %d", ct.RowsAffected())
	}
	return nil
}

// gameIDForPlayer returns the game a player is in.
func (d *DB) gameIDForPlayer(ctx context.Context, playerID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := d.pool.QueryRow(ctx,
		`SELECT g.game_id FROM game g JOIN game_player gp ON gp.game_id = g.game_id WHERE gp.player_id = $1`,
		playerID).Scan(&id)
	return id, err
}

func (d *DB) roomIDForPlayer(ctx context.Context, playerID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := d.pool.QueryRow(ctx,
		`SELECT room_id FROM room_player WHERE player_id = $1`, playerID).Scan(&id)
	return id, err
}

func (d *DB) gameCompleted(ctx context.Context, gameID uuid.UUID) (bool, *uuid.UUID, error) {
	var completed bool
	var winner *uuid.UUID
	err := d.pool.QueryRow(ctx, `SELECT completed, winner FROM game WHERE game_id = $1`, gameID).Scan(&completed, &winner)
	return completed, winner, err
}

