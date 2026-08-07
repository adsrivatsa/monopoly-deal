// DealNoMercyGameMount renders the No Mercy board. This is a Phase 4a
// scaffold: it renders a minimal-but-real shell (opponents, your hand, debt
// chips) and leaves the full interactive board to Phase 4b. It mirrors the
// shape of MonopolyDealGameMount so Phase 4b can grow it into a real board
// component (board/DealNoMercyBoard.tsx) with per-play callbacks, exactly as
// the classic game did.
//
// Phase 4b: replace this file's body with a <DealNoMercyBoard /> and thread
// the play/comply/deny callbacks (see api/dealNoMercySocket.ts for the full
// send-function inventory) down through props, the way the classic Mount does.
import type { Card, GameState } from "../../generated/deal_no_mercy";

export type DealNoMercyGameMountProps = {
  gameState: GameState | null;
  // assetKey -> large image URL, assembled from GameState.assetImages.
  assetImageByKey: Record<number, string>;
  selfPlayerId: string | null;
  playerNameById: Record<string, string>;
};

const cardLabel = (card: Card): string => {
  // Phase 4b: render a proper card face (color pips, value, category badge).
  // For the scaffold we fall back to the asset key so unmapped cards are still
  // debuggable.
  return `#${card.assetKey}`;
};

const DealNoMercyGameMount = ({
  gameState,
  assetImageByKey,
  selfPlayerId,
  playerNameById,
}: DealNoMercyGameMountProps) => {
  if (!gameState) {
    return <p className="deal-no-mercy-shell__loading">Loading game...</p>;
  }

  const handCards = gameState.yourHand?.cards ?? [];

  // Outstanding debt chips per player (each chip is one obligation).
  const debtCountByPlayer = gameState.debtChips.reduce<Record<string, number>>(
    (counts, chip) => {
      counts[chip.debtorId] = (counts[chip.debtorId] ?? 0) + 1;
      return counts;
    },
    {},
  );

  const nameFor = (playerId: string): string => {
    if (playerId === selfPlayerId) {
      return "You";
    }
    return playerNameById[playerId] ?? playerId.slice(0, 8);
  };

  return (
    <div className="deal-no-mercy-shell">
      {/* Phase 4b: the real board lives here — opponent bank/property columns,
          the play surface, demand overlays (Heist/Market Crash picks, Repo
          Man/Tax Day distribution, Property Raid color, Bank Swap, Pickpocket
          category, Go Again), the all-players rent flow, and the turn/discard
          controls. This scaffold only surfaces public state. */}
      <section className="deal-no-mercy-shell__players">
        <h2 className="deal-no-mercy-shell__heading">Players</h2>
        <ul className="deal-no-mercy-shell__player-list">
          {gameState.players.map((player) => {
            const isTurn = player.playerId === gameState.currentPlayerId;
            const debt = debtCountByPlayer[player.playerId] ?? 0;
            return (
              <li
                key={player.playerId}
                className="deal-no-mercy-shell__player"
                data-active={isTurn ? "true" : undefined}
              >
                <span className="deal-no-mercy-shell__player-name">
                  {nameFor(player.playerId)}
                  {isTurn ? " (turn)" : ""}
                </span>
                <span className="deal-no-mercy-shell__player-stats">
                  {player.completedSets} sets · ${player.money} · hand{" "}
                  {player.handCards}
                </span>
                {/* Debt chips: outstanding obligations owed by this player. */}
                {debt > 0 ? (
                  <span
                    className="deal-no-mercy-shell__debt"
                    title={`${debt} outstanding debt chip(s)`}
                  >
                    {Array.from({ length: debt }, (_, index) => (
                      <span
                        key={index}
                        className="deal-no-mercy-shell__debt-chip"
                        aria-hidden="true"
                      />
                    ))}
                    <span className="deal-no-mercy-shell__debt-count">
                      {debt} debt
                    </span>
                  </span>
                ) : null}
              </li>
            );
          })}
        </ul>
      </section>

      <section className="deal-no-mercy-shell__hand">
        <h2 className="deal-no-mercy-shell__heading">
          Your hand ({handCards.length})
        </h2>
        {/* Phase 4b: make cards interactive (click to play money/property,
            drag to sets, right-click for the action menu). Today they are
            display-only using the large asset image from GameState. */}
        <div className="deal-no-mercy-shell__cards">
          {handCards.map((card) => {
            const imageUrl = assetImageByKey[card.assetKey];
            return (
              <div key={card.cardId} className="deal-no-mercy-shell__card">
                {imageUrl ? (
                  <img
                    src={imageUrl}
                    alt={cardLabel(card)}
                    loading="lazy"
                    draggable={false}
                  />
                ) : (
                  <span className="deal-no-mercy-shell__card-fallback">
                    {cardLabel(card)}
                  </span>
                )}
              </div>
            );
          })}
          {handCards.length === 0 ? (
            <p className="deal-no-mercy-shell__empty">No cards in hand.</p>
          ) : null}
        </div>
      </section>
    </div>
  );
};

export default DealNoMercyGameMount;
