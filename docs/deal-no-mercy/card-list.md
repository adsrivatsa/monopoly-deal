# Deal No Mercy — Card List & Asset Plan

Target game: **Monopoly Deal: No Mercy** (Hasbro G3071, Walmart-exclusive late 2025, wide
release 2026; 2–5 players). Deck: **120 cards + 15 debt chips** (3 per player) + game guide.

Sources (see "Confidence" column; full URLs at the bottom):
counts and rules cross-checked between a written rules guide [GH], three video tutorials with
card close-ups [MW][GR][TS], Hasbro's official product photos [AZ], and retail listings [SM].
The component list in [GH] sums exactly to the advertised 120 cards.

Rules deltas that matter for assets: debt chips are a physical component (needs a UI asset);
Double The Rent is replaced by standalone Double Rent cards; House/Hotel are replaced by
Shack; Deal Breaker drops from 5M to 3M (needs a re-issued card face); a new 15M money
denomination exists; Rent-Any-Color now charges **all** players (classic `rent_wild` face says
"Force one player…", so it cannot be reused).

Naming follows this repo's re-theme idiom (classic set renamed Hasbro cards: Pass Go→PAYDAY,
Just Say No→NAH!, Sly Deal→PROPERTY STEAL, Deal Breaker→SET SNATCHER, …). Proposed re-themed
names below are open for review.

---

## (a) Reused from classic Deal

These faces are identical to the classic set, but each game's asset directory is
self-contained: the files below are duplicated into `deal-no-mercy/card/` rather
than served from `monopoly-deal/card/`.

| Card | Count | Value | Classic asset (large / small) | Notes | Confidence |
|---|---|---|---|---|---|
| Properties, all 10 colours | 28 | 1M–4M | `brown1`…`rail4` / colour tiles | Same 28-card roster, same set sizes, values and rent tables as classic | CONFIRMED [GH][MW] |
| Two-colour wild properties | 9 | 1M–4M | `brown_sky`…`util_rail` / same | Same 7 colour pairs; assumed same 9-card spread as classic (2× pink/orange, 2× red/yellow, 1× each other) | counts CONFIRMED, exact spread PROPOSED [GH] |
| Any-colour wild property | 2 | 0 | `wild` / `wild` | 2 copies (classic had different count; asset identical) | CONFIRMED [GH][AZ] |
| Money 1M 2M 3M 4M 5M 10M | 17 of 18 | — | `1`…`10` / same | 18 money cards total incl. new 15M; per-denomination split unpublished | denoms CONFIRMED, split PROPOSED [MW] |
| Rent (two-colour) | 5 | 1M | `rent_brown_sky` … `rent_util_rail` / money tile `1` | "Collect rent from each player" — matches classic face text ("All players pay you rent…") | CONFIRMED; 5-pair spread PROPOSED (classic five pairs) [GH][MW] |
| Just Say No → **NAH!** | 5 | 4M | `nah` / money tile `4` | Same card, now 5 copies | CONFIRMED [GH] |

Classic cards **absent** from No Mercy (do not ship in this set): Sly Deal, Forced Deal,
Debt Collector, It's My Birthday, House, Hotel, Double The Rent (modifier), Pass Go,
classic 5M Deal Breaker face, classic one-player `rent_wild`. CONFIRMED [GH][MW][TS].

## (b) New to No Mercy — assets authored under `src/backend/public/deal-no-mercy/card/`

All action-card faces use the classic action layout; background tint = tint of the card's
money value (see style guide §5). New 15M tint introduced: `#b8ebc6` (pale mint).

| Official card | Proposed key (this repo) | Display title | Category | Count | Value | Effect summary | Confidence |
|---|---|---|---|---|---|---|---|
| Money 15M | `15` | 15M | money | 1 | 15M | New denomination | CONFIRMED [MW][AZ]; count of 1 PROPOSED |
| Deal Breaker (3M) | `set_snatcher` | SET SNATCHER | action | 3 | 3M | Steal a complete set incl. buildings (re-issued face: 5M→3M) | CONFIRMED [GH][MW] |
| Forced Debt | `debt_trap` | DEBT TRAP | action | 3 | 2M | Take a debt chip from any player; they must settle up with you next turn | CONFIRMED [GH][MW] |
| Go Again | `go_again` | GO AGAIN! | action | 3 | 1M | Play as your last card, end turn, take another turn | CONFIRMED [GH][MW] |
| Heist | `heist` | HEIST | action | 3 | 4M | Steal 1 chosen card from every player's bank | CONFIRMED [GH][MW] |
| Market Crash | `market_crash` | MARKET CRASH | action | 3 | 4M | Steal 1 chosen property from every player | CONFIRMED [GH][MW] |
| Pass Go, Go, Go! | `big_payday` | BIG PAYDAY | action | 4 | 1M | Draw until 7 cards in hand | CONFIRMED [GH][MW] |
| Repossession | `repo_man` | REPO MAN | action | 2 | 5M | Target keeps 1 property, gives the rest away (their choice of recipients) | CONFIRMED [GH][MW] |
| Shack | `shack` | SHACK | action | 3 | 3M | +5M rent on any set — incomplete sets, railroads, utilities OK; max 1 per set | CONFIRMED [GH][MW] |
| Super Sly Deal | `property_raid` | PROPERTY RAID | action | 3 | 3M | Pick a colour; steal all properties of that colour from everyone, full sets included | CONFIRMED [GH][MW] |
| Tax Collector | `tax_day` | TAX DAY | action | 2 | 5M | Target keeps 1 bank card, gives the rest away (their choice of recipients) | CONFIRMED [GH][MW] |
| Tough Luck | `pickpocket` | PICKPOCKET | action | 3 | 5M | Name property/money/action; steal all of that type from one player's hand | CONFIRMED [GH][MW] |
| Unfair Trade | `bank_swap` | BANK SWAP | action | 2 | 4M | Swap entire banks with another player (not if yours is empty) | CONFIRMED [GH][MW] |
| Yoink! | `yoink` | YOINK! | action | 3 | 2M | Collect 10M from any player, no change given | CONFIRMED [GH][MW] |
| Rent Any Color | `rent_wild` | RENT | action (rent) | 3 | 3M | **All** players pay rent in a colour of your choice (classic face targets one player → new asset) | CONFIRMED [GH][MW] |
| Double Rent (two-colour) | `double_rent_brown_sky`, `double_rent_pink_orange`, `double_rent_red_yellow`, `double_rent_green_blue`, `double_rent_util_rail` | DOUBLE RENT | action (rent) | 5 (1 each) | 1M | All players pay double rent in one of the two colours | CONFIRMED [GH][MW]; pair spread PROPOSED (classic five pairs) |
| Double Rent Any Color | `double_rent_wild` | DOUBLE RENT | action (rent) | 3 | 3M | All players pay double rent in a colour of your choice | CONFIRMED [GH][MW] |
| Debt chip | `debt_chip` | DEBT | UI asset | 15 (3/player) | — | Given when a player can't pay in full; repaid with any 1 hand card at start of next turn (consumes a play) | CONFIRMED [GH][AZ][SM] |
| Reference card | — | — | — | 5 | — | Not needed as a card asset (rules belong in app UI) | CONFIRMED [GH]; out of scope |

### Proposed money split (18 cards, PROPOSED — no published breakdown)

4× 1M, 4× 2M, 3× 3M, 2× 4M, 2× 5M, 2× 10M, 1× 15M. Flag for playtest adjustment once a
physical deck can be counted.

### Proposed action→small-tile mapping (extends `actionMoneyMap` convention)

Actions have no dedicated small tiles; each maps to its money value's tile, exactly like the
classic set. Two genuinely new small tiles are needed: `15` and `debt_chip`.

| Action key | Small tile |
|---|---|
| `big_payday`, `go_again`, `double_rent_brown_sky/pink_orange/red_yellow/green_blue/util_rail` | `1` |
| `debt_trap`, `yoink` | `2` |
| `set_snatcher`, `shack`, `property_raid`, `rent_wild`, `double_rent_wild` | `3` |
| `heist`, `market_crash`, `bank_swap`, (`nah` — already mapped) | `4` |
| `repo_man`, `tax_day`, `pickpocket` | `5` |

## Rules differences (for the future engine phase — recorded here for context)

Unchanged: draw 2 (5 on empty hand), 3 plays/turn, 7-card hand limit, 3 different complete
sets to win, payments from table only, no change given.

1. **Debt chips**: can't pay in full → hand over everything you can + one of your 3 chips.
   Next turn, after drawing and before your own plays, settle each chip with any 1 hand card
   (receiver plays it immediately as if their own); each settlement consumes one of your 3
   plays. Out of chips → further debt forgiven. Just Say No can't prevent debt.
2. Double Rent is a standalone play, not a rent modifier.
3. Complete sets are stealable (Property Raid; Market Crash per [GH]).
4. Shack: +5M rent, one per set, placeable on incomplete sets/rails/utils, immovable, travels
   with the property.
5. NAH! ×5; blocks Debt Trap but not debt itself.

## Sources

- [GH] https://www.geekyhobbies.com/monopoly-deal-no-mercy-rules/ — full rules + component counts (sums to 120)
- [MW] https://www.youtube.com/watch?v=jM4i_ZZBbY4 — tutorial with card close-ups (values/texts)
- [GR] https://www.youtube.com/watch?v=YeTFoRNSaEs — Game Rules tutorial (cross-check)
- [TS] https://www.youtube.com/watch?v=FkpUyr41XU0 — Triple S Games (cross-check)
- [AZ] https://www.amazon.com/Monopoly-Deal-Mercy-Card-Game/dp/B0GGV66P58 — official product photos (box back: "120 Cards, 15 Debt Chips")
- [SM] https://shop.shopofmagic.com/products/monopoly-deal-no-mercy — retail listing (chip colours)
- https://www.meeplemountain.com/reviews/monopoly-deal-no-mercy/ — review corroboration
- https://boardgamegeek.com/boardgame/465961/monopoly-deal-no-mercy — BGG entry
