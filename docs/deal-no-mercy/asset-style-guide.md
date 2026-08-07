# The Deal — Card Asset Style Guide

Extracted from the classic-Deal assets in `src/backend/public/card/{large,small}/` (58 large,
24 small SVGs). This document is the reference for authoring the Deal No Mercy asset set under
`src/backend/public/deal-no-mercy/card/`.

## 1. Canvas & frame (both sizes)

Every SVG — large card faces **and** small tiles — uses the same canvas:

```
viewBox="0 0 282.96 461.03998"   (poker-card ratio, ~1:1.63)
width="378" height="605"          (design-tool export attributes, harmless)
```

Small tiles are *not* a different canvas; they are simplified full-card compositions on the
same viewBox (the frontend just renders them smaller).

Frame construction (large cards):

| Element | Geometry | Style |
|---|---|---|
| Page background | full-bleed rect (0,0 → 283×453; exports leave ~8px dead space at the bottom) | `#ffffff` |
| Card body | rect x=14.4 y=14.1 w=254.6 h=424.6, rx≈4.5 | fill = family colour (money/action) or white (property/rent); stroke `#000` ~3 |
| Pinstripe inset frame | 3 nested rounded rects at insets ≈28.6 / 31.1 / 33.9 from the page edge, rx≈4.5 | stroke `#000` ~3, fill none — **money and action cards only** |
| Value badge | white square x=28.5 y=28.3 w=57 h=57, rx≈4.5 | fill `#fff`, stroke `#000` ~3; sits on top of the pinstripes / property band so it "notches" the top-left corner |

The value badge text is the money value + `M` (e.g. `1M`, `5M`), black, bold, centred,
cap-height ≈ 21px (font-size ≈ 30).

## 2. Typography

The existing exports contain **no `<text>` elements** — all lettering was converted to outline
paths by the design tool. The typeface is clearly **JetBrains Mono Bold** (it matches the
app-wide font loaded in `src/frontend/src/index.css`).

**Deliberate deviation for the No Mercy set:** new assets are hand-authored SVG using real
`<text>` elements with:

```
font-family="'JetBrains Mono','SF Mono',Menlo,Monaco,Consolas,monospace"
font-weight="700"
```

Rationale: reviewable diffs, editable copy, tiny files. Risk: text rendering depends on the
viewer having JetBrains Mono (the app always loads it via Google Fonts, so in-app rendering is
faithful; a bare file manager preview may fall back to another monospace). If pixel-identical
parity is ever required, run the finished files through an outline pass — the geometry is
already final.

Reference sizes (in viewBox units):

| Text | Size / weight | Notes |
|---|---|---|
| Value badge (`1M`) | 30 bold | centred in 57×57 badge |
| Action/rent title in circle | 30 bold | centred, letter-spacing ~0, may wrap to 2 lines (line gap ≈ 36) |
| Money denomination (`5M` big) | 110 bold | centred in circle |
| Effect / body text | 17 bold | centred, line height ≈ 24, wraps at ~20 chars |
| Property band title | 26 bold, white | rotated -90°, reads bottom-to-top, centred in band |
| Rent-table caption ("Properties owned in set") | 12–13 bold | |

## 3. Card family layouts (large)

### 3.1 Money (`1.svg` … `10.svg`)
- Card body filled with the denomination colour, triple pinstripe frame.
- Large circle (r ≈ 100) centred at (141.7, ~226), stroke `#000` ~3, **fill = same body colour**
  (the circle reads as an outline on the coloured field).
- Denomination `NM` huge in the circle centre. Badge top-left repeats the value.

### 3.2 Action (`payday.svg`, `nah.svg`, `property_steal.svg`, …)
- Card body filled with a pale family colour (see palette), triple pinstripe frame.
- **White** circle, stroke `#000` ~3, r ≈ 100.4, centred at (141.7, 185.3) — i.e. upper half.
- Title inside the circle, ALL CAPS (1–2 lines).
- Effect text below the circle (y ≈ 310–420), centred, wrapped.

### 3.3 Rent (`rent_brown_sky.svg`, …, `rent_wild.svg`)
- White card body, **no pinstripes**.
- Circle in the upper half as actions, but with a thick ring (donut, ring width ≈ 22) split
  horizontally: top half = colour A, bottom half = colour B. `rent_wild` uses a multi-segment
  rainbow ring. Inner disc white with `RENT` title.
- Effect text below, e.g. "All players pay you rent for properties you own in one of these
  colors."

### 3.4 Pure property (`brown1.svg`, …)
- White card body, no pinstripes.
- Vertical colour band: rect x=14.9 y=28.3 w=56.6 h=410.2, rx≈4.5, family colour, black stroke.
  The badge covers its top 57px.
- Band title: white bold caps, rotated -90° (this repo uses its own street names, e.g. brown =
  "SKID ROW").
- Right region: rent table — caption "Properties owned in set", one small house-tile glyph per
  set size (numbered 1..N), dotted leader lines down to the rent values (`1M 2M …`), and the
  word `RENT` beneath.

### 3.5 Two-colour wild property (`brown_sky.svg`, …)
- Like a property card, but the band is split into two halves (top = colour A with rotated
  "WILD", bottom = colour B with rotated "CARD") and the right region stacks **two** rent
  tables, one per colour.

### 3.6 Wild-wild (`wild.svg`)
- Band is a stack of 10 colour stripes (rail black, util `#d2dbb2`, blue, green, yellow, red,
  orange, pink, sky, brown) with a white rotated label "WILD PROPERTY".
- Body text: "This card can take any color. Add it to any set or create a new one!". Value 0 —
  **no badge**.

## 4. Small tiles

Same viewBox; radically simplified; **no body/effect text**:

- **Property colour tile** (`brown.svg` …): white card, badge top-left, vertical colour band
  x≈14 y≈14 w≈100 h≈425 (full height, square corners inside a black ~9 stroke). Rest white.
- **Two-colour wild tile** (`brown_sky.svg` …): same band split top/bottom into the two colours.
- **Wild-wild tile** (`wild.svg`): band = 10 rainbow stripes, no badge.
- **Money tile** (`1.svg` … `10.svg`): full body filled with denomination colour, **triple**
  concentric rounded-rect strokes (insets ≈14 / 21 / 28), badge top-left with `NM`.
- **Actions have no small tiles.** The backend maps each action to a money-denomination tile
  via `actionMoneyMap` in `src/backend/internal/service/monopoly-deal/card.go`
  (e.g. Deal Breaker→`5.svg`, Just Say No→`4.svg`, rents→`1.svg`).

## 5. Palette

Property / family colours (used at full strength on bands, wild stripes, rent rings):

| Family | Hex |
|---|---|
| Brown | `#643232` |
| Sky | `#c2dcff` |
| Pink | `#d43790` |
| Orange | `#ff914d` |
| Red | `#ff3131` |
| Yellow | `#ffff00` |
| Green | `#247531` |
| Blue (dark) | `#1800ad` |
| Utility | `#d2dbb2` |
| Railroad | `#000000` |
| Ink / strokes | `#000000` |
| Paper | `#ffffff` |

Money / action background tints (pale washes; each action card borrows the tint of a money
denomination — by design the action's tint matches its money value):

| Value | Hex | Used by (large) |
|---|---|---|
| 1M | `#fff9c7` (pale yellow) | `1.svg`, payday, double_the_rent, all two-colour rents |
| 2M | `#ffd9c2` (pale peach) | `2.svg`, party_bill |
| 3M | `#d2dbb2` (pale olive) | `3.svg`, property_steal, property_swap, settle_up, house, rent_wild(ring uses rainbow) |
| 4M | `#c2dcff` (pale blue) | `4.svg`, nah, hotel |
| 5M | `#bea1f7` (lavender) | `5.svg`, set_snatcher |
| 10M | `#ffbd59` (amber) | `10.svg` |

Note the deliberate collisions: 3M tint = utility colour, 4M tint = sky colour. Keep the rule
**action background tint = tint of its money value** for all new No Mercy actions.

## 6. Naming & this repo's re-theme

The repo does not use Hasbro card names. Existing mapping (see `largeCardMap` in
`src/backend/internal/service/monopoly-deal/card.go`):

| Official | This repo (file / display name) |
|---|---|
| Pass Go | `payday` / PAYDAY |
| Just Say No | `nah` / NAH! |
| Sly Deal | `property_steal` / PROPERTY STEAL |
| Forced Deal | `property_swap` / PROPERTY SWAP |
| Deal Breaker | `set_snatcher` / SET SNATCHER |
| Debt Collector | `settle_up` / SETTLE UP |
| It's My Birthday | `party_bill` / PARTY BILL |
| House / Hotel / Double The Rent | same names |

New No Mercy cards follow the same idiom: snake_case file names, punchy re-themed ALL-CAPS
display titles, official name recorded in `docs/deal-no-mercy/card-list.md`.

## 7. Authoring rules for new No Mercy assets

1. Same viewBox `0 0 282.96 461.03998`; content laid out inside 283×453.
2. Reuse the exact frame recipe from §1 and the family layout from §3/§4.
3. Hand-written SVG: `<rect>`, `<circle>`, `<path>` for small glyphs, `<text>`/`<tspan>` for
   copy. No design-tool clip-path pyramids, no outlined text.
4. Stroke widths: 3 for frames/circles/badges (matches the 4 × 0.748 export scale), ~9 for the
   small-tile band outline.
5. New action cards: pick the pale tint by money value (§5). If No Mercy introduces values
   without a tint, extend the tint table here first.
6. Files live in `src/backend/public/deal-no-mercy/card/{large,small}/` — never overwrite the
   classic set.
