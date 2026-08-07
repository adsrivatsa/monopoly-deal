// Optimistic action reducer for "Monopoly Deal: No Mercy".
//
// Unlike a full-snapshot protocol, the No Mercy service broadcasts a single
// Action per play (see service/deal-no-mercy/event.go PublishAction). The full
// authoritative GameState is only sent on connect/reconnect. So the page must
// fan each incoming Action out into local GameState mutations, otherwise the
// board (which renders purely from GameState) would freeze after the initial
// snapshot.
//
// applyActionToGameState is a PURE function: given the previous GameState and
// one Action, it returns the next GameState. It is intentionally conservative —
// every branch is defensive and returns the input unchanged when it cannot
// safely apply a mutation, because the next reconnect snapshot is always
// authoritative and will heal any drift. Masked action variants (another
// player's private info) only move public counts, never card identities.
//
// The reducer never mutates its inputs; it produces new arrays/objects for the
// slices it touches and shares references for the rest.
import {
  ActionKind,
  type Action,
  type Card,
  type DebtChip,
  type Demand,
  type GameState,
  type Money,
  type Player,
  type PropertySet,
} from "../../../generated/deal_no_mercy";

// Whether this action moved a card the current player could see leave their own
// hand. Used to keep yourHand in sync when the self player is the actor.
const removeFromYourHand = (
  state: GameState,
  cardIds: Set<string>,
): GameState["yourHand"] => {
  const hand = state.yourHand;
  if (!hand) {
    return hand;
  }
  const nextCards = hand.cards.filter((card) => !cardIds.has(card.cardId));
  if (nextCards.length === hand.cards.length) {
    return hand;
  }
  return { ...hand, cards: nextCards };
};

const upsertMoney = (money: Money[], playerId: string, cards: Card[]): Money[] => {
  if (cards.length === 0) {
    return money;
  }
  const index = money.findIndex((pile) => pile.playerId === playerId);
  if (index === -1) {
    return [...money, { playerId, cards: [...cards] }];
  }
  const next = [...money];
  next[index] = { ...next[index], cards: [...next[index].cards, ...cards] };
  return next;
};

// Remove specific card ids from a player's bank pile.
const removeFromMoney = (
  money: Money[],
  playerId: string,
  cardIds: Set<string>,
): Money[] => {
  const index = money.findIndex((pile) => pile.playerId === playerId);
  if (index === -1) {
    return money;
  }
  const filtered = money[index].cards.filter((card) => !cardIds.has(card.cardId));
  if (filtered.length === money[index].cards.length) {
    return money;
  }
  const next = [...money];
  next[index] = { ...next[index], cards: filtered };
  return next;
};

// Merge/replace a property set on the owner. Sets are keyed by propertySetId;
// a new set is appended, an existing one is replaced (the server sends the
// authoritative post-play set contents).
const upsertPropertySet = (
  properties: PropertySet[],
  set: PropertySet,
): PropertySet[] => {
  const index = properties.findIndex(
    (existing) => existing.propertySetId === set.propertySetId,
  );
  if (index === -1) {
    return [...properties, set];
  }
  const next = [...properties];
  next[index] = set;
  return next;
};

const removePropertySets = (
  properties: PropertySet[],
  setIds: Set<string>,
): PropertySet[] => properties.filter((set) => !setIds.has(set.propertySetId));

const patchPlayer = (
  players: Player[],
  playerId: string,
  patch: (player: Player) => Player,
): Player[] =>
  players.map((player) => (player.playerId === playerId ? patch(player) : player));

const recountCompletedSets = (properties: PropertySet[], playerId: string): number => {
  // Best-effort: the authoritative snapshot corrects this on reconnect. We use
  // the same minimum-count rule the board uses.
  const minFor = (color: number): number => {
    switch (color) {
      case 1: // brown
      case 8: // blue
      case 9: // utility
        return 2;
      case 2: // sky
      case 3: // pink
      case 4: // orange
      case 5: // red
      case 6: // yellow
      case 7: // green
        return 3;
      case 10: // railroad
        return 4;
      default:
        return Number.POSITIVE_INFINITY;
    }
  };
  let count = 0;
  for (const set of properties) {
    if (set.playerId !== playerId) {
      continue;
    }
    const propertyCards = set.cards.filter(
      (card) => card.category === 3 || card.category === 4,
    ).length;
    const required = minFor(set.color);
    if (Number.isFinite(required) && propertyCards >= required) {
      count += 1;
    }
  }
  return count;
};

const sumMoney = (money: Money[], playerId: string): number => {
  const pile = money.find((entry) => entry.playerId === playerId);
  if (!pile) {
    return 0;
  }
  return pile.cards.reduce((total, card) => total + (card.value ?? 0), 0);
};

// After property/money changes, resync a player's money + completedSets stats
// so the roster (which the board and sidebar read) stays consistent without a
// snapshot.
const resyncPlayerStats = (
  players: Player[],
  properties: PropertySet[],
  money: Money[],
  playerId: string,
): Player[] =>
  patchPlayer(players, playerId, (player) => ({
    ...player,
    money: sumMoney(money, playerId),
    completedSets: recountCompletedSets(properties, playerId),
  }));

export type ApplyActionResult = {
  state: GameState;
  // demand ids created by this action (so the page can start their timers).
  createdDemandIds: string[];
};

export const applyActionToGameState = (
  state: GameState,
  action: Action,
): ApplyActionResult => {
  const actorId = action.playerId;
  let next = state;
  const createdDemandIds: string[] = [];

  const withSeq = (s: GameState): GameState =>
    action.seqNum && action.seqNum > s.seqNum ? { ...s, seqNum: action.seqNum } : s;

  switch (action.kind) {
    case ActionKind.ACTION_KIND_PLAY_MONEY: {
      const card = action.actionPlayMoney?.card;
      if (!card || !actorId) {
        break;
      }
      const money = upsertMoney(next.money, actorId, [card]);
      const players = patchPlayer(next.players, actorId, (player) => ({
        ...player,
        money: player.money + (card.value ?? 0),
        handCards: Math.max(0, player.handCards - 1),
      }));
      next = {
        ...next,
        money,
        players,
        yourHand:
          actorId === selfId(next) ? removeFromYourHand(next, new Set([card.cardId])) : next.yourHand,
        movesLeft: isSelfActor(next, actorId) ? Math.max(0, next.movesLeft - 1) : next.movesLeft,
        lastAction: card,
      };
      break;
    }

    case ActionKind.ACTION_KIND_PLAY_PROPERTY: {
      const set = action.actionPlayProperty?.propertySet;
      if (!set || !actorId) {
        break;
      }
      const properties = upsertPropertySet(next.properties, set);
      // The played card is whichever card in the set is not already elsewhere;
      // we can't know its id reliably, so only touch the self hand by set diff.
      const beforeIds = new Set(
        next.properties
          .filter((existing) => existing.propertySetId === set.propertySetId)
          .flatMap((existing) => existing.cards.map((card) => card.cardId)),
      );
      const addedCardId = set.cards.find((card) => !beforeIds.has(card.cardId))?.cardId;
      const players = resyncPlayerStats(
        patchPlayer(next.players, actorId, (player) => ({
          ...player,
          handCards: Math.max(0, player.handCards - 1),
        })),
        properties,
        next.money,
        actorId,
      );
      next = {
        ...next,
        properties,
        players,
        yourHand:
          actorId === selfId(next) && addedCardId
            ? removeFromYourHand(next, new Set([addedCardId]))
            : next.yourHand,
        movesLeft: isSelfActor(next, actorId) ? Math.max(0, next.movesLeft - 1) : next.movesLeft,
      };
      break;
    }

    case ActionKind.ACTION_KIND_PLAY_SHACK: {
      const set = action.actionPlayShack?.propertySet;
      if (!set || !actorId) {
        break;
      }
      const properties = upsertPropertySet(next.properties, set);
      const shackCardId = set.shack?.cardId;
      const players = patchPlayer(next.players, actorId, (player) => ({
        ...player,
        handCards: Math.max(0, player.handCards - 1),
      }));
      next = {
        ...next,
        properties,
        players,
        yourHand:
          actorId === selfId(next) && shackCardId
            ? removeFromYourHand(next, new Set([shackCardId]))
            : next.yourHand,
        movesLeft: isSelfActor(next, actorId) ? Math.max(0, next.movesLeft - 1) : next.movesLeft,
      };
      break;
    }

    case ActionKind.ACTION_KIND_PLAY_BIG_PAYDAY: {
      // Full variant carries drawn cards for the self player; masked carries
      // only a count. Either way it consumes a move for the actor.
      const drawn = action.actionPlayBigPayday?.cards;
      const maskedCount = action.maskedActionPlayBigPayday?.numCards;
      let players = next.players;
      let yourHand = next.yourHand;
      if (drawn && actorId === selfId(next)) {
        yourHand = yourHand
          ? { ...yourHand, cards: [...yourHand.cards, ...drawn] }
          : { cards: [...drawn] };
        players = patchPlayer(players, actorId, (player) => ({
          ...player,
          handCards: player.handCards + drawn.length,
        }));
      } else if (typeof maskedCount === "number" && actorId) {
        players = patchPlayer(players, actorId, (player) => ({
          ...player,
          handCards: player.handCards + maskedCount,
        }));
      }
      next = {
        ...next,
        players,
        yourHand,
        movesLeft: isSelfActor(next, actorId) ? Math.max(0, next.movesLeft - 1) : next.movesLeft,
      };
      break;
    }

    case ActionKind.ACTION_KIND_PLAY_GO_AGAIN: {
      // GO AGAIN is played as the actor's last move and keeps their turn alive
      // by granting more moves. The action carries only lastPlayedCard (no move
      // count), so we conservatively leave movesLeft unchanged — the play would
      // otherwise drop it to 0 and freeze the board; the next authoritative
      // snapshot corrects the real allotment. It also leaves the actor's hand.
      const card = action.actionPlayGoAgain?.lastPlayedCard;
      next = {
        ...next,
        lastAction: card ?? next.lastAction,
        yourHand:
          actorId === selfId(next) && card
            ? removeFromYourHand(next, new Set([card.cardId]))
            : next.yourHand,
        players:
          actorId && card
            ? patchPlayer(next.players, actorId, (player) => ({
                ...player,
                handCards: Math.max(0, player.handCards - 1),
              }))
            : next.players,
      };
      break;
    }

    case ActionKind.ACTION_KIND_REARRANGE_CARD: {
      const set = action.actionRearrangeCard?.propertySet;
      if (!set || !actorId) {
        break;
      }
      // Rearrange moves ONE card from a source set into the target set the
      // server sends. We must first strip the moved card out of every OTHER
      // set the actor owns, otherwise it lives in both the source and the
      // target set (a phantom duplicate the next snapshot would silently heal).
      // Mirrors classic's rearrange handling: remove-from-source, drop now-empty
      // sets, then upsert the authoritative target set.
      const movedCardId = action.actionRearrangeCard?.card?.cardId;
      let properties = next.properties;
      if (movedCardId) {
        properties = properties.map((existing) => {
          if (
            existing.playerId !== actorId ||
            existing.propertySetId === set.propertySetId
          ) {
            return existing;
          }
          const filtered = existing.cards.filter(
            (card) => card.cardId !== movedCardId,
          );
          return filtered.length === existing.cards.length
            ? existing
            : { ...existing, cards: filtered };
        });
      }
      properties = upsertPropertySet(properties, set);
      // Rearrange can empty a source set; drop any now-empty sets for the actor.
      properties = properties.filter(
        (existing) =>
          existing.propertySetId === set.propertySetId ||
          existing.playerId !== actorId ||
          existing.cards.length > 0,
      );
      const players = resyncPlayerStats(next.players, properties, next.money, actorId);
      next = {
        ...next,
        properties,
        players,
        // Rearranging is FREE — the engine's RearrangeProperty never calls
        // CompleteMove, so we must NOT consume a move here.
        movesLeft: next.movesLeft,
      };
      break;
    }

    case ActionKind.ACTION_KIND_DEMAND_CREATED: {
      const created = action.actionDemandsCreated;
      if (!created) {
        break;
      }
      // The demanding play consumes a move for the actor. Add the new demands to
      // state.demands; drop a denied demand if present.
      let demands = next.demands;
      const deniedId = created.deniedDemand?.id;
      if (deniedId) {
        demands = demands.filter((demand) => demand.id !== deniedId);
      }
      const incoming: Demand[] = created.demands ?? [];
      const byId = new Map(demands.map((demand) => [demand.id, demand]));
      for (const demand of incoming) {
        byId.set(demand.id, demand);
        createdDemandIds.push(demand.id);
      }
      demands = Array.from(byId.values());
      next = {
        ...next,
        demands,
        lastAction: created.lastPlayedCard ?? next.lastAction,
        // A demand-creating card leaves the actor's hand.
        yourHand:
          actorId === selfId(next) && created.lastPlayedCard
            ? removeFromYourHand(next, new Set([created.lastPlayedCard.cardId]))
            : next.yourHand,
        players:
          actorId && created.lastPlayedCard
            ? patchPlayer(next.players, actorId, (player) => ({
                ...player,
                handCards: Math.max(0, player.handCards - 1),
              }))
            : next.players,
        movesLeft: isSelfActor(next, actorId) ? Math.max(0, next.movesLeft - 1) : next.movesLeft,
      };
      break;
    }

    case ActionKind.ACTION_KIND_DEMAND_COMPLIED: {
      next = applyDemandComplied(next, action);
      break;
    }

    case ActionKind.ACTION_KIND_DEBT_SETTLED: {
      const settled = action.actionDebtSettled;
      if (!settled?.debt) {
        break;
      }
      const debt = settled.debt;
      const debtChips = next.debtChips.filter((chip) => chip.id !== debt.id);
      let money = next.money;
      let properties = next.properties;
      let players = next.players;
      const card = settled.card;
      if (!settled.forgiven && card) {
        if (settled.propertySet) {
          properties = upsertPropertySet(properties, settled.propertySet);
        } else {
          money = upsertMoney(money, debt.creditorId, [card]);
        }
        // The debtor loses the card from hand.
        players = patchPlayer(players, debt.debtorId, (player) => ({
          ...player,
          handCards: Math.max(0, player.handCards - 1),
        }));
        if (debt.debtorId === selfId(next)) {
          next = { ...next, yourHand: removeFromYourHand(next, new Set([card.cardId])) };
        }
      }
      players = resyncPlayerStats(players, properties, money, debt.creditorId);
      next = {
        ...next,
        debtChips,
        money,
        properties,
        players,
        // Settling a debt consumes a move for the debtor: the engine's
        // SettleDebt runs checkMoves + CompleteMove. The debtor is always the
        // current player (settlement is gated by checkTurn), so decrement for
        // the self-actor. Without this the optimistic movesLeft is 1 too high
        // and a reconnect snapshot would "reset" it.
        movesLeft: isSelfActor(next, actorId) ? Math.max(0, next.movesLeft - 1) : next.movesLeft,
      };
      break;
    }

    case ActionKind.ACTION_KIND_DISCARD_CARDS: {
      const discarded = action.actionDiscardCards?.cards;
      const maskedCount = action.maskedActionDiscardCards?.numCards;
      if (discarded && actorId) {
        const ids = new Set(discarded.map((card) => card.cardId));
        next = {
          ...next,
          players: patchPlayer(next.players, actorId, (player) => ({
            ...player,
            handCards: Math.max(0, player.handCards - discarded.length),
          })),
          yourHand:
            actorId === selfId(next) ? removeFromYourHand(next, ids) : next.yourHand,
        };
      } else if (typeof maskedCount === "number" && actorId) {
        next = {
          ...next,
          players: patchPlayer(next.players, actorId, (player) => ({
            ...player,
            handCards: Math.max(0, player.handCards - maskedCount),
          })),
        };
      }
      break;
    }

    case ActionKind.ACTION_KIND_START_TURN: {
      next = applyStartTurn(next, action);
      break;
    }

    default:
      break;
  }

  return { state: withSeq(next), createdDemandIds };
};

const applyStartTurn = (state: GameState, action: Action): GameState => {
  const actorId = action.playerId;
  const full = action.actionStartTurn;
  const masked = action.maskedActionStartTurn;
  const movesLeft = full?.movesLeft ?? masked?.movesLeft;
  const debts = full?.debts ?? masked?.debts;

  let players = state.players;
  let yourHand = state.yourHand;

  if (full && actorId === selfId(state)) {
    yourHand = yourHand
      ? { ...yourHand, cards: [...yourHand.cards, ...full.cards] }
      : { cards: [...full.cards] };
    players = patchPlayer(players, actorId, (player) => ({
      ...player,
      handCards: player.handCards + full.cards.length,
    }));
  } else if (masked && actorId) {
    players = patchPlayer(players, actorId, (player) => ({
      ...player,
      handCards: player.handCards + masked.numCards,
    }));
  }

  // The whole outstanding-debt list for this player is authoritative in the
  // action; merge it in (replacing any prior chips owed by this debtor).
  let debtChips = state.debtChips;
  if (debts && actorId) {
    const others = debtChips.filter((chip) => chip.debtorId !== actorId);
    debtChips = [...others, ...debts];
  }

  return {
    ...state,
    players,
    yourHand,
    debtChips,
    currentPlayerId: actorId ?? state.currentPlayerId,
    movesLeft: typeof movesLeft === "number" ? movesLeft : state.movesLeft,
  };
};

const applyDemandComplied = (state: GameState, action: Action): GameState => {
  const complied = action.actionDemandComplied;
  if (!complied) {
    return state;
  }
  let next = { ...state };

  // Clear the resolved demand.
  next.demands = next.demands.filter((demand) => demand.id !== complied.demandId);

  const t = complied;

  // TransferCards: a payment. Cards land in the target's bank (money/action) or
  // as new sets (property). A debt chip may be issued on shortfall.
  if (t.transferCards) {
    const tc = t.transferCards;
    let money = next.money;
    let properties = next.properties;
    // Remove paid cards from the payer's bank/sets.
    const paidIds = new Set(tc.cards.map((card) => card.cardId));
    money = removeFromMoney(money, tc.sourceId, paidIds);
    properties = properties.map((set) => {
      if (set.playerId !== tc.sourceId) {
        return set;
      }
      const filtered = set.cards.filter((card) => !paidIds.has(card.cardId));
      return filtered.length === set.cards.length ? set : { ...set, cards: filtered };
    });
    // Add non-property cards to target bank.
    const bankCards = tc.cards.filter((card) => card.category !== 3 && card.category !== 4);
    money = upsertMoney(money, tc.targetId, bankCards);
    // Add property sets to target.
    for (const set of tc.propertySets) {
      properties = upsertPropertySet(properties, set);
    }
    let debtChips = next.debtChips;
    if (tc.debt) {
      debtChips = [...debtChips, tc.debt];
    }
    let players = resyncPlayerStats(next.players, properties, money, tc.sourceId);
    players = resyncPlayerStats(players, properties, money, tc.targetId);
    next = { ...next, money, properties, debtChips, players };
  }

  // TransferProperty: single property moved (MARKET CRASH).
  if (t.transferProperty) {
    const tp = t.transferProperty;
    let properties = next.properties;
    if (tp.card) {
      const movedId = tp.card.cardId;
      properties = properties
        .map((set) => {
          if (set.playerId !== tp.sourceId) {
            return set;
          }
          const filtered = set.cards.filter((card) => card.cardId !== movedId);
          return filtered.length === set.cards.length ? set : { ...set, cards: filtered };
        })
        .filter((set) => set.playerId !== tp.sourceId || set.cards.length > 0);
    }
    if (tp.propertySet) {
      properties = upsertPropertySet(properties, tp.propertySet);
    }
    let players = resyncPlayerStats(next.players, properties, next.money, tp.sourceId);
    players = resyncPlayerStats(players, properties, next.money, tp.targetId);
    next = { ...next, properties, players };
  }

  // TransferPropertySets: whole sets move (SET SNATCHER, PROPERTY RAID).
  if (t.transferPropertySets) {
    const ts = t.transferPropertySets;
    const movedIds = new Set(ts.propertySets.map((set) => set.propertySetId));
    let properties = removePropertySets(next.properties, movedIds);
    for (const set of ts.propertySets) {
      // Reassign ownership to the receiver.
      properties = upsertPropertySet(properties, { ...set, playerId: ts.targetId });
    }
    let players = resyncPlayerStats(next.players, properties, next.money, ts.sourceId);
    players = resyncPlayerStats(players, properties, next.money, ts.targetId);
    next = { ...next, properties, players };
  }

  // TransferBankSwap: both banks are replaced wholesale.
  if (t.transferBankSwap) {
    const tb = t.transferBankSwap;
    let money = next.money.filter(
      (pile) => pile.playerId !== tb.sourceId && pile.playerId !== tb.targetId,
    );
    money = [
      ...money,
      { playerId: tb.sourceId, cards: [...tb.sourceBank] },
      { playerId: tb.targetId, cards: [...tb.targetBank] },
    ];
    let players = resyncPlayerStats(next.players, next.properties, money, tb.sourceId);
    players = resyncPlayerStats(players, next.properties, money, tb.targetId);
    next = { ...next, money, players };
  }

  // TransferDistribution: REPO MAN / TAX DAY. Source kept one card; the rest
  // were handed out (to bank or as property). The distributed cards leave the
  // SOURCE's zones (RepoMan pulls from the source's property sets, TaxDay from
  // the source's bank), so we must strip them from the source before adding
  // them to recipients — otherwise each distributed card lives in both places
  // (a phantom duplicate the next snapshot would silently heal).
  if (t.transferDistribution) {
    const td = t.transferDistribution;
    let money = next.money;
    let properties = next.properties;
    const distributedIds = new Set(
      td.entries
        .map((entry) => entry.card?.cardId)
        .filter((id): id is string => Boolean(id)),
    );
    // Remove distributed cards from the source's bank and property sets, then
    // drop any now-empty source sets.
    if (distributedIds.size > 0) {
      money = removeFromMoney(money, td.sourceId, distributedIds);
      properties = properties
        .map((set) => {
          if (set.playerId !== td.sourceId) {
            return set;
          }
          const filtered = set.cards.filter((card) => !distributedIds.has(card.cardId));
          return filtered.length === set.cards.length ? set : { ...set, cards: filtered };
        })
        .filter((set) => set.playerId !== td.sourceId || set.cards.length > 0);
    }
    const players = new Set<string>();
    for (const entry of td.entries) {
      if (!entry.card) {
        continue;
      }
      if (entry.propertySet) {
        properties = upsertPropertySet(properties, entry.propertySet);
      } else {
        money = upsertMoney(money, entry.recipientId, [entry.card]);
      }
      players.add(entry.recipientId);
    }
    // The source lost cards too, so resync its stats as well.
    players.add(td.sourceId);
    let nextPlayers = next.players;
    for (const playerId of players) {
      nextPlayers = resyncPlayerStats(nextPlayers, properties, money, playerId);
    }
    next = { ...next, money, properties, players: nextPlayers };
  }

  // TransferPickpocket: hand cards move source -> target (self-visible).
  if (t.transferPickpocket) {
    const tp = t.transferPickpocket;
    let players = next.players;
    players = patchPlayer(players, tp.sourceId, (player) => ({
      ...player,
      handCards: Math.max(0, player.handCards - tp.cards.length),
    }));
    players = patchPlayer(players, tp.targetId, (player) => ({
      ...player,
      handCards: player.handCards + tp.cards.length,
    }));
    let yourHand = next.yourHand;
    const self = selfId(next);
    if (tp.sourceId === self) {
      const ids = new Set(tp.cards.map((card) => card.cardId));
      yourHand = removeFromYourHand(next, ids);
    } else if (tp.targetId === self && yourHand) {
      yourHand = { ...yourHand, cards: [...yourHand.cards, ...tp.cards] };
    }
    next = { ...next, players, yourHand };
  }

  // MaskedTransferPickpocket: only counts move (cards hidden).
  if (t.maskedTransferPickpocket) {
    const mp = t.maskedTransferPickpocket;
    let players = next.players;
    players = patchPlayer(players, mp.sourceId, (player) => ({
      ...player,
      handCards: Math.max(0, player.handCards - mp.numCards),
    }));
    players = patchPlayer(players, mp.targetId, (player) => ({
      ...player,
      handCards: player.handCards + mp.numCards,
    }));
    next = { ...next, players };
  }

  // TransferDebtChip: DEBT TRAP compliance issues an outstanding chip.
  if (t.transferDebtChip?.debt) {
    next = { ...next, debtChips: [...next.debtChips, t.transferDebtChip.debt] };
  }

  return next;
};

// selfId is threaded through GameState.yourHand presence: yourHand is only
// populated for the connected player. We can't read selfPlayerId from
// GameState directly, so the page sets it via a module-level setter before
// applying actions. This keeps the reducer pure w.r.t. its (state, action)
// inputs while still knowing whose hand yourHand represents.
let selfPlayerIdForReducer: string | null = null;
export const setReducerSelfPlayerId = (playerId: string | null): void => {
  selfPlayerIdForReducer = playerId;
};
const selfId = (_state: GameState): string | null => selfPlayerIdForReducer;
const isSelfActor = (state: GameState, actorId: string): boolean =>
  actorId === state.currentPlayerId;

// DebtChip re-export so callers can build settlement flows without a second
// import path.
export type { DebtChip };
