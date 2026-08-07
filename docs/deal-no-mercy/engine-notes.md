# Deal No Mercy — engine notes (Phase 1)

Package: `src/backend/internal/engine/deal-no-mercy/` (package `deal_no_mercy`).
Pure Go, no I/O, no protobuf references — `Proto()` conversions are deferred to
Phase 2/3 when `schema/deal_no_mercy.proto` exists. Idioms mirror the classic
engine (`Identifier`/`IdentifierGenerator` with the `%03x` counter, `Card`/
`Cards` value types, `PropertySet`/`PropertySets`, msgpack `Settings` +
`GameSnapshot`, demand system with NAH! `Deny` re-issue).

## Decisions on ambiguous rules

1. **120-card total vs card-list counts.** The per-face counts in
   `card-list.md` sum to 115 playing cards; the advertised "120 Cards" is only
   reached by counting the 5 reference cards, which have no gameplay role.
   The engine ships no reference cards, and the roadmap requires a 120-card
   deck, so the *proposed* (unpublished) money split absorbs the 5 slots:
   **23 money cards — 6×1M, 5×2M, 4×3M, 3×4M, 2×5M, 2×10M, 1×15M** (card-list
   proposed 4/4/3/2/2/2/1 = 18). All action/property/rent counts match
   card-list.md exactly. Flagged for playtest adjustment once a physical deck
   can be counted.

2. **Market Crash vs complete sets.** Card-list rules note 3 says complete
   sets are stealable "(Property Raid; Market Crash per [GH])", so Market
   Crash may take a card **out of a complete set** (unlike classic single
   steals). If the theft empties a set, an attached shack is orphaned and
   discarded to the deck.

3. **Property Raid takes whole sets.** "Steal ALL properties of a colour" is
   implemented as transferring every set whose color matches, as intact units
   (shack included). The attacker may transiently hold two incomplete sets of
   one color; `PropertySets.Valid()` is only enforced at their own turn end,
   and `RearrangeProperty` merges.

4. **Shack.** Modeled as a dedicated `PropertySet.Shack *Card` attachment
   (not a member of `Cards` as classic house/hotel were): max 1 per set, any
   set (incomplete, rails, utils), +`Settings.ShackRentBonus` (5M) rent.
   It stays on the table when played (never returned to the deck — the
   classic card-duplication lesson), is **immovable and never payable**
   (excluded from `getPayableCards`), travels with the set on Set Snatcher /
   Property Raid, and is discarded to the deck the moment its set empties
   (payment break-up, Market Crash, Repo Man, rearrange).

5. **Debt chips.** `Game.DebtChips` (available pool per player, start 3) +
   `Game.Debts` (`[]DebtObligation{ID, DebtorID, CreditorID}`), both in the
   snapshot. A chip is issued when a payment compliance surrenders everything
   payable and still falls short, or when a Debt Trap compliance resolves.
   With zero chips available the shortfall is written off (no obligation).
   Settling returns the chip to the debtor's pool. While the current player
   has outstanding debts, **every** play/rearrange/discard/turn-end is
   blocked (`OutstandingDebtExists`) until each chip is settled via
   `SettleDebt` (any one hand card; consumes one of the 3 plays). The
   creditor receives the settlement card "as if played": property → new set
   on their table, money/action → their bank. If the debtor's hand is empty
   (deck exhausted) the remaining debts are forgiven at `CompleteTurn`.
   NAH! can deny the Debt Trap *demand*, but never debt settlement itself.

6. **Rent has no pending phase.** Classic's `PendingRent` existed for the
   double-the-rent modifier, which No Mercy removes. `PlayRent` therefore
   creates payment demands immediately — one per opponent, always all
   opponents. Double Rent is a standalone card handled by the same method
   (multiplier 2 from the asset key). Charging a color you have no set in is
   allowed (0-rent demands, matching classic's tolerance); pair cards reject
   colors outside their pair. A 0-shortfall never issues a chip.

7. **Go Again = last play.** `PlayGoAgain` sets `MovesLeft = 0` ("play as
   your last card") and queues one extra turn (`GoAgainQueued`, snapshotted).
   Hand-limit discard still applies before `CompleteTurn`; the same player
   then starts a fresh full turn (draws 2 — or 5 on an empty hand — and gets
   3 plays). Chaining is possible across turns but not within one turn.
   Like any action card it may instead be banked as 1M via `PlayMoney`.

8. **Multi-target picks are mandatory and total.** Heist / Market Crash
   `picks` must cover exactly the opponents that have at least one bank card
   / property; with zero eligible opponents the play is rejected
   (`NoValidTargets`) rather than wasted. Property Raid likewise requires at
   least one opponent holding the named color. Each target still gets an
   individual NAH!-deniable demand.

9. **Pickpocket** carries the named category (`StealCategory`
   property/money/action) in the demand; compliance moves every matching
   card from the target's **hand to the attacker's hand**. An empty or
   non-matching hand complies with zero cards (the attacker cannot see the
   hand, so the play is never rejected for emptiness).

10. **Repo Man / Tax Day** are target-choice compliances:
    `ComplyRepoManDemand(keepCardID, distribution)` /
    `ComplyTaxDayDemand(...)`. The distribution must cover every other
    property/bank card exactly once; recipients may be any player except the
    target (the attacker included). Distributed properties arrive as new
    single-card sets; distributed bank cards go to recipients' banks.

11. **Debt Trap requires an available chip at play time**
    (`NoDebtChipsAvailable` otherwise); if the target somehow has none left
    at comply time the trap fizzles (comply succeeds, no chip).

12. **Bank Swap** is rejected when the attacker's own bank is empty; an
    empty target bank is allowed (the card only forbids the former).

13. **Transfers place properties into new sets.** Payment / steal receivers
    get each property as its own new set (classic behavior) and merge later
    via `RearrangeProperty`. Payments may break complete sets (everything on
    the table is payable, as in classic).

## Public API surface (for Phase 2/3 proto work)

Constructors / persistence:
- `NewGame(cfg Settings, playerIDs []uuid.UUID) *Game`
- `DefaultSettings() Settings`, `(*Settings).Encode/Decode`
- `(*Game).Snapshot() (GameSnapshot, error)`, `NewGameFromSnapshot`,
  `(*Game).EncodeMsgpack`, `DecodeMsgpack`

Turn lifecycle:
- `StartTurn(playerID)` → `*ActionStartTurn` (includes outstanding debts +
  go-again flag)
- `CompleteTurn(playerID)` → `*ActionStartTurn` (advances or repeats on
  go-again; enforces demands/debts/hand limit/valid sets)
- `DiscardCards(playerID, cardIDs...)` → `*ActionDiscardCards`
- `RearrangeProperty(playerID, cardID, targetSetID?, activeColor?)` →
  `*ActionRearrangeCard`

Plays (all gated on turn, moves, no demands, no outstanding debts):
- `PlayMoney(playerID, cardID)` → `*ActionPlayMoney`
- `PlayProperty(playerID, cardID, propSetID?, activeColor?)` → `*ActionPlayProperty`
- `PlayShack(playerID, cardID, propSetID)` → `*ActionPlayShack`
- `PlayBigPayday(playerID, cardID)` → `*ActionPlayBigPayday`
- `PlayGoAgain(playerID, cardID)` → `*ActionPlayGoAgain`
- `PlaySetSnatcher(playerID, targetID, cardID, setID)` → `*ActionDemandsCreated`
- `PlayDebtTrap(playerID, targetID, cardID)` → `*ActionDemandsCreated`
- `PlayHeist(playerID, cardID, picks map[uuid.UUID]Identifier)` → `*ActionDemandsCreated`
- `PlayMarketCrash(playerID, cardID, picks map[uuid.UUID]Identifier)` → `*ActionDemandsCreated`
- `PlayRepoMan(playerID, targetID, cardID)` → `*ActionDemandsCreated`
- `PlayPropertyRaid(playerID, cardID, color)` → `*ActionDemandsCreated`
- `PlayTaxDay(playerID, targetID, cardID)` → `*ActionDemandsCreated`
- `PlayPickpocket(playerID, targetID, cardID, category StealCategory)` → `*ActionDemandsCreated`
- `PlayBankSwap(playerID, targetID, cardID)` → `*ActionDemandsCreated`
- `PlayYoink(playerID, targetID, cardID)` → `*ActionDemandsCreated`
- `PlayRent(playerID, cardID, color)` → `*ActionDemandsCreated`
  (handles two-color rent, rent-any-color, and both double rent variants)

Demand resolution (each returns `*ActionDemandComplied`; an inactive/denied
demand complies as a dismissal):
- `ComplyPaymentDemand(playerID, demandID, cardIDs...)` (chip on shortfall)
- `ComplyPropertyDemand(playerID, demandID)` (market crash)
- `ComplyPropertySetDemand(playerID, demandID)` (set snatcher)
- `ComplyColorPropertiesDemand(playerID, demandID)` (property raid)
- `ComplyBankCardDemand(playerID, demandID)` (heist)
- `ComplyBankSwapDemand(playerID, demandID)`
- `ComplyDebtTrapDemand(playerID, demandID)`
- `ComplyRepoManDemand(playerID, demandID, keepCardID, distribution)`
- `ComplyTaxDayDemand(playerID, demandID, keepCardID, distribution)`
- `ComplyPickpocketDemand(playerID, demandID)`
- `DenyDemand(playerID, demandID, cardID)` → `*ActionDemandsCreated` (NAH!)

Debt settlement:
- `SettleDebt(playerID, debtID, cardID)` → `*ActionDebtSettled`
- `PlayerDebts(playerID) []DebtObligation`, `CountAvailableDebtChips`

Timeout defaults:
- `DefaultMove(playerID)` → `[]Action` (settles debts, merges invalid sets,
  discards to hand limit)
- `DefaultDemand(playerID, demandID)` → `Action` (best-payment subset for
  payments; keep-highest-value + give-rest-to-attacker for repo man/tax day;
  straight comply otherwise)

Queries: `CheckWinConditions`, `CountMoney`, `CountCompletedSets`, `CountHands`.

Demand kinds: `Payment`, `Property`, `PropertySet`, `ColorProperties`,
`BankCard`, `BankSwap`, `DebtTrap`, `RepoMan`, `TaxDay`, `Pickpocket`.
Action kinds: `play_money`, `play_property`, `play_shack`,
`play_big_payday`, `play_go_again`, `demand_created`, `demand_complied`,
`debt_settled`, `discard_cards`, `start_turn`, `rearrange_card`.
`ActionDemandComplied` carries one of: `TransferCards` (with optional
`Debt`), `TransferProperty`, `TransferPropertySets`, `TransferBankSwap`,
`TransferDistribution`, `TransferPickpocket`, `TransferDebtChip` — these are
the shapes Phase 2 will mirror as proto messages (with masked variants for
hidden information, e.g. pickpocket hand contents).

## Notes for Phase 2/3

- `Settings` mirrors classic's msgpack+validate style; all rule knobs
  (debt chips, yoink amount, shack bonus, big-payday target, per-card deck
  amounts) are configurable, defaults reproduce the official 120-card game.
- The deck doubles as the discard pile exactly as in classic; the snapshot
  preserves deck order, the `IdentifierGenerator` counter, debt state and
  the go-again flag (`TestSnapshotRoundTrip`).
- `DemandSource` distinguishes every card for UI/history rendering.
- Timeout scheduling, masking, and default-move wiring belong to the Phase 3
  controller; `DefaultMove`/`DefaultDemand` are ready for it.
