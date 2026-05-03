import Button from "../../../components/ui/button";
import type { WonGameResult } from "./monopolyDealPageTypes";

type MonopolyDealGameEndOverlayProps = {
  wonGame: WonGameResult;
  winnerName: string;
  winnerAvatarUrl: string;
  winnerInitial: string;
  onGoHome: () => void;
};

const MonopolyDealGameEndOverlay = ({
  wonGame,
  winnerName,
  winnerAvatarUrl,
  winnerInitial,
  onGoHome,
}: MonopolyDealGameEndOverlayProps) => {
  return (
    <section
      className="game-end-overlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="game-end-title"
    >
      <article className="game-end-modal">
        <p className="game-end-modal__eyebrow">Game over</p>
        <h2 id="game-end-title" className="game-end-modal__title">
          {winnerName} won the game
        </h2>
        <div className="game-end-modal__winner">
          {winnerAvatarUrl ? (
            <img
              className="game-end-modal__avatar"
              src={winnerAvatarUrl}
              alt={winnerName}
              loading="lazy"
              referrerPolicy="no-referrer"
            />
          ) : (
            <div className="game-end-modal__avatar game-end-modal__avatar--fallback" aria-hidden="true">
              {winnerInitial}
            </div>
          )}
          <p className="game-end-modal__winner-name">{winnerName}</p>
        </div>
        <dl className="game-end-modal__stats" aria-label="Winning stats">
          <div className="game-end-modal__stat">
            <dt>Completed sets</dt>
            <dd>{wonGame.numCompletedSets}</dd>
          </div>
          <div className="game-end-modal__stat">
            <dt>Money</dt>
            <dd>${wonGame.money}</dd>
          </div>
        </dl>
        <Button className="game-end-modal__button" onClick={onGoHome}>
          Go Home
        </Button>
      </article>
    </section>
  );
};

export default MonopolyDealGameEndOverlay;
