# Deal No Mercy — implementation roadmap

Adding Monopoly Deal: No Mercy as the second game on The Deal. Same stack
conventions as classic: protobuf for client↔server messages, msgpack for game
state persisted in Postgres, sqlc for query generation. Assets already shipped
under `src/backend/public/deal-no-mercy/card/` (see `card-list.md` for the full
120-card set and `asset-style-guide.md` for the visual language).

Status: Phases 0–4 complete and committed. Phase 5 integration verification
complete — a scripted backend E2E harness (`src/backend/cmd/e2e`) drives a full
two-player No Mercy game over the real HTTP + websocket protocol and passes
41/41 checks (all mechanics, timeouts, win, reconnect, cross-game isolation).
Browser playtest still pending (human).

Each phase lands as its own verified commit(s); later phases depend on earlier
ones. Engine work is test-driven — the engine is pure Go with no I/O, so it
gets the deepest coverage.

## Phase 0 — Game type plumbing (backend) — DONE

Teach the platform that a second game exists, without exposing it to players
yet.

- `migration/0000001_init.up.sql`: add `deal_no_mercy` to the `game_type` enum
  (repo practice: init migration is edited in place during development).
- `sqlc generate` → `store.GameTypeDealNoMercy`, `Valid()`, `AllGameTypeValues()`.
- `schema/room.proto`: add `DealNoMercy = 1` to the `Game` enum; `make gen`.
- `store/enum.go`: map `GameTypeDealNoMercy` ↔ `room_schema.Game_DealNoMercy`.
- Frontend stays monopoly-deal-only until Phase 4 (no lobby exposure).

Verify: backend builds, existing tests pass, frontend typechecks untouched.

## Phase 1 — Game engine (`internal/engine/deal-no-mercy/`) — DONE

Pure-Go engine package mirroring the classic engine's idioms (Identifier
generator, Cards/PropertySets, msgpack Settings + Snapshot) with No Mercy
rules from `card-list.md`:

- Card definitions and deck composition summing to exactly 120 cards.
- Debt chips: 3 per player, issued on failed payment after surrendering
  everything; each outstanding chip must be repaid with any one hand card at
  the start of the debtor's next turn, consuming one of the 3 plays.
- New actions: Set Snatcher (3M), Debt Trap, Go Again!, Heist, Market Crash,
  Big Payday (draw to 7), Repo Man, Shack (+5M rent attachment, any set, max 1),
  Property Raid, Tax Day, Pickpocket, Bank Swap, Yoink!, Rent-any-color and
  Double Rent variants that charge **all** players. NAH! denial as in classic.
- Removed relative to classic: Sly/Forced Deal, Debt Collector, Birthday,
  House/Hotel, Pass Go, Double-The-Rent modifier.
- Win: first to 3 complete sets.

Tests: deck composition/unique IDs, snapshot round-trip (incl. debt state),
per-action semantics + edge cases, debt chip lifecycle, no card duplication
across zones (learned that lesson in classic), win detection.

## Phase 2 — Wire protocol (`schema/deal_no_mercy.proto`) — DONE

Mirror `monopoly_deal.proto`'s shape for the new game: GameState, Card/asset
enums, demands, actions (with masked variants for hidden information), client
and server messages, plus debt chip state. Add the payload arm to
`gateway.proto`. `make gen` for Go + TS.

Verify: both codebases build with generated code, no changes to existing
message numbering.

## Phase 3 — Service layer + routing (`internal/service/deal-no-mercy/`) — DONE

- Controller: create-game-from-room, event handling, action masking, timeout /
  default-move scheduling, msgpack persistence — following the classic
  controller's structure.
- Asset image maps pointing at `/static/deal-no-mercy/card/{large,small}` (all
  faces self-contained in the game's own directory).
- Routing: game-type switches in room service (create, quick play lock key,
  CreateGameFromRoomTx), game socket event dispatch, event bus kinds.

Verify: backend builds + tests; a room with game=deal_no_mercy can start a
game end-to-end against a local DB.

## Phase 4 — Frontend — DONE (build/typecheck; browser playtest pending)

- Generated TS protos; settings model + defaults (capacity 2–5, deck of 120).
- Enable in lobby: `supportedGames`, create-room modal, quick play.
- Game UI: extend the board for No Mercy — debt chip display, new demand
  overlays (Heist/Market Crash picks, Repo Man / Tax Day distribution,
  Property Raid color pick, Bank Swap, Pickpocket type naming, Go Again),
  all-players rent flow. Reuse the soft-tock sound palette.

Verify: typecheck + build; manual play-through in two browsers.

## Phase 5 — Integration polish — DONE (E2E acceptance passing; browser playtest pending)

Timeout default moves for every new demand type, action history sidebar
entries for new actions, reconnect behavior, docs refresh, and a full
two-player E2E game as the acceptance test.

E2E acceptance (`src/backend/cmd/e2e`, run against a live stack): seeds players
directly, mints session cookies with the server's own token package, and drives
the real room→game HTTP flow + game websockets exchanging proto messages. It
verifies the initial deal (5 dealt + starter's draw), masking (opponent hand /
discard / start-turn are counts only), debt chips, asset image URLs, the core
turn loop, and one full flow of every new mechanic (shack +rent, big payday, go
again, yoink, rent-all double variant, heist, market crash, property raid, set
snatcher, tax day distribution, pickpocket, bank swap, debt trap, NAH!
deny+counter, debt-chip issue on short payment + mandatory settlement), the
move- and demand-timeout default paths (scoped: a classic monopoly_deal game
runs simultaneously and is proven untouched), the win (3 sets → WonGame + game
row completed), and mid-game reconnect (authoritative snapshot + masked history
replay). 41/41 checks pass.

Integration bug fixed during verification: the game-socket handlers
(`cmd/api/deal-no-mercy.go` and the shared-path `cmd/api/monopoly-deal.go`)
removed their `gameSockets` map entry on disconnect by comparing against
`oldSock` (the previous connection) instead of `sock` (the current one),
leaving stale entries and mis-deleting a reconnecting socket's entry. Corrected
both to `s2 == sock`, matching the room-socket handlers.
