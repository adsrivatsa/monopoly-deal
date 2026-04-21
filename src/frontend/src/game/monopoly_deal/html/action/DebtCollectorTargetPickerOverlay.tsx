import { type Player } from "../../../../generated/monopoly_deal";

type DebtCollectorTargetPickerOverlayProps = {
  players: Player[];
  selfPlayerId?: string;
  eyebrow: string;
  ariaLabel: string;
  onSelectPlayer: (playerId: string) => void;
  onClose: () => void;
};

const DebtCollectorTargetPickerOverlay = ({
  players,
  selfPlayerId,
  eyebrow,
  ariaLabel,
  onSelectPlayer,
  onClose,
}: DebtCollectorTargetPickerOverlayProps) => {
  const selectablePlayers = players.filter((player) => {
    if (!selfPlayerId) {
      return true;
    }

    return player.playerId !== selfPlayerId;
  });

  return (
    <div
      className="md-player-picker-backdrop"
      role="presentation"
      onClick={onClose}
    >
      <aside
        className="md-player-picker"
        role="dialog"
        aria-modal="true"
        aria-label={ariaLabel}
        onClick={(event) => event.stopPropagation()}
        onPointerDown={(event) => event.stopPropagation()}
      >
        <p className="md-player-picker__eyebrow">{eyebrow}</p>
        <p className="md-player-picker__title">Choose a player</p>

        {selectablePlayers.length === 0 ? (
          <p className="md-player-picker__empty">No valid player available.</p>
        ) : (
          <div className="md-player-picker__list">
            {selectablePlayers.map((player) => (
              <button
                key={player.playerId}
                type="button"
                className="md-player-picker__item"
                onClick={() => onSelectPlayer(player.playerId)}
              >
                <img
                  className="md-player-picker__avatar"
                  src={player.avatarUrl}
                  alt={player.displayName}
                  loading="lazy"
                  referrerPolicy="no-referrer"
                />
                <span className="md-player-picker__name">
                  {player.displayName}
                </span>
              </button>
            ))}
          </div>
        )}
      </aside>
    </div>
  );
};

export default DebtCollectorTargetPickerOverlay;
