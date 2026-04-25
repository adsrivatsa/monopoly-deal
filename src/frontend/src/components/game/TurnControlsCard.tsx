type TurnControlsCardProps = {
  onPassTurn: () => void;
  onSubmitDiscard: () => void;
  movesLeft: number;
  showMovesLeft: boolean;
  isDiscardRequired: boolean;
  selectedDiscardCount: number;
  requiredDiscardCount: number;
  className?: string;
};

const TurnControlsCard = ({
  onPassTurn,
  onSubmitDiscard,
  movesLeft,
  showMovesLeft,
  isDiscardRequired,
  selectedDiscardCount,
  requiredDiscardCount,
  className,
}: TurnControlsCardProps) => {
  const isSubmitEnabled = selectedDiscardCount === requiredDiscardCount;
  const cardClassName = ["game-sidebar-card", "game-turn-controls-card", className]
    .filter(Boolean)
    .join(" ");

  return (
    <section className={cardClassName}>
      <h2 className="game-sidebar-title">{isDiscardRequired ? "Discard" : "Turn"}</h2>
      <button
        type="button"
        className="ui-button ui-button--outline ui-button--sm game-pass-turn-button"
        onClick={isDiscardRequired ? onSubmitDiscard : onPassTurn}
        disabled={isDiscardRequired ? !isSubmitEnabled : false}
      >
        {isDiscardRequired
          ? `Submit Discard (${selectedDiscardCount}/${requiredDiscardCount})`
          : showMovesLeft
            ? `Pass Turn (${movesLeft} left)`
            : "Pass Turn"}
      </button>
    </section>
  );
};

export default TurnControlsCard;
