type TurnControlsCardProps = {
  onPassTurn: () => void;
  movesLeft: number;
  showMovesLeft: boolean;
};

const TurnControlsCard = ({
  onPassTurn,
  movesLeft,
  showMovesLeft,
}: TurnControlsCardProps) => {
  return (
    <section className="game-sidebar-card game-turn-controls-card">
      <h2 className="game-sidebar-title">Turn</h2>
      <button
        type="button"
        className="ui-button ui-button--outline ui-button--sm game-pass-turn-button"
        onClick={onPassTurn}
      >
        {showMovesLeft ? `Pass Turn (${movesLeft} left)` : "Pass Turn"}
      </button>
    </section>
  );
};

export default TurnControlsCard;
